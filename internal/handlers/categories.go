package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type categoryService interface {
	List(context.Context, service.CategoryScope, string, bool) ([]service.Category, error)
	Get(context.Context, service.CategoryScope, uuid.UUID, bool) (service.Category, error)
	Create(context.Context, service.CategoryScope, string, string) (service.Category, error)
	Update(context.Context, service.CategoryScope, uuid.UUID, string, int32) (service.Category, error)
	Archive(context.Context, service.CategoryScope, uuid.UUID, int32) error
	Restore(context.Context, service.CategoryScope, uuid.UUID, int32) (service.Category, error)
}

// Categories exposes personal and explicitly-scoped admin category routes.
type Categories struct{ service categoryService }

func NewCategories(categoryService categoryService) *Categories {
	return &Categories{service: categoryService}
}

func (h *Categories) List(c *echo.Context, params api.ListCategoriesParams) error {
	categoryType := ""
	if params.Type != nil {
		categoryType = string(*params.Type)
	}
	deleted := params.IsDelete != nil && *params.IsDelete
	items, err := h.service.List(c.Request().Context(), h.personalScope(c), categoryType, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponses(items)})
}

func (h *Categories) Create(c *echo.Context) error {
	var request categoryCreateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Create(c.Request().Context(), h.personalScope(c), request.Type, request.Name)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) Get(c *echo.Context, id uuid.UUID, params api.GetCategoryParams) error {
	deleted := params.IsDelete != nil && *params.IsDelete
	item, err := h.service.Get(c.Request().Context(), h.personalScope(c), id, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) Update(c *echo.Context, id uuid.UUID) error {
	var request categoryUpdateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Update(c.Request().Context(), h.personalScope(c), id, request.Name, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) Delete(c *echo.Context, id uuid.UUID, ifMatch string) error {
	version, err := ifMatchVersion(ifMatch)
	if err != nil {
		return categoryBadRequest(c)
	}
	if err := h.service.Archive(c.Request().Context(), h.personalScope(c), id, version); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Categories) Restore(c *echo.Context, id uuid.UUID) error {
	var request categoryRestoreRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Restore(c.Request().Context(), h.personalScope(c), id, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) AdminList(c *echo.Context, owner uuid.UUID, params api.AdminListCategoriesParams) error {
	categoryType := ""
	if params.Type != nil {
		categoryType = string(*params.Type)
	}
	deleted := params.IsDelete != nil && *params.IsDelete
	items, err := h.service.List(c.Request().Context(), h.adminScope(c, owner), categoryType, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponses(items)})
}

func (h *Categories) AdminCreate(c *echo.Context, owner uuid.UUID) error {
	var request categoryCreateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Create(c.Request().Context(), h.adminScope(c, owner), request.Type, request.Name)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) AdminGet(c *echo.Context, owner, id uuid.UUID, params api.AdminGetCategoryParams) error {
	deleted := params.IsDelete != nil && *params.IsDelete
	item, err := h.service.Get(c.Request().Context(), h.adminScope(c, owner), id, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) AdminUpdate(c *echo.Context, owner, id uuid.UUID) error {
	var request categoryUpdateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Update(c.Request().Context(), h.adminScope(c, owner), id, request.Name, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) AdminDelete(c *echo.Context, owner, id uuid.UUID, ifMatch string) error {
	version, err := ifMatchVersion(ifMatch)
	if err != nil {
		return categoryBadRequest(c)
	}
	if err := h.service.Archive(c.Request().Context(), h.adminScope(c, owner), id, version); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Categories) AdminRestore(c *echo.Context, owner, id uuid.UUID) error {
	var request categoryRestoreRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Restore(c.Request().Context(), h.adminScope(c, owner), id, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": categoryResponse(item)})
}

func (h *Categories) personalScope(c *echo.Context) service.CategoryScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.CategoryScope{Actor: actor, Owner: actor.UserID}
}

func (h *Categories) adminScope(c *echo.Context, owner uuid.UUID) service.CategoryScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.CategoryScope{Actor: actor, Owner: owner, Admin: true}
}

func (h *Categories) writeError(c *echo.Context, err error) error {
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
	default:
		return internalError(c)
	}
}

func categoryBadRequest(c *echo.Context) error {
	return c.JSON(http.StatusBadRequest, errorBody("BAD_REQUEST", "Invalid request."))
}

type categoryCreateRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type categoryUpdateRequest struct {
	Name    string `json:"name"`
	Version int32  `json:"version"`
}

type categoryRestoreRequest struct {
	Version int32 `json:"version"`
}

type categoryPayload struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	IsDelete bool   `json:"isDelete"`
	Version  int32  `json:"version"`
}

func categoryResponse(item service.Category) categoryPayload {
	return categoryPayload{ID: item.ID.String(), UserID: item.UserID.String(), Type: item.Type, Name: item.Name, IsDelete: item.IsDelete, Version: item.Version}
}

func categoryResponses(items []service.Category) []categoryPayload {
	result := make([]categoryPayload, 0, len(items))
	for _, item := range items {
		result = append(result, categoryResponse(item))
	}
	return result
}
