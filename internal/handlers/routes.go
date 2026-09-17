package handlers

import (
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
)

// RegisterRoutes registers the OpenAPI-defined HTTP routes.
func RegisterRoutes(e *echo.Echo, health *Health, auth *Auth, categories *Categories, transactions *Transactions, authenticate echo.MiddlewareFunc) {
	RegisterSwagger(e)
	if auth == nil || authenticate == nil {
		e.GET("/health/live", health.Live)
		e.GET("/health/ready", health.Ready)
		return
	}
	api.RegisterHandlersWithOptions(e, newOpenAPIServer(health, auth, categories, transactions), api.RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			"deleteMe":         {authenticate},
			"getMe":            {authenticate},
			"updateMe":         {authenticate},
			"getReportSummary": {authenticate}, "getReportBreakdown": {authenticate},
			"listCategories": {authenticate}, "createCategory": {authenticate}, "getCategory": {authenticate}, "updateCategory": {authenticate}, "deleteCategory": {authenticate}, "restoreCategory": {authenticate},
			"listTransactions": {authenticate}, "createTransaction": {authenticate}, "getTransaction": {authenticate}, "updateTransaction": {authenticate}, "deleteTransaction": {authenticate}, "restoreTransaction": {authenticate}, "listDailySummaries": {authenticate},
			"adminListCategories": {authenticate, middleware.RequireAdmin}, "adminCreateCategory": {authenticate, middleware.RequireAdmin}, "adminGetCategory": {authenticate, middleware.RequireAdmin}, "adminUpdateCategory": {authenticate, middleware.RequireAdmin}, "adminDeleteCategory": {authenticate, middleware.RequireAdmin}, "adminRestoreCategory": {authenticate, middleware.RequireAdmin},
			"adminListTransactions": {authenticate, middleware.RequireAdmin}, "adminCreateTransaction": {authenticate, middleware.RequireAdmin}, "adminGetTransaction": {authenticate, middleware.RequireAdmin}, "adminUpdateTransaction": {authenticate, middleware.RequireAdmin}, "adminDeleteTransaction": {authenticate, middleware.RequireAdmin}, "adminRestoreTransaction": {authenticate, middleware.RequireAdmin}, "adminListDailySummaries": {authenticate, middleware.RequireAdmin},
			"adminGetReportSummary": {authenticate, middleware.RequireAdmin}, "adminGetReportBreakdown": {authenticate, middleware.RequireAdmin},
		},
	})
}
