package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
)

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Monelog API documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <p style="font-family: sans-serif; margin: 1rem 2rem;">Login melalui <code>POST /api/v1/auth/login</code>, salin <code>access_token</code>, lalu klik <strong>Authorize</strong> dan tempel token tersebut. Swagger UI akan mengirimkannya sebagai header <code>Authorization: Bearer &lt;token&gt;</code> ke endpoint yang membutuhkan autentikasi.</p>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>window.ui = SwaggerUIBundle({ url: '/api/openapi.json', dom_id: '#swagger-ui', deepLinking: true, persistAuthorization: true, tryItOutEnabled: true });</script>
</body>
</html>`

// RegisterSwagger exposes the generated API contract and an interactive UI.
func RegisterSwagger(e *echo.Echo) {
	e.GET("/api/openapi.json", openAPIDocument)
	e.GET("/swagger", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, "/swagger/")
	})
	e.GET("/swagger/", func(c *echo.Context) error {
		return c.HTML(http.StatusOK, swaggerUIHTML)
	})
}

func openAPIDocument(c *echo.Context) error {
	document, err := api.GetSwagger()
	if err != nil {
		return internalError(c)
	}
	return c.JSON(http.StatusOK, document)
}
