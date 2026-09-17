package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satriaardiperdana-2020/monelog-api/internal/repository"
	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

const (
	CategoryTypeIncome  = "income"
	CategoryTypeExpense = "expense"
)

var ErrForbidden = errors.New("forbidden")

// CategoryScope carries server-derived actor and owner identities.
type CategoryScope struct {
	Actor Actor
	Owner uuid.UUID
	Admin bool
}

type Category struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Type     string
	Name     string
	IsDelete bool
	Version  int32
}

// Categories contains category lifecycle and authorization rules.
type Categories struct {
	repository *repository.Categories
}

func NewCategories(pool *pgxpool.Pool) *Categories {
	return &Categories{repository: repository.NewCategories(pool)}
}

func (s *Categories) List(ctx context.Context, scope CategoryScope, categoryType string, deleted bool) ([]Category, error) {
	if err := s.validateScope(ctx, scope); err != nil {
		return nil, err
	}
	if categoryType != "" && !validCategoryType(categoryType) {
		return nil, ErrValidation
	}
	items, err := s.repository.List(ctx, scope.Owner, deleted)
	if err != nil {
		return nil, classifyCategoryError(err)
	}
	result := make([]Category, 0, len(items))
	for _, item := range items {
		if categoryType == "" || item.Type == categoryType {
			result = append(result, categoryFromDB(item))
		}
	}
	return result, nil
}

func (s *Categories) Get(ctx context.Context, scope CategoryScope, id uuid.UUID, deleted bool) (Category, error) {
	if err := s.validateScope(ctx, scope); err != nil {
		return Category{}, err
	}
	item, err := s.repository.Get(ctx, id, scope.Owner, deleted)
	if err != nil {
		return Category{}, classifyCategoryError(err)
	}
	return categoryFromDB(item), nil
}

func (s *Categories) Create(ctx context.Context, scope CategoryScope, categoryType, name string) (Category, error) {
	if err := s.validateScope(ctx, scope); err != nil {
		return Category{}, err
	}
	name, err := validateCategoryName(name)
	if err != nil || !validCategoryType(categoryType) {
		return Category{}, ErrValidation
	}
	id := uuid.New()
	var item db.Category
	if scope.Admin {
		item, err = s.repository.CreateAdmin(ctx, id, scope.Actor.UserID, scope.Owner, categoryType, name)
	} else {
		item, err = s.repository.Create(ctx, id, scope.Owner, categoryType, name)
	}
	if err != nil {
		return Category{}, classifyCategoryError(err)
	}
	return categoryFromDB(item), nil
}

func (s *Categories) Update(ctx context.Context, scope CategoryScope, id uuid.UUID, name string, version int32) (Category, error) {
	if err := s.validateScope(ctx, scope); err != nil {
		return Category{}, err
	}
	name, err := validateCategoryName(name)
	if err != nil || version <= 0 {
		return Category{}, ErrValidation
	}
	if _, err := s.repository.Get(ctx, id, scope.Owner, false); err != nil {
		return Category{}, classifyCategoryError(err)
	}
	var item db.Category
	if scope.Admin {
		item, err = s.repository.UpdateAdmin(ctx, id, scope.Actor.UserID, scope.Owner, name, version)
	} else {
		item, err = s.repository.Update(ctx, id, scope.Owner, name, version)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrConflict
		}
		return Category{}, classifyCategoryError(err)
	}
	return categoryFromDB(item), nil
}

func (s *Categories) Archive(ctx context.Context, scope CategoryScope, id uuid.UUID, version int32) error {
	if err := s.validateScope(ctx, scope); err != nil {
		return err
	}
	if version <= 0 {
		return ErrValidation
	}
	if _, err := s.repository.Get(ctx, id, scope.Owner, false); err != nil {
		return classifyCategoryError(err)
	}
	if scope.Admin {
		_, err := s.repository.ArchiveAdmin(ctx, id, scope.Actor.UserID, scope.Owner, version)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		return classifyCategoryError(err)
	}
	_, err := s.repository.Archive(ctx, id, scope.Owner, version)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrConflict
	}
	return classifyCategoryError(err)
}

func (s *Categories) Restore(ctx context.Context, scope CategoryScope, id uuid.UUID, version int32) (Category, error) {
	if err := s.validateScope(ctx, scope); err != nil {
		return Category{}, err
	}
	if version <= 0 {
		return Category{}, ErrValidation
	}
	if _, err := s.repository.Get(ctx, id, scope.Owner, true); err != nil {
		return Category{}, classifyCategoryError(err)
	}
	var item db.Category
	var err error
	if scope.Admin {
		item, err = s.repository.RestoreAdmin(ctx, id, scope.Actor.UserID, scope.Owner, version)
	} else {
		item, err = s.repository.Restore(ctx, id, scope.Owner, version)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrConflict
		}
		return Category{}, classifyCategoryError(err)
	}
	return categoryFromDB(item), nil
}

func (s *Categories) validateScope(ctx context.Context, scope CategoryScope) error {
	if scope.Actor.UserID == uuid.Nil || scope.Owner == uuid.Nil {
		return ErrValidation
	}
	if scope.Admin {
		if scope.Actor.Role != RoleAdmin {
			return ErrForbidden
		}
		if err := s.repository.OwnerIsActive(ctx, scope.Owner); err != nil {
			return classifyCategoryError(err)
		}
		return nil
	}
	if scope.Actor.UserID != scope.Owner {
		return ErrForbidden
	}
	return nil
}

func validCategoryType(value string) bool {
	return value == CategoryTypeIncome || value == CategoryTypeExpense
}

func validateCategoryName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 80 {
		return "", ErrValidation
	}
	return value, nil
}

func categoryFromDB(item db.Category) Category {
	return Category{ID: uuid.UUID(item.ID.Bytes), UserID: uuid.UUID(item.UserID.Bytes), Type: item.Type, Name: item.Name, IsDelete: item.IsDelete, Version: item.Version}
}

func classifyCategoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" {
		return ErrConflict
	}
	return err
}
