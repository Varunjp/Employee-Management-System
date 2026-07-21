package middleware

import (
	"employee_management/pkg/auth"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const authContextKey = "auth_username"

// JWTAuth returns an echo middleware that requires a valid
// "Authorization: Bearer <token>" header signed with secret. It is applied
// only to the mutating employee endpoints (create/update/delete); GET
// endpoints remain public per the assignment's read-heavy caching focus.
func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "authorization header must be 'Bearer <token>'")
			}

			claims, err := auth.ParseToken(secret, parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set(authContextKey, claims.Username)
			return next(c)
		}
	}
}
