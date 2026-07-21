package http

import (
	"employee_management/internal/delivery/http/handler"
	appmw "employee_management/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "employee_management/docs"
)

// Dependencies bundles every handler and the JWT secret the router needs to
// wire up routes. It is populated in cmd/api/main.go, which is the single
// place in the codebase that constructs concrete infrastructure.
type Dependencies struct {
	EmployeeHandler *handler.EmployeeHandler
	AuthHandler     *handler.AuthHandler
	HealthHandler   *handler.HealthHandler
	JWTSecret       string
}

func NewRouter(deps Dependencies) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = appmw.NewHTTPErrorHandler()

	e.Use(echomw.Recover())
	e.Use(echomw.Logger())
	e.Use(echomw.CORS())
	e.Use(echomw.RequestID())

	e.GET("/health", deps.HealthHandler.Check)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/login", deps.AuthHandler.Login)

	employees := api.Group("/employees")
	employees.GET("", deps.EmployeeHandler.List)
	employees.GET("/:id", deps.EmployeeHandler.GetByID)

	jwtGuard := appmw.JWTAuth(deps.JWTSecret)
	employees.POST("", deps.EmployeeHandler.Create, jwtGuard)
	employees.PUT("/:id", deps.EmployeeHandler.Update, jwtGuard)
	employees.DELETE("/:id", deps.EmployeeHandler.Delete, jwtGuard)

	return e
}
