package handlers

import (
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
)

// openAPIServer adapts existing handlers to the generated Echo interface.
type openAPIServer struct {
	health *Health
	auth   *Auth
}

var _ api.ServerInterface = (*openAPIServer)(nil)

func newOpenAPIServer(health *Health, auth *Auth) *openAPIServer {
	return &openAPIServer{health: health, auth: auth}
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
