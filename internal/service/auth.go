package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	appauth "github.com/satriaardiperdana-2020/monelog-api/internal/auth"
	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var (
	ErrAuthentication = errors.New("authentication failed")
	ErrValidation     = errors.New("validation failed")
	ErrConflict       = errors.New("conflict")
	ErrNotFound       = errors.New("not found")
)

// User is the safe user representation returned by authentication services.
type User struct {
	ID       int64
	Email    string
	Role     string
	Timezone string
	Currency string
	IsDelete bool
	Version  int32
}

// Actor represents the currently authenticated database-backed identity.
type Actor struct {
	UserID int64
	Role   string
}

type RegisterInput struct {
	Email    string
	Password string
	Timezone string
}

type LoginInput struct {
	Email    string
	Password string
}

type Session struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken string
}

type Auth struct {
	pool      *pgxpool.Pool
	queries   *db.Queries
	jwt       *appauth.JWT
	now       func() time.Time
	dummyHash string
}

type categorySeed struct {
	Type string
	Name string
}

var defaultCategories = []categorySeed{
	{Type: "income", Name: "Gaji"},
	{Type: "income", Name: "Bonus"},
	{Type: "income", Name: "Investasi"},
	{Type: "income", Name: "Lainnya"},
	{Type: "expense", Name: "Makanan"},
	{Type: "expense", Name: "Transportasi"},
	{Type: "expense", Name: "Tagihan"},
	{Type: "expense", Name: "Belanja"},
	{Type: "expense", Name: "Kesehatan"},
	{Type: "expense", Name: "Hiburan"},
	{Type: "expense", Name: "Lainnya"},
}

func NewAuth(pool *pgxpool.Pool, signingKey []byte, issuer, audience string, accessTokenLifetime time.Duration) (*Auth, error) {
	jwtService, err := appauth.NewJWT(signingKey, issuer, audience, accessTokenLifetime)
	if err != nil {
		return nil, err
	}
	dummyHash, err := appauth.HashPassword("not-a-real-login-password")
	if err != nil {
		return nil, fmt.Errorf("create password verifier: %w", err)
	}
	return &Auth{pool: pool, queries: db.New(pool), jwt: jwtService, now: time.Now, dummyHash: dummyHash}, nil
}

