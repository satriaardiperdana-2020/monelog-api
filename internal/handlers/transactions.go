package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type transactionService interface {
	List(context.Context, service.TransactionScope, service.TransactionFilter) (service.TransactionPage, error)
	Get(context.Context, service.TransactionScope, int64, bool) (service.Transaction, error)
	Create(context.Context, service.TransactionScope, service.TransactionInput) (service.CreateTransactionResult, error)
	Update(context.Context, service.TransactionScope, int64, service.TransactionInput) (service.Transaction, error)
	SoftDeleteTransaction(context.Context, service.TransactionScope, int64, int32) error
	RestoreTransaction(context.Context, service.TransactionScope, int64, int32) (service.Transaction, error)
	ListDailySummaries(context.Context, service.TransactionScope, string, string, int32, string) (service.DailySummaryPage, error)
	GetReportSummary(context.Context, service.TransactionScope, service.ReportFilter) (service.ReportSummary, error)
	GetReportBreakdown(context.Context, service.TransactionScope, service.ReportFilter) (service.ReportBreakdown, error)
}

func (h *Transactions) Summary(c *echo.Context, p api.GetReportSummaryParams) error {
	return h.summary(c, h.personalScope(c), service.ReportFilter{Range: string(p.Range), StartDate: dateParam(p.StartDate), EndDate: dateParam(p.EndDate)})
}
func (h *Transactions) AdminSummary(c *echo.Context, owner int64, p api.AdminGetReportSummaryParams) error {
	return h.summary(c, h.adminScope(c, owner), service.ReportFilter{Range: string(p.Range), StartDate: dateParam(p.StartDate), EndDate: dateParam(p.EndDate)})
}
func (h *Transactions) summary(c *echo.Context, scope service.TransactionScope, filter service.ReportFilter) error {
	item, err := h.service.GetReportSummary(c.Request().Context(), scope, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.ReportSummaryResponse{Data: reportSummaryPayload(item)})
}
func (h *Transactions) Breakdown(c *echo.Context, p api.GetReportBreakdownParams) error {
	return h.breakdown(c, h.personalScope(c), service.ReportFilter{Range: string(p.Range), StartDate: dateParam(p.StartDate), EndDate: dateParam(p.EndDate), GroupBy: string(p.GroupBy)})
}
func (h *Transactions) AdminBreakdown(c *echo.Context, owner int64, p api.AdminGetReportBreakdownParams) error {
	return h.breakdown(c, h.adminScope(c, owner), service.ReportFilter{Range: string(p.Range), StartDate: dateParam(p.StartDate), EndDate: dateParam(p.EndDate), GroupBy: string(p.GroupBy)})
}
func (h *Transactions) breakdown(c *echo.Context, scope service.TransactionScope, filter service.ReportFilter) error {
	item, err := h.service.GetReportBreakdown(c.Request().Context(), scope, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.ReportBreakdownResponse{Data: reportBreakdownPayload(item)})
}

func dateParam(value *openapi_types.Date) string {
	if value == nil {
		return ""
	}
	return value.Time.Format(time.DateOnly)
}
func reportCategoryPayload(items []service.ReportCategoryTotal) []api.ReportCategoryTotal {
	result := make([]api.ReportCategoryTotal, 0, len(items))
	for _, item := range items {
		result = append(result, api.ReportCategoryTotal{CategoryId: item.CategoryID, Name: item.Name, Type: api.TransactionType(item.Type), Amount: api.Money(item.Amount)})
	}
	return result
}
func reportSummaryPayload(item service.ReportSummary) api.ReportSummary {
	return api.ReportSummary{Period: api.ReportPeriod{StartDate: openapi_types.Date{Time: item.StartDate}, EndDate: openapi_types.Date{Time: item.EndDate}}, Income: api.Money(item.Income), Expense: api.Money(item.Expense), Difference: item.Difference, TopIncomeCategories: reportCategoryPayload(item.TopIncomeCategories), TopExpenseCategories: reportCategoryPayload(item.TopExpenseCategories)}
}
func reportBreakdownPayload(item service.ReportBreakdown) api.ReportBreakdown {
	periods := make([]api.ReportPeriodTotal, 0, len(item.Periods))
	for _, value := range item.Periods {
		periods = append(periods, api.ReportPeriodTotal{PeriodStart: openapi_types.Date{Time: value.PeriodStart}, Income: api.Money(value.Income), Expense: api.Money(value.Expense), Difference: value.Difference})
	}
	return api.ReportBreakdown{Period: api.ReportPeriod{StartDate: openapi_types.Date{Time: item.StartDate}, EndDate: openapi_types.Date{Time: item.EndDate}}, GroupBy: api.ReportBreakdownGroupBy(item.GroupBy), Periods: periods, Categories: reportCategoryPayload(item.Categories)}
}

