package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestSwaggerRoutesExposeGeneratedContractAndUI(t *testing.T) {
	e := echo.New()
	RegisterSwagger(e)

	contract := httptest.NewRecorder()
	e.ServeHTTP(contract, httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
	if contract.Code != http.StatusOK || !strings.Contains(contract.Body.String(), "Monelog API") || !strings.Contains(contract.Body.String(), "/api/v1/auth/login") {
		t.Fatalf("contract status=%d body=%s", contract.Code, contract.Body.String())
	}

	ui := httptest.NewRecorder()
	e.ServeHTTP(ui, httptest.NewRequest(http.MethodGet, "/swagger/", nil))
	if ui.Code != http.StatusOK || !strings.Contains(ui.Body.String(), "SwaggerUIBundle") || !strings.Contains(ui.Body.String(), "/api/openapi.json") {
		t.Fatalf("Swagger UI status=%d body=%s", ui.Code, ui.Body.String())
	}
}
