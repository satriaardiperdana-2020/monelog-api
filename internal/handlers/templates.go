package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type templateService interface {
	List(context.Context, service.TemplateScope, bool) ([]service.TransactionTemplate, error)
	Get(context.Context, service.TemplateScope, int64, bool) (service.TransactionTemplate, error)
	Create(context.Context, service.TemplateScope, service.TemplateInput) (service.TransactionTemplate, error)
	Update(context.Context, service.TemplateScope, int64, service.TemplateInput) (service.TransactionTemplate, error)
	SoftDelete(context.Context, service.TemplateScope, int64, int32) error
	Restore(context.Context, service.TemplateScope, int64, int32) (service.TransactionTemplate, error)
	Apply(context.Context, service.TemplateScope, int64) (service.TemplateDraft, error)
}

// Templates exposes personal and explicitly-scoped admin template routes.
type Templates struct{ service templateService }

func NewTemplates(templateService templateService) *Templates {
	return &Templates{service: templateService}
}

func (h *Templates) List(c *echo.Context, params api.ListTransactionTemplatesParams) error {
	return h.list(c, h.personalScope(c), params.IsDelete != nil && *params.IsDelete)
}

func (h *Templates) AdminList(c *echo.Context, owner int64, params api.AdminListTransactionTemplatesParams) error {
	return h.list(c, h.adminScope(c, owner), params.IsDelete != nil && *params.IsDelete)
}

func (h *Templates) list(c *echo.Context, scope service.TemplateScope, deleted bool) error {
	items, err := h.service.List(c.Request().Context(), scope, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionTemplateListResponse{Data: templatePayloads(items)})
}

func (h *Templates) Create(c *echo.Context) error { return h.create(c, h.personalScope(c)) }
func (h *Templates) AdminCreate(c *echo.Context, owner int64) error {
	return h.create(c, h.adminScope(c, owner))
}
func (h *Templates) create(c *echo.Context, scope service.TemplateScope) error {
	var request api.TransactionTemplateCreateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Create(c.Request().Context(), scope, templateInput(request.Name, request.CategoryId, string(request.Type), request.Amount, request.Title, 0))
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusCreated, api.TransactionTemplateResponse{Data: templatePayload(item)})
}

func (h *Templates) Get(c *echo.Context, id int64, params api.GetTransactionTemplateParams) error {
	return h.get(c, h.personalScope(c), id, params.IsDelete != nil && *params.IsDelete)
}
func (h *Templates) AdminGet(c *echo.Context, owner, id int64, params api.AdminGetTransactionTemplateParams) error {
	return h.get(c, h.adminScope(c, owner), id, params.IsDelete != nil && *params.IsDelete)
}
func (h *Templates) get(c *echo.Context, scope service.TemplateScope, id int64, deleted bool) error {
	item, err := h.service.Get(c.Request().Context(), scope, id, deleted)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionTemplateResponse{Data: templatePayload(item)})
}

func (h *Templates) Update(c *echo.Context, id int64) error {
	return h.update(c, h.personalScope(c), id)
}
func (h *Templates) AdminUpdate(c *echo.Context, owner, id int64) error {
	return h.update(c, h.adminScope(c, owner), id)
}
func (h *Templates) update(c *echo.Context, scope service.TemplateScope, id int64) error {
	var request api.TransactionTemplateUpdateRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Update(c.Request().Context(), scope, id, templateInput(request.Name, request.CategoryId, string(request.Type), request.Amount, request.Title, request.Version))
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionTemplateResponse{Data: templatePayload(item)})
}

func (h *Templates) Delete(c *echo.Context, id int64, ifMatch string) error {
	return h.delete(c, h.personalScope(c), id, ifMatch)
}
func (h *Templates) AdminDelete(c *echo.Context, owner, id int64, ifMatch string) error {
	return h.delete(c, h.adminScope(c, owner), id, ifMatch)
}
func (h *Templates) delete(c *echo.Context, scope service.TemplateScope, id int64, ifMatch string) error {
	version, err := ifMatchVersion(ifMatch)
	if err != nil {
		return categoryBadRequest(c)
	}
	if err := h.service.SoftDelete(c.Request().Context(), scope, id, version); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Templates) Restore(c *echo.Context, id int64) error {
	return h.restore(c, h.personalScope(c), id)
}
func (h *Templates) AdminRestore(c *echo.Context, owner, id int64) error {
	return h.restore(c, h.adminScope(c, owner), id)
}
func (h *Templates) restore(c *echo.Context, scope service.TemplateScope, id int64) error {
	var request api.VersionRequest
	if err := decodeJSON(c, &request); err != nil {
		return categoryBadRequest(c)
	}
	item, err := h.service.Restore(c.Request().Context(), scope, id, request.Version)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionTemplateResponse{Data: templatePayload(item)})
}

func (h *Templates) Apply(c *echo.Context, id int64) error { return h.apply(c, h.personalScope(c), id) }
func (h *Templates) AdminApply(c *echo.Context, owner, id int64) error {
	return h.apply(c, h.adminScope(c, owner), id)
}
func (h *Templates) apply(c *echo.Context, scope service.TemplateScope, id int64) error {
	draft, err := h.service.Apply(c.Request().Context(), scope, id)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, api.TransactionTemplateApplyResponse{Data: api.TransactionTemplateDraft{CategoryId: draft.CategoryID, Type: api.TransactionType(draft.Type), Amount: api.Money(draft.Amount), Title: draft.Title}})
}

func (h *Templates) personalScope(c *echo.Context) service.TemplateScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.TemplateScope{Actor: actor, Owner: actor.UserID}
}
func (h *Templates) adminScope(c *echo.Context, owner int64) service.TemplateScope {
	actor, _ := middleware.Actor(c.Request().Context())
	return service.TemplateScope{Actor: actor, Owner: owner, Admin: true}
}
func (h *Templates) writeError(c *echo.Context, err error) error {
	return (&Categories{}).writeError(c, err)
}

func templateInput(name string, categoryID int64, kind, amount, title string, version int32) service.TemplateInput {
	return service.TemplateInput{Name: name, CategoryID: categoryID, Type: kind, Amount: amount, Title: title, Version: version}
}

func templatePayload(item service.TransactionTemplate) api.TransactionTemplate {
	return api.TransactionTemplate{Id: item.ID, UserId: item.UserID, CategoryId: item.CategoryID, Type: api.TransactionType(item.Type), Name: item.Name, Amount: api.Money(item.Amount), Title: item.Title, IsDelete: item.IsDelete, Version: item.Version}
}
func templatePayloads(items []service.TransactionTemplate) []api.TransactionTemplate {
	result := make([]api.TransactionTemplate, 0, len(items))
	for _, item := range items {
		result = append(result, templatePayload(item))
	}
	return result
}