type Transactions struct{ service transactionService }

func NewTransactions(transactionService transactionService) *Transactions {
	return &Transactions{service: transactionService}
}

func (h *Transactions) List(c *echo.Context, params api.ListTransactionsParams) error {
	return h.list(c, h.personalScope(c), transactionFilter(params.StartDate.Time, params.EndDate.Time, params.Type, params.CategoryId, params.IsDelete, params.Limit, params.Cursor))
}
func (h *Transactions) AdminList(c *echo.Context, owner int64, params api.AdminListTransactionsParams) error {
	return h.list(c, h.adminScope(c, owner), transactionFilter(params.StartDate.Time, params.EndDate.Time, params.Type, params.CategoryId, params.IsDelete, params.Limit, params.Cursor))
}
func (h *Transactions) list(c *echo.Context, scope service.TransactionScope, filter service.TransactionFilter) error {
	page, err := h.service.List(c.Request().Context(), scope, filter)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionListResponse{Data: transactionPayloads(page.Items), Page: api.Page{NextCursor: page.NextCursor}})
}

func (h *Transactions) Get(c *echo.Context, id int64, params api.GetTransactionParams) error {
	return h.get(c, h.personalScope(c), id, params.IsDelete != nil && *params.IsDelete)
}
func (h *Transactions) AdminGet(c *echo.Context, owner, id int64, params api.AdminGetTransactionParams) error {
	return h.get(c, h.adminScope(c, owner), id, params.IsDelete != nil && *params.IsDelete)
}
func (h *Transactions) get(c *echo.Context, scope service.TransactionScope, id int64, deleted bool) error {
	item, err := h.service.Get(c.Request().Context(), scope, id, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionResponse{Data: transactionPayload(item)})
}

func (h *Transactions) Create(c *echo.Context) error { return h.create(c, h.personalScope(c)) }
func (h *Transactions) AdminCreate(c *echo.Context, owner int64) error {
	return h.create(c, h.adminScope(c, owner))
}
func (h *Transactions) create(c *echo.Context, scope service.TransactionScope) error {
	var request api.TransactionCreateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	result, err := h.service.Create(c.Request().Context(), scope, service.TransactionInput{TransactionDate: request.TransactionDate.Time.Format(time.DateOnly), Type: string(request.Type), CategoryID: request.CategoryId, Amount: request.Amount, Title: request.Title, ClientRequestID: request.ClientRequestId})
	if err != nil {
		return h.writeError(c, err)
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	return c.JSON(status, api.TransactionResponse{Data: transactionPayload(result.Transaction)})
}

func (h *Transactions) Update(c *echo.Context, id int64) error {
	return h.update(c, h.personalScope(c), id)
}
func (h *Transactions) AdminUpdate(c *echo.Context, owner, id int64) error {
	return h.update(c, h.adminScope(c, owner), id)
}
func (h *Transactions) update(c *echo.Context, scope service.TransactionScope, id int64) error {
	var request api.TransactionUpdateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Update(c.Request().Context(), scope, id, service.TransactionInput{TransactionDate: request.TransactionDate.Time.Format(time.DateOnly), Type: string(request.Type), CategoryID: request.CategoryId, Amount: request.Amount, Title: request.Title, Version: request.Version})
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionResponse{Data: transactionPayload(item)})
}

func (h *Transactions) Delete(c *echo.Context, id int64, ifMatch string) error {
	return h.delete(c, h.personalScope(c), id, ifMatch)
}
func (h *Transactions) AdminDelete(c *echo.Context, owner, id int64, ifMatch string) error {
	return h.delete(c, h.adminScope(c, owner), id, ifMatch)
}
func (h *Transactions) delete(c *echo.Context, scope service.TransactionScope, id int64, ifMatch string) error {
	version, err := ifMatchVersion(ifMatch)
	if err != nil {
		return categoryBadRequest(c)
	}
	if err := h.service.SoftDeleteTransaction(c.Request().Context(), scope, id, version); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Transactions) Restore(c *echo.Context, id int64) error {
	return h.restore(c, h.personalScope(c), id)
}
func (h *Transactions) AdminRestore(c *echo.Context, owner, id int64) error {
	return h.restore(c, h.adminScope(c, owner), id)
}
func (h *Transactions) restore(c *echo.Context, scope service.TransactionScope, id int64) error {
	var request api.VersionRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.RestoreTransaction(c.Request().Context(), scope, id, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionResponse{Data: transactionPayload(item)})
}

