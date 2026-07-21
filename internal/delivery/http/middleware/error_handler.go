package middleware

import (
	"employee_management/internal/delivery/http/response"
	"employee_management/internal/domain"
	"employee_management/pkg/logger"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func NewHTTPErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status, message := mapError(err)
		if status >= http.StatusInternalServerError {
			logger.Error("unhandled error on %s %s: %v", c.Request().Method, c.Request().URL.Path, err)
		}

		if writeErr := c.JSON(status, response.NewError(message)); writeErr != nil {
			logger.Error("failed to write error response: %v", writeErr)
		}
	}
}

func mapError(err error) (int, string) {
	var he *echo.HTTPError
	if errors.As(err, &he) {
		if msg, ok := he.Message.(string); ok {
			return he.Code, msg
		}
		return he.Code, http.StatusText(he.Code)
	}

	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
