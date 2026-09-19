package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satriaardiperdana-2020/monelog-api/internal/repository"
	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

// TemplateScope is an owner-scoped request. It has the same actor/owner rules
// as transaction operations, while templates themselves never create a transaction.
type TemplateScope = TransactionScope

type TransactionTemplate struct {
	ID, UserID, CategoryID    int64
	Type, Name, Amount, Title string
	IsDelete                  bool
	Version                   int32
}

type TemplateInput struct {
	CategoryID int64
	Type       string
	Name       string
	Amount     string
	Title      string
	Version    int32
}

type TemplateDraft struct {
	CategoryID int64
	Type       string
	Amount     string
	Title      string
}

// Templates owns reusable transaction input. It never inserts into transactions.
type Templates struct {
	pool       *pgxpool.Pool
	repository *repository.Templates
}

func NewTemplates(pool *pgxpool.Pool) *Templates {
	return &Templates{pool: pool, repository: repository.NewTemplates(pool)}
}

func (s *Templates) List(ctx context.Context, scope TemplateScope, deleted bool) ([]TransactionTemplate, error) {
	if err := s.validateReadScope(ctx, scope); err != nil {
		return nil, err
	}
	items, err := s.repository.List(ctx, scope.Owner, deleted)
	if err != nil {
		return nil, classifyTemplateError(err)
	}
	result := make([]TransactionTemplate, 0, len(items))
	for _, item := range items {
		mapped, err := templateFromDB(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "list", nil); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Templates) Get(ctx context.Context, scope TemplateScope, id int64, deleted bool) (TransactionTemplate, error) {
	if id <= 0 {
		return TransactionTemplate{}, ErrValidation
	}
	if err := s.validateReadScope(ctx, scope); err != nil {
		return TransactionTemplate{}, err
	}
	item, err := s.repository.Get(ctx, id, scope.Owner, deleted)
	if err != nil {
		return TransactionTemplate{}, classifyTemplateError(err)
	}
	mapped, err := templateFromDB(item)
	if err != nil {
		return TransactionTemplate{}, err
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "read", &id); err != nil {
			return TransactionTemplate{}, err
		}
	}
	return mapped, nil
}