func (a *Auth) Register(ctx context.Context, input RegisterInput) (User, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return User{}, ErrValidation
	}
	if err := appauth.ValidatePassword(input.Password); err != nil {
		return User{}, ErrValidation
	}
	timezone, err := ValidateTimezone(input.Timezone)
	if err != nil {
		return User{}, ErrValidation
	}
	passwordHash, err := appauth.HashPassword(input.Password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin registration: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	created, err := q.CreateUser(ctx, db.CreateUserParams{Email: email, PasswordHash: passwordHash, Timezone: timezone, Currency: "IDR"})
	if err != nil {
		return User{}, translateWriteError(err)
	}
	for _, seed := range defaultCategories {
		if _, err := q.CreateCategory(ctx, db.CreateCategoryParams{UserID: created.ID, Type: seed.Type, Name: seed.Name}); err != nil {
			return User{}, fmt.Errorf("seed default categories: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit registration: %w", err)
	}
	return safeUser(created), nil
}

func (a *Auth) Login(ctx context.Context, input LoginInput) (Session, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil || appauth.ValidatePassword(input.Password) != nil {
		return Session{}, ErrAuthentication
	}
	user, err := a.queries.GetActiveUserByEmail(ctx, email)
	if err != nil {
		_ = appauth.VerifyPassword(a.dummyHash, input.Password)
		return Session{}, ErrAuthentication
	}
	if !appauth.VerifyPassword(user.PasswordHash, input.Password) {
		return Session{}, ErrAuthentication
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin login: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	lockedUsers, err := q.LockUsersForUpdate(ctx, []int64{user.ID})
	if err != nil {
		return Session{}, fmt.Errorf("lock login user: %w", err)
	}
	if len(lockedUsers) != 1 || lockedUsers[0].IsDelete {
		return Session{}, ErrAuthentication
	}
	created, err := a.createSessionWithQueries(ctx, q, lockedUsers[0], nil)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit login: %w", err)
	}
	return created.Session, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	if refreshToken == "" {
		return Session{}, ErrAuthentication
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	session, err := q.GetRefreshSessionByTokenHashForUpdate(ctx, appauth.HashRefreshSecret(refreshToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrAuthentication
		}
		return Session{}, fmt.Errorf("load refresh session: %w", err)
	}
	lockedUsers, err := q.LockUsersForUpdate(ctx, []int64{session.UserID})
	if err != nil {
		return Session{}, fmt.Errorf("lock refresh user: %w", err)
	}
	if len(lockedUsers) != 1 || lockedUsers[0].IsDelete {
		_, _ = q.RevokeRefreshSessionFamily(ctx, db.RevokeRefreshSessionFamilyParams{UserID: session.UserID, FamilyID: session.FamilyID})
		if err := tx.Commit(ctx); err != nil {
			return Session{}, fmt.Errorf("commit invalid refresh: %w", err)
		}
		return Session{}, ErrAuthentication
	}
	if session.RevokedAt.Valid || !session.ExpiresAt.Valid || !session.ExpiresAt.Time.After(a.now()) {
		_, err := q.RevokeRefreshSessionFamily(ctx, db.RevokeRefreshSessionFamilyParams{UserID: session.UserID, FamilyID: session.FamilyID})
		if err != nil {
			return Session{}, fmt.Errorf("revoke refresh family: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return Session{}, fmt.Errorf("commit invalid refresh: %w", err)
		}
		return Session{}, ErrAuthentication
	}

	result, err := a.createSessionWithQueries(ctx, q, lockedUsers[0], &session.FamilyID)
	if err != nil {
		return Session{}, err
	}
	if _, err := q.RevokeRefreshSession(ctx, db.RevokeRefreshSessionParams{ID: session.ID, ReplacedBy: &result.sessionID}); err != nil {
		return Session{}, fmt.Errorf("revoke rotated session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit refresh: %w", err)
	}
	return result.Session, nil
}

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin logout: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	session, err := q.GetRefreshSessionByTokenHashForUpdate(ctx, appauth.HashRefreshSecret(refreshToken))
	if err == nil {
		if _, err := q.RevokeRefreshSessionFamily(ctx, db.RevokeRefreshSessionFamilyParams{UserID: session.UserID, FamilyID: session.FamilyID}); err != nil {
			return fmt.Errorf("revoke logout family: %w", err)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("load logout session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit logout: %w", err)
	}
	return nil
}

func (a *Auth) AuthenticateAccess(ctx context.Context, rawToken string) (Actor, error) {
	userID, err := a.jwt.Validate(rawToken)
	if err != nil {
		return Actor{}, ErrAuthentication
	}
	user, err := a.queries.GetActiveUserByID(ctx, userID)
	if err != nil {
		return Actor{}, ErrAuthentication
	}
	return Actor{UserID: userID, Role: user.Role}, nil
}

func (a *Auth) Me(ctx context.Context, actor Actor) (User, error) {
	user, err := a.queries.GetActiveUserByID(ctx, actor.UserID)
	if err != nil {
		return User{}, ErrAuthentication
	}
	return safeUser(user), nil
}

func (a *Auth) UpdateProfile(ctx context.Context, actor Actor, timezone string, expectedVersion int32) (User, error) {
	timezone, err := ValidateTimezone(timezone)
	if err != nil || expectedVersion <= 0 {
		return User{}, ErrValidation
	}
	current, err := a.queries.GetActiveUserByID(ctx, actor.UserID)
	if err != nil {
		return User{}, ErrAuthentication
	}
	updated, err := a.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{ID: current.ID, Email: current.Email, Timezone: timezone, Currency: "IDR", ExpectedVersion: expectedVersion})
	if err != nil {
		return User{}, translateWriteError(err)
	}
	return safeUser(updated), nil
}

func (a *Auth) DeleteMe(ctx context.Context, actor Actor, expectedVersion int32) error {
	if expectedVersion <= 0 {
		return ErrValidation
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin account deletion: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	users, err := q.LockUsersForUpdate(ctx, []int64{actor.UserID})
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	if len(users) != 1 || users[0].IsDelete {
		return ErrAuthentication
	}
	if _, err := q.SoftDeleteUser(ctx, db.SoftDeleteUserParams{ID: users[0].ID, ExpectedVersion: expectedVersion}); err != nil {
		return translateWriteError(err)
	}
	if _, err := q.RevokeAllUserRefreshSessions(ctx, users[0].ID); err != nil {
		return fmt.Errorf("revoke account sessions: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit account deletion: %w", err)
	}
	return nil
}

func (a *Auth) BootstrapInitialAdmin(ctx context.Context, input RegisterInput) (User, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil || appauth.ValidatePassword(input.Password) != nil {
		return User{}, ErrValidation
	}
	timezone, err := ValidateTimezone(input.Timezone)
	if err != nil {
		return User{}, ErrValidation
	}
	hash, err := appauth.HashPassword(input.Password)
	if err != nil {
		return User{}, fmt.Errorf("hash bootstrap password: %w", err)
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin bootstrap: %w", err)
	}
	defer tx.Rollback(ctx)
	q := a.queries.WithTx(tx)
	if err := q.LockInitialAdminBootstrap(ctx); err != nil {
		return User{}, fmt.Errorf("lock bootstrap: %w", err)
	}
	created, err := q.CreateInitialAdmin(ctx, db.CreateInitialAdminParams{Email: email, PasswordHash: hash, Timezone: timezone})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrConflict
		}
		return User{}, translateWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit bootstrap: %w", err)
	}
	return safeUser(created), nil
}

type createdSession struct {
	Session
	sessionID int64
}

func (a *Auth) createSessionWithQueries(ctx context.Context, q *db.Queries, user db.User, familyID *int64) (createdSession, error) {
	refreshToken, err := appauth.NewRefreshSecret()
	if err != nil {
		return createdSession{}, err
	}
	created, err := q.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{UserID: user.ID, FamilyID: familyID, TokenHash: appauth.HashRefreshSecret(refreshToken), ExpiresAt: pgtype.Timestamptz{Time: a.now().Add(30 * 24 * time.Hour), Valid: true}})
	if err != nil {
		return createdSession{}, fmt.Errorf("create refresh session: %w", err)
	}
	accessToken, expiresAt, err := a.jwt.Issue(user.ID)
	if err != nil {
		return createdSession{}, err
	}
	return createdSession{Session: Session{AccessToken: accessToken, ExpiresAt: expiresAt, RefreshToken: refreshToken}, sessionID: created.ID}, nil
}

func ValidateTimezone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 64 || value == "Local" || !strings.Contains(value, "/") && value != "UTC" {
		return "", errors.New("invalid timezone")
	}
	location, err := time.LoadLocation(value)
	if err != nil || location.String() == "Local" {
		return "", errors.New("invalid timezone")
	}
	return location.String(), nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 0 || len(value) > 254 {
		return "", errors.New("invalid email")
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || !strings.Contains(value, "@") {
		return "", errors.New("invalid email")
	}
	return value, nil
}

func safeUser(user db.User) User {
	return User{ID: user.ID, Email: user.Email, Role: user.Role, Timezone: user.Timezone, Currency: user.Currency, IsDelete: user.IsDelete, Version: user.Version}
}

func translateWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	return fmt.Errorf("database write: %w", err)
}
