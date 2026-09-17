package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type fakeCategoryService struct {
	createScope service.CategoryScope
	createCalls int
	listDeleted bool
}

func (f *fakeCategoryService) List(_ context.Context, _ service.CategoryScope, _ string, deleted bool) ([]service.Category, error) {
	f.listDeleted = deleted
	return nil, nil
}
func (f *fakeCategoryService) Get(context.Context, service.CategoryScope, uuid.UUID, bool) (service.Category, error) {
	return service.Category{}, service.ErrNotFound
}
func (f *fakeCategoryService) Create(_ context.Context, scope service.CategoryScope, categoryType, name string) (service.Category, error) {
	f.createCalls++
	f.createScope = scope
	return service.Category{ID: uuid.New(), UserID: scope.Owner, Type: categoryType, Name: name, Version: 1}, nil
}
func (f *fakeCategoryService) Update(context.Context, service.CategoryScope, uuid.UUID, string, int32) (service.Category, error) {
	return service.Category{}, nil
}
func (f *fakeCategoryService) Archive(context.Context, service.CategoryScope, uuid.UUID, int32) error {
	return nil
}
func (f *fakeCategoryService) Restore(context.Context, service.CategoryScope, uuid.UUID, int32) (service.Category, error) {
	return service.Category{}, nil
}

type categoryAuthenticator struct{ actor service.Actor }

func (a categoryAuthenticator) AuthenticateAccess(context.Context, string) (service.Actor, error) {
	return a.actor, nil
}

func TestCategoryCreateRejectsOwnerInjection(t *testing.T) {
	fake := &fakeCategoryService{}
	handler := NewCategories(fake)
	e := echo.New()
	e.POST("/categories", handler.Create, middleware.Authenticate(categoryAuthenticator{actor: service.Actor{UserID: uuid.New(), Role: service.RoleUser}}))
	request := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Food","type":"expense","owner_id":"ignored"}`))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || fake.createCalls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, fake.createCalls)
	}
}

func TestCategoryCreateUsesAuthenticatedOwner(t *testing.T) {
	actorID := uuid.New()
	fake := &fakeCategoryService{}
	handler := NewCategories(fake)
	e := echo.New()
	e.POST("/categories", handler.Create, middleware.Authenticate(categoryAuthenticator{actor: service.Actor{UserID: actorID, Role: service.RoleUser}}))
	request := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Food","type":"expense"}`))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || fake.createScope.Owner != actorID || fake.createScope.Admin {
		t.Fatalf("status=%d scope=%+v", response.Code, fake.createScope)
	}
}

func TestCategoryUpdateRejectsTypeChange(t *testing.T) {
	handler := NewCategories(&fakeCategoryService{})
	e := echo.New()
	request := httptest.NewRequest(http.MethodPatch, "/categories/id", strings.NewReader(`{"name":"Food","type":"income","version":1}`))
	response := httptest.NewRecorder()
	if err := handler.Update(e.NewContext(request, response), uuid.New()); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestCategoryListDefaultsToActive(t *testing.T) {
	fake := &fakeCategoryService{}
	handler := NewCategories(fake)
	e := echo.New()
	e.GET("/categories", func(c *echo.Context) error { return handler.List(c, api.ListCategoriesParams{}) }, middleware.Authenticate(categoryAuthenticator{actor: service.Actor{UserID: uuid.New(), Role: service.RoleUser}}))
	request := httptest.NewRequest(http.MethodGet, "/categories", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusOK || fake.listDeleted {
		t.Fatalf("status=%d deleted=%t", response.Code, fake.listDeleted)
	}
}
