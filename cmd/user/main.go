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
	"user/models"
	"user/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	log.SetupLogger()

	db := resource.InitDB(&cfg)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Logger.Fatalf("failed to migrate schema: %v", err)
	}

	redisClient := resource.InitRedis(&cfg)

	userRepo := repository.NewUserRepository(db, redisClient)
	userService := service.NewUserService(*userRepo)
	userUsecase := usecase.NewUserUsecase(*userService, cfg.Secret.JWTSecret)
	userHandler := handler.NewUserHandler(userUsecase)

	router := gin.Default()
	routes.SetupRoutes(router, *userHandler, cfg.Secret.JWTSecret)

	if err := router.Run(fmt.Sprintf(":%s", cfg.App.Port)); err != nil {
		log.Logger.Fatalf("failed to run server: %v", err)
	}
}
