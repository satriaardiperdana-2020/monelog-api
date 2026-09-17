package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type fakeTransactionService struct {
	createCalls int
	scope       service.TransactionScope
}

func (f *fakeTransactionService) List(context.Context, service.TransactionScope, service.TransactionFilter) (service.TransactionPage, error) {
	return service.TransactionPage{}, nil
}
func (f *fakeTransactionService) Get(context.Context, service.TransactionScope, uuid.UUID, bool) (service.Transaction, error) {
	return service.Transaction{}, nil
}
func (f *fakeTransactionService) Create(_ context.Context, scope service.TransactionScope, input service.TransactionInput) (service.CreateTransactionResult, error) {
	f.createCalls++
	f.scope = scope
	return service.CreateTransactionResult{Created: true, Transaction: service.Transaction{ID: uuid.New(), UserID: scope.Owner, CategoryID: input.CategoryID, ClientRequestID: input.ClientRequestID, Type: input.Type, Amount: input.Amount, Title: input.Title, Version: 1}}, nil
}
func (f *fakeTransactionService) Update(context.Context, service.TransactionScope, uuid.UUID, service.TransactionInput) (service.Transaction, error) {
	return service.Transaction{}, nil
}
func (f *fakeTransactionService) SoftDeleteTransaction(context.Context, service.TransactionScope, uuid.UUID, int32) error {
	return nil
}
func (f *fakeTransactionService) RestoreTransaction(context.Context, service.TransactionScope, uuid.UUID, int32) (service.Transaction, error) {
	return service.Transaction{}, nil
}
func (f *fakeTransactionService) ListDailySummaries(context.Context, service.TransactionScope, string, string, int32, string) (service.DailySummaryPage, error) {
	return service.DailySummaryPage{}, nil
}
func (f *fakeTransactionService) GetReportSummary(context.Context, service.TransactionScope, service.ReportFilter) (service.ReportSummary, error) {
	return service.ReportSummary{}, nil
}
func (f *fakeTransactionService) GetReportBreakdown(context.Context, service.TransactionScope, service.ReportFilter) (service.ReportBreakdown, error) {
	return service.ReportBreakdown{}, nil
}

func TestTransactionCreateUsesActorOwnerAndRejectsOwnerInjection(t *testing.T) {
	actorID := uuid.New()
	fake := &fakeTransactionService{}
	handler := NewTransactions(fake)
	e := echo.New()
	e.POST("/transactions", handler.Create, middleware.Authenticate(categoryAuthenticator{actor: service.Actor{UserID: actorID, Role: service.RoleUser}}))
	body := `{"transaction_date":"2026-09-17","type":"expense","category_id":"` + uuid.NewString() + `","amount":"10.00","title":"Food","client_request_id":"` + uuid.NewString() + `","user_id":"` + uuid.NewString() + `"}`
	request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || fake.createCalls != 0 {
		t.Fatalf("injected owner status=%d calls=%d", response.Code, fake.createCalls)
	}

	body = `{"transaction_date":"2026-09-17","type":"expense","category_id":"` + uuid.NewString() + `","amount":"10.00","title":"Food","client_request_id":"` + uuid.NewString() + `"}`
	request = httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	response = httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || fake.scope.Owner != actorID || fake.scope.Admin {
		t.Fatalf("status=%d scope=%+v", response.Code, fake.scope)
	}
}
