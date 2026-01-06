package probe

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Mount(e *echo.Echo) {
	e.GET("/healthz", h.health)
}

func (h *Handler) health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
