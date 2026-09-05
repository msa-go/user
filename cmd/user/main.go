package main

import (
	"fmt"

	"user/cmd/user/handler"
	"user/cmd/user/repository"
	"user/cmd/user/resource"
	"user/cmd/user/service"
	"user/cmd/user/usecase"
	"user/config"
	"user/infrastructure/log"
	"user/routes"
	"user/trace"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	log.SetupLogger()

	shutdownTracer, err := trace.InitTracer(cfg.Observability.ServiceName, cfg.Observability.OTLPEndpoint)
	if err != nil {
		log.Logger.Fatalf("failed to init tracer: %v", err)
	}
	defer shutdownTracer()

	db := resource.InitDB(&cfg)
	redisClient := resource.InitRedis(&cfg)

	userRepo := repository.NewUserRepository(db, redisClient)
	userService := service.NewUserService(*userRepo)
	userUsecase := usecase.NewUserUsecase(*userService, cfg.Secret.JWTSecret, cfg.App.TokenExpiry)
	userHandler := handler.NewUserHandler(userUsecase)

	router := gin.Default()
	routes.SetupRoutes(router, *userHandler, cfg.Secret.JWTSecret, cfg.App.RequestTimeout)

	if err := router.Run(fmt.Sprintf(":%s", cfg.App.Port)); err != nil {
		log.Logger.Fatalf("failed to run server: %v", err)
	}
}
