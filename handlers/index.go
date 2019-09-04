package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

// Index returns the running API version.
func Index(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