func (s *Templates) Create(ctx context.Context, scope TemplateScope, input TemplateInput) (TransactionTemplate, error) {
	validated, err := validateTemplateInput(scope.Owner, input, false)
	if err != nil {
		return TransactionTemplate{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TransactionTemplate{}, err
	}
	defer tx.Rollback(ctx)
	q := s.repository.Queries().WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return TransactionTemplate{}, err
	}
	if err := validateLockedTemplateCategory(ctx, q, scope.Owner, validated.categoryID, validated.kind); err != nil {
		return TransactionTemplate{}, err
	}
	created, err := q.CreateTransactionTemplate(ctx, db.CreateTransactionTemplateParams{UserID: scope.Owner, CategoryID: validated.categoryID, Type: validated.kind, Name: validated.name, Amount: validated.amount.Numeric(), Title: validated.title})
	if err != nil {
		return TransactionTemplate{}, classifyTemplateError(err)
	}
	if scope.Admin {
		if err := insertTemplateAudit(ctx, q, scope, &created.ID, "create"); err != nil {
			return TransactionTemplate{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return TransactionTemplate{}, err
	}
	return templateFromDB(created)
}

func (s *Templates) Update(ctx context.Context, scope TemplateScope, id int64, input TemplateInput) (TransactionTemplate, error) {
	validated, err := validateTemplateInput(scope.Owner, input, true)
	if err != nil || id <= 0 {
		return TransactionTemplate{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TransactionTemplate{}, err
	}
	defer tx.Rollback(ctx)
	q := s.repository.Queries().WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return TransactionTemplate{}, err
	}
	state, err := q.GetTransactionTemplateStateForUpdate(ctx, db.GetTransactionTemplateStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return TransactionTemplate{}, ErrNotFound
	}
	if err != nil {
		return TransactionTemplate{}, err
	}
	if state.IsDelete || state.Version != input.Version {
		return TransactionTemplate{}, ErrConflict
	}
	if err := validateLockedTemplateCategory(ctx, q, scope.Owner, validated.categoryID, validated.kind); err != nil {
		return TransactionTemplate{}, err
	}
	updated, err := q.UpdateTransactionTemplate(ctx, db.UpdateTransactionTemplateParams{ID: id, UserID: scope.Owner, CategoryID: validated.categoryID, Type: validated.kind, Name: validated.name, Amount: validated.amount.Numeric(), Title: validated.title, ExpectedVersion: input.Version})
	if errors.Is(err, pgx.ErrNoRows) {
		return TransactionTemplate{}, ErrConflict
	}
	if err != nil {
		return TransactionTemplate{}, classifyTemplateError(err)
	}
	if scope.Admin {
		if err := insertTemplateAudit(ctx, q, scope, &updated.ID, "update"); err != nil {
			return TransactionTemplate{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return TransactionTemplate{}, err
	}
	return templateFromDB(updated)
}

func (s *Templates) SoftDelete(ctx context.Context, scope TemplateScope, id int64, version int32) error {
	if id <= 0 || version <= 0 {
		return ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.repository.Queries().WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return err
	}
	state, err := q.GetTransactionTemplateStateForUpdate(ctx, db.GetTransactionTemplateStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if state.IsDelete || state.Version != version {
		return ErrConflict
	}
	deleted, err := q.SoftDeleteTransactionTemplate(ctx, db.SoftDeleteTransactionTemplateParams{ID: id, UserID: scope.Owner, ExpectedVersion: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return classifyTemplateError(err)
	}
	if scope.Admin {
		if err := insertTemplateAudit(ctx, q, scope, &deleted.ID, "delete"); err != nil {
			return ErrUnavailable
		}
	}
	return tx.Commit(ctx)
}

func (s *Templates) Restore(ctx context.Context, scope TemplateScope, id int64, version int32) (TransactionTemplate, error) {
	if id <= 0 || version <= 0 {
		return TransactionTemplate{}, ErrValidation
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TransactionTemplate{}, err
	}
	defer tx.Rollback(ctx)
	q := s.repository.Queries().WithTx(tx)
	if _, err := validateWriteUsers(ctx, q, scope); err != nil {
		return TransactionTemplate{}, err
	}
	state, err := q.GetTransactionTemplateStateForUpdate(ctx, db.GetTransactionTemplateStateForUpdateParams{ID: id, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return TransactionTemplate{}, ErrNotFound
	}
	if err != nil {
		return TransactionTemplate{}, err
	}
	if !state.IsDelete || state.Version != version {
		return TransactionTemplate{}, ErrConflict
	}
	if err := validateLockedTemplateCategory(ctx, q, scope.Owner, state.CategoryID, state.Type); err != nil {
		return TransactionTemplate{}, err
	}
	restored, err := q.RestoreTransactionTemplate(ctx, db.RestoreTransactionTemplateParams{ID: id, UserID: scope.Owner, ExpectedVersion: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return TransactionTemplate{}, ErrConflict
	}
	if err != nil {
		return TransactionTemplate{}, classifyTemplateError(err)
	}
	if scope.Admin {
		if err := insertTemplateAudit(ctx, q, scope, &restored.ID, "restore"); err != nil {
			return TransactionTemplate{}, ErrUnavailable
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return TransactionTemplate{}, err
	}
	return templateFromDB(restored)
}

func (s *Templates) Apply(ctx context.Context, scope TemplateScope, id int64) (TemplateDraft, error) {
	if id <= 0 {
		return TemplateDraft{}, ErrValidation
	}
	if err := s.validateReadScope(ctx, scope); err != nil {
		return TemplateDraft{}, err
	}
	stored, err := s.repository.Get(ctx, id, scope.Owner, false)
	if err != nil {
		return TemplateDraft{}, classifyTemplateError(err)
	}
	item, err := templateFromDB(stored)
	if err != nil {
		return TemplateDraft{}, err
	}
	category, err := s.repository.Queries().LockScopedCategoryForTemplate(ctx, db.LockScopedCategoryForTemplateParams{CategoryID: item.CategoryID, UserID: scope.Owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return TemplateDraft{}, ErrNotFound
	}
	if err != nil {
		return TemplateDraft{}, err
	}
	if category.IsDelete {
		return TemplateDraft{}, ErrConflict
	}
	if category.Type != item.Type {
		return TemplateDraft{}, ErrValidation
	}
	if scope.Admin {
		if err := s.auditRead(ctx, scope, "apply", &id); err != nil {
			return TemplateDraft{}, err
		}
	}
	return TemplateDraft{CategoryID: item.CategoryID, Type: item.Type, Amount: item.Amount, Title: item.Title}, nil
}

func (s *Templates) validateReadScope(ctx context.Context, scope TemplateScope) error {
	if scope.Actor.UserID <= 0 || scope.Owner <= 0 {
		return ErrValidation
	}
	if !scope.Admin {
		if scope.Actor.UserID != scope.Owner {
			return ErrForbidden
		}
		return nil
	}
	if scope.Actor.Role != RoleAdmin {
		return ErrForbidden
	}
	if _, err := s.repository.Queries().GetUserByID(ctx, scope.Owner); errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else {
		return err
	}
}

type validatedTemplate struct {
	categoryID int64
	kind       string
	name       string
	amount     Money
	title      string
}

func validateTemplateInput(owner int64, input TemplateInput, update bool) (validatedTemplate, error) {
	name, nameErr := validateCategoryName(input.Name)
	amount, amountErr := ParseMoney(input.Amount)
	title, titleErr := normalizeTitle(input.Title)
	if owner <= 0 || input.CategoryID <= 0 || !validCategoryType(input.Type) || nameErr != nil || amountErr != nil || titleErr != nil || (update && input.Version <= 0) {
		return validatedTemplate{}, ErrValidation
	}
	return validatedTemplate{categoryID: input.CategoryID, kind: input.Type, name: name, amount: amount, title: title}, nil
}

func validateLockedTemplateCategory(ctx context.Context, q *db.Queries, owner, categoryID int64, kind string) error {
	category, err := q.LockScopedCategoryForTemplate(ctx, db.LockScopedCategoryForTemplateParams{CategoryID: categoryID, UserID: owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if category.IsDelete {
		return ErrConflict
	}
	if category.Type != kind {
		return ErrValidation
	}
	return nil
}

func templateFromDB(item db.TransactionTemplate) (TransactionTemplate, error) {
	amount, err := moneyFromNumeric(item.Amount)
	if err != nil {
		return TransactionTemplate{}, err
	}
	return TransactionTemplate{ID: item.ID, UserID: item.UserID, CategoryID: item.CategoryID, Type: item.Type, Name: item.Name, Amount: amount.String(), Title: item.Title, IsDelete: item.IsDelete, Version: item.Version}, nil
}

func classifyTemplateError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func insertTemplateAudit(ctx context.Context, q *db.Queries, scope TemplateScope, resourceID *int64, action string) error {
	requestID := scope.RequestID
	if requestID == "" {
		requestID = uuid.NewString()
	}
	_, err := q.InsertAdminAccessEvent(ctx, db.InsertAdminAccessEventParams{ActorUserID: scope.Actor.UserID, TargetUserID: &scope.Owner, ResourceType: "template", ResourceID: resourceID, Action: action, Outcome: "success", RequestID: requestID, SafeMetadata: []byte(`{}`)})
	return err
}

func (s *Templates) auditRead(ctx context.Context, scope TemplateScope, action string, resourceID *int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ErrUnavailable
	}
	defer tx.Rollback(ctx)
	q := s.repository.Queries().WithTx(tx)
	actor, err := q.GetActiveUserByID(ctx, scope.Actor.UserID)
	if err != nil || actor.Role != RoleAdmin {
		return ErrForbidden
	}
	if _, err := q.GetUserByID(ctx, scope.Owner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return ErrUnavailable
	}
	if err := insertTemplateAudit(ctx, q, scope, resourceID, action); err != nil {
		return ErrUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrUnavailable
	}
	return nil
}