func (h *Transactions) Daily(c *echo.Context, params api.ListDailySummariesParams) error {
	return h.daily(c, h.personalScope(c), params.StartDate.Time, params.EndDate.Time, params.Limit, params.Cursor)
}
func (h *Transactions) AdminDaily(c *echo.Context, owner int64, params api.AdminListDailySummariesParams) error {
	return h.daily(c, h.adminScope(c, owner), params.StartDate.Time, params.EndDate.Time, params.Limit, params.Cursor)
}
func (h *Transactions) daily(c *echo.Context, scope service.TransactionScope, start, end time.Time, limit *int32, cursor *string) error {
	page, err := h.service.ListDailySummaries(c.Request().Context(), scope, start.Format(time.DateOnly), end.Format(time.DateOnly), valueOr(limit, 0), valueOr(cursor, ""))
	if err != nil {
		return h.writeError(c, err)
	}
	data := make([]api.DailySummary, 0, len(page.Items))
	for _, item := range page.Items {
		data = append(data, api.DailySummary{Date: openapi_types.Date{Time: item.Date}, Income: item.Income, Expense: item.Expense, Difference: item.Difference})
	}
	return c.JSON(http.StatusOK, api.DailySummaryListResponse{Data: data, Page: api.Page{NextCursor: page.NextCursor}})
}

func (h *Transactions) personalScope(c *echo.Context) service.TransactionScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.TransactionScope{Actor: actor, Owner: actor.UserID, RequestID: requestID(c)}
}
func (h *Transactions) adminScope(c *echo.Context, owner int64) service.TransactionScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.TransactionScope{Actor: actor, Owner: owner, Admin: true, RequestID: requestID(c)}
}
func requestID(c *echo.Context) string {
	if value := c.Response().Header().Get(echo.HeaderXRequestID); value != "" {
		return value
	}
	return c.Request().Header.Get(echo.HeaderXRequestID)
}

func (h *Transactions) writeError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrAuthentication):
		return authenticationError(c)
	case errors.Is(err, service.ErrForbidden):
		return forbiddenError(c)
	case errors.Is(err, service.ErrValidation):
		return validationError(c)
	case errors.Is(err, service.ErrNotFound):
		return c.JSON(http.StatusNotFound, errorBody("NOT_FOUND", "Not found."))
	case errors.Is(err, service.ErrConflict):
		return c.JSON(http.StatusConflict, errorBody("CONFLICT", "Conflict."))
	case errors.Is(err, service.ErrUnavailable):
		return c.JSON(http.StatusServiceUnavailable, errorBody("SERVICE_UNAVAILABLE", "Service unavailable."))
	default:
		return internalError(c)
	}
}

func transactionFilter(start, end time.Time, kind *api.TransactionType, categoryID *int64, deleted *bool, limit *int32, cursor *string) service.TransactionFilter {
	filter := service.TransactionFilter{StartDate: start.Format(time.DateOnly), EndDate: end.Format(time.DateOnly), Limit: valueOr(limit, 0), Cursor: valueOr(cursor, "")}
	if kind != nil {
		filter.Type = string(*kind)
	}
	filter.CategoryID = categoryID
	filter.Deleted = deleted != nil && *deleted
	return filter
}
func transactionPayload(item service.Transaction) api.Transaction {
	return api.Transaction{Id: item.ID, UserId: item.UserID, CategoryId: item.CategoryID, CategoryName: item.CategoryName, TransactionDate: openapi_types.Date{Time: item.TransactionDate}, Type: api.TransactionType(item.Type), Amount: item.Amount, Title: item.Title, ClientRequestId: item.ClientRequestID, CreatedBy: item.CreatedBy, UpdatedBy: item.UpdatedBy, IsDelete: item.IsDelete, Version: item.Version, CreatedAt: item.CreatedAt.UTC(), UpdatedAt: item.UpdatedAt.UTC()}
}
func transactionPayloads(items []service.Transaction) []api.Transaction {
	result := make([]api.Transaction, 0, len(items))
	for _, item := range items {
		result = append(result, transactionPayload(item))
	}
	return result
}
func valueOr[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}
