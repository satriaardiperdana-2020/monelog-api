package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
)

// openAPIServer adapts existing handlers to the generated Echo interface.
type openAPIServer struct {
	health     *Health
	auth       *Auth
	categories *Categories
}

var _ api.ServerInterface = (*openAPIServer)(nil)

func newOpenAPIServer(health *Health, auth *Auth, categories *Categories) *openAPIServer {
	return &openAPIServer{health: health, auth: auth, categories: categories}
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

func (s *openAPIServer) ListCategories(c *echo.Context, params api.ListCategoriesParams) error {
	return s.categories.List(c, params)
}
func (s *openAPIServer) CreateCategory(c *echo.Context) error { return s.categories.Create(c) }
func (s *openAPIServer) GetCategory(c *echo.Context, id openapi_types.UUID, params api.GetCategoryParams) error {
	return s.categories.Get(c, uuid.UUID(id), params)
}
func (s *openAPIServer) UpdateCategory(c *echo.Context, id openapi_types.UUID) error {
	return s.categories.Update(c, uuid.UUID(id))
}
func (s *openAPIServer) DeleteCategory(c *echo.Context, id openapi_types.UUID, params api.DeleteCategoryParams) error {
	return s.categories.Delete(c, uuid.UUID(id), params.IfMatch)
}
func (s *openAPIServer) RestoreCategory(c *echo.Context, id openapi_types.UUID) error {
	return s.categories.Restore(c, uuid.UUID(id))
}
func (s *openAPIServer) AdminListCategories(c *echo.Context, userID openapi_types.UUID, params api.AdminListCategoriesParams) error {
	return s.categories.AdminList(c, uuid.UUID(userID), params)
}
func (s *openAPIServer) AdminCreateCategory(c *echo.Context, userID openapi_types.UUID) error {
	return s.categories.AdminCreate(c, uuid.UUID(userID))
}
func (s *openAPIServer) AdminGetCategory(c *echo.Context, userID, id openapi_types.UUID, params api.AdminGetCategoryParams) error {
	return s.categories.AdminGet(c, uuid.UUID(userID), uuid.UUID(id), params)
}
func (s *openAPIServer) AdminUpdateCategory(c *echo.Context, userID, id openapi_types.UUID) error {
	return s.categories.AdminUpdate(c, uuid.UUID(userID), uuid.UUID(id))
}
func (s *openAPIServer) AdminDeleteCategory(c *echo.Context, userID, id openapi_types.UUID, params api.AdminDeleteCategoryParams) error {
	return s.categories.AdminDelete(c, uuid.UUID(userID), uuid.UUID(id), params.IfMatch)
}
func (s *openAPIServer) AdminRestoreCategory(c *echo.Context, userID, id openapi_types.UUID) error {
	return s.categories.AdminRestore(c, uuid.UUID(userID), uuid.UUID(id))
}
