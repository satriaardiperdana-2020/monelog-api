package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type fakeTemplateService struct {
	createCalls int
	scope       service.TemplateScope
}

func (f *fakeTemplateService) List(context.Context, service.TemplateScope, bool) ([]service.TransactionTemplate, error) {
	return nil, nil
}
func (f *fakeTemplateService) Get(context.Context, service.TemplateScope, int64, bool) (service.TransactionTemplate, error) {
	return service.TransactionTemplate{}, service.ErrNotFound
}
func (f *fakeTemplateService) Create(_ context.Context, scope service.TemplateScope, input service.TemplateInput) (service.TransactionTemplate, error) {
	f.createCalls++
	f.scope = scope
	return service.TransactionTemplate{ID: 3, UserID: scope.Owner, CategoryID: input.CategoryID, Type: input.Type, Name: input.Name, Amount: input.Amount, Title: input.Title, Version: 1}, nil
}
func (f *fakeTemplateService) Update(context.Context, service.TemplateScope, int64, service.TemplateInput) (service.TransactionTemplate, error) {
	return service.TransactionTemplate{}, nil
}
func (f *fakeTemplateService) SoftDelete(context.Context, service.TemplateScope, int64, int32) error {
	return nil
}
func (f *fakeTemplateService) Restore(context.Context, service.TemplateScope, int64, int32) (service.TransactionTemplate, error) {
	return service.TransactionTemplate{}, nil
}
func (f *fakeTemplateService) Apply(_ context.Context, scope service.TemplateScope, _ int64) (service.TemplateDraft, error) {
	f.scope = scope
	return service.TemplateDraft{CategoryID: 2, Type: service.CategoryTypeExpense, Amount: "10.00", Title: "Food"}, nil
}

func TestTemplateCreateUsesAuthenticatedOwnerAndRejectsInjection(t *testing.T) {
	var actorID int64 = 1
	fake := &fakeTemplateService{}
	handler := NewTemplates(fake)
	e := echo.New()
	e.POST("/templates", handler.Create, middleware.Authenticate(categoryAuthenticator{actor: service.Actor{UserID: actorID, Role: service.RoleUser}}))
	request := httptest.NewRequest(http.MethodPost, "/templates", strings.NewReader(`{"name":"Lunch","category_id":2,"type":"expense","amount":"10.00","title":"Food","user_id":4}`))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || fake.createCalls != 0 {
		t.Fatalf("injected owner status=%d calls=%d", response.Code, fake.createCalls)
	}

	request = httptest.NewRequest(http.MethodPost, "/templates", strings.NewReader(`{"name":"Lunch","category_id":2,"type":"expense","amount":"10.00","title":"Food"}`))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response = httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || fake.scope.Owner != actorID || fake.scope.Admin {
		t.Fatalf("status=%d scope=%+v", response.Code, fake.scope)
	}
}

func TestTemplateApplyReturnsDraftOnly(t *testing.T) {
	fake := &fakeTemplateService{}
	handler := NewTemplates(fake)
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/templates/3/apply", nil)
	response := httptest.NewRecorder()
	if err := handler.Apply(e.NewContext(request, response), 3); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"category_id":2`) || strings.Contains(response.Body.String(), "transaction_date") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

var _ api.TransactionTemplate = api.TransactionTemplate{}
