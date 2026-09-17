package handlers

import (
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
)

// openAPIServer adapts existing handlers to the generated Echo interface.
type openAPIServer struct {
	health       *Health
	auth         *Auth
	categories   *Categories
	transactions *Transactions
}

var _ api.ServerInterface = (*openAPIServer)(nil)

func newOpenAPIServer(health *Health, auth *Auth, categories *Categories, transactions *Transactions) *openAPIServer {
	return &openAPIServer{health: health, auth: auth, categories: categories, transactions: transactions}
}

func (s *openAPIServer) HealthLive(c *echo.Context) error  { return s.health.Live(c) }
func (s *openAPIServer) HealthReady(c *echo.Context) error { return s.health.Ready(c) }
func (s *openAPIServer) Register(c *echo.Context) error    { return s.auth.Register(c) }
func (s *openAPIServer) Login(c *echo.Context) error       { return s.auth.Login(c) }
func (s *openAPIServer) Refresh(c *echo.Context) error     { return s.auth.Refresh(c) }
func (s *openAPIServer) Logout(c *echo.Context) error      { return s.auth.Logout(c) }
func (s *openAPIServer) GetMe(c *echo.Context) error       { return s.auth.Me(c) }
func (s *openAPIServer) UpdateMe(c *echo.Context) error    { return s.auth.UpdateProfile(c) }

func (s *openAPIServer) DeleteMe(c *echo.Context, _ api.DeleteMeParams) error {
	return s.auth.DeleteMe(c)
}

func (s *openAPIServer) ListCategories(c *echo.Context, p api.ListCategoriesParams) error {
	return s.categories.List(c, p)
}
func (s *openAPIServer) CreateCategory(c *echo.Context) error { return s.categories.Create(c) }
func (s *openAPIServer) GetCategory(c *echo.Context, id api.ResourceId, p api.GetCategoryParams) error {
	return s.categories.Get(c, id, p)
}
func (s *openAPIServer) UpdateCategory(c *echo.Context, id api.ResourceId) error {
	return s.categories.Update(c, id)
}
func (s *openAPIServer) DeleteCategory(c *echo.Context, id api.ResourceId, p api.DeleteCategoryParams) error {
	return s.categories.Delete(c, id, p.IfMatch)
}
func (s *openAPIServer) RestoreCategory(c *echo.Context, id api.ResourceId) error {
	return s.categories.Restore(c, id)
}
func (s *openAPIServer) AdminListCategories(c *echo.Context, owner api.UserId, p api.AdminListCategoriesParams) error {
	return s.categories.AdminList(c, owner, p)
}
func (s *openAPIServer) AdminCreateCategory(c *echo.Context, owner api.UserId) error {
	return s.categories.AdminCreate(c, owner)
}
func (s *openAPIServer) AdminGetCategory(c *echo.Context, owner api.UserId, id api.ResourceId, p api.AdminGetCategoryParams) error {
	return s.categories.AdminGet(c, owner, id, p)
}
func (s *openAPIServer) AdminUpdateCategory(c *echo.Context, owner api.UserId, id api.ResourceId) error {
	return s.categories.AdminUpdate(c, owner, id)
}
func (s *openAPIServer) AdminDeleteCategory(c *echo.Context, owner api.UserId, id api.ResourceId, p api.AdminDeleteCategoryParams) error {
	return s.categories.AdminDelete(c, owner, id, p.IfMatch)
}
func (s *openAPIServer) AdminRestoreCategory(c *echo.Context, owner api.UserId, id api.ResourceId) error {
	return s.categories.AdminRestore(c, owner, id)
}

func (s *openAPIServer) ListTransactions(c *echo.Context, p api.ListTransactionsParams) error {
	return s.transactions.List(c, p)
}
func (s *openAPIServer) CreateTransaction(c *echo.Context) error { return s.transactions.Create(c) }
func (s *openAPIServer) GetTransaction(c *echo.Context, id api.ResourceId, p api.GetTransactionParams) error {
	return s.transactions.Get(c, id, p)
}
func (s *openAPIServer) UpdateTransaction(c *echo.Context, id api.ResourceId) error {
	return s.transactions.Update(c, id)
}
func (s *openAPIServer) DeleteTransaction(c *echo.Context, id api.ResourceId, p api.DeleteTransactionParams) error {
	return s.transactions.Delete(c, id, p.IfMatch)
}
func (s *openAPIServer) RestoreTransaction(c *echo.Context, id api.ResourceId) error {
	return s.transactions.Restore(c, id)
}
func (s *openAPIServer) ListDailySummaries(c *echo.Context, p api.ListDailySummariesParams) error {
	return s.transactions.Daily(c, p)
}
func (s *openAPIServer) AdminListTransactions(c *echo.Context, owner api.UserId, p api.AdminListTransactionsParams) error {
	return s.transactions.AdminList(c, owner, p)
}
func (s *openAPIServer) AdminCreateTransaction(c *echo.Context, owner api.UserId) error {
	return s.transactions.AdminCreate(c, owner)
}
func (s *openAPIServer) AdminGetTransaction(c *echo.Context, owner api.UserId, id api.ResourceId, p api.AdminGetTransactionParams) error {
	return s.transactions.AdminGet(c, owner, id, p)
}
func (s *openAPIServer) AdminUpdateTransaction(c *echo.Context, owner api.UserId, id api.ResourceId) error {
	return s.transactions.AdminUpdate(c, owner, id)
}
func (s *openAPIServer) AdminDeleteTransaction(c *echo.Context, owner api.UserId, id api.ResourceId, p api.AdminDeleteTransactionParams) error {
	return s.transactions.AdminDelete(c, owner, id, p.IfMatch)
}
func (s *openAPIServer) AdminRestoreTransaction(c *echo.Context, owner api.UserId, id api.ResourceId) error {
	return s.transactions.AdminRestore(c, owner, id)
}
func (s *openAPIServer) AdminListDailySummaries(c *echo.Context, owner api.UserId, p api.AdminListDailySummariesParams) error {
	return s.transactions.AdminDaily(c, owner, p)
}

func (s *openAPIServer) GetReportSummary(c *echo.Context, p api.GetReportSummaryParams) error {
	return s.transactions.Summary(c, p)
}
func (s *openAPIServer) GetReportBreakdown(c *echo.Context, p api.GetReportBreakdownParams) error {
	return s.transactions.Breakdown(c, p)
}
func (s *openAPIServer) AdminGetReportSummary(c *echo.Context, owner api.UserId, p api.AdminGetReportSummaryParams) error {
	return s.transactions.AdminSummary(c, owner, p)
}
func (s *openAPIServer) AdminGetReportBreakdown(c *echo.Context, owner api.UserId, p api.AdminGetReportBreakdownParams) error {
	return s.transactions.AdminBreakdown(c, owner, p)
}
