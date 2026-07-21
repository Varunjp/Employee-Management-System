package handler

import (
	"employee_management/pkg/auth"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// AuthHandler issues JWTs for the demo credentials configured via
// ADMIN_USERNAME / ADMIN_PASSWORD. In a production system this would
// delegate to a real user store with hashed passwords; it is kept simple
// here since authentication is a bonus requirement, not the core feature.
type AuthHandler struct {
	adminUsername string
	adminPassword string
	jwtSecret     string
	jwtTTL        time.Duration
}

type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin123"`
} //@name LoginRequest

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int64  `json:"expires_in" example:"3600"`
} //@name LoginResponse

// NewAuthHandler wires the credentials, JWT signing secret, and token
// lifetime used to issue access tokens from the /auth/login endpoint.
func NewAuthHandler(adminUsername, adminPassword, jwtSecret string, jwtTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		adminUsername: adminUsername,
		adminPassword: adminPassword,
		jwtSecret:     jwtSecret,
		jwtTTL:        jwtTTL,
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Username != h.adminUsername || req.Password != h.adminPassword {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}

	token, err := auth.GenerateToken(h.jwtSecret, req.Username, h.jwtTTL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(http.StatusOK, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.jwtTTL.Seconds()),
	})
}
