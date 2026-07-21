package main

import (
	"context"
	"employee_management/config"
	redisadapter "employee_management/internal/cache/redis"
	pgrepo "employee_management/internal/repository/postgres"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "employee_management/internal/delivery/http"
	"employee_management/internal/delivery/http/handler"

	"employee_management/internal/usecase"
	"employee_management/pkg/database"
	"employee_management/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config: %v", err)
	}

	pgPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		logger.Fatal("failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()
	logger.Info("connect to postgres at %s:%s", cfg.DBHost, cfg.DBPort)

	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info("connect to redis at %s", cfg.RedisAddr)

	employeeRepo := pgrepo.NewEmployeeRepository(pgPool)
	employeeCache := redisadapter.NewCache(redisClient)
	employeeUsecase := usecase.NewEmployeeUsecase(employeeRepo, employeeCache)

	deps := deliveryhttp.Dependencies{
		EmployeeHandler: handler.NewEmployeeHandler(employeeUsecase),
		AuthHandler:     handler.NewAuthHandler(cfg.AdminUsername, cfg.AdminPassword, cfg.JWTSecret, cfg.JWTExpiration),
		HealthHandler:   handler.NewHealthHandler(),
		JWTSecret:       cfg.JWTSecret,
	}

	e := deliveryhttp.NewRouter(deps)

	// Start the server in a goroutine so the main goroutine can listen for
	// shutdown signals and perform a graceful drain of in-flight requests.
	go func() {
		logger.Info("starting server on port %s ", cfg.AppPort)
		if err := e.Start(":" + cfg.AppPort); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received, draining in-flight requests...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed: %v", err)
	} else {
		logger.Info("server shut down cleanly")
	}
}
