package routes

import (
	"time"
	"user/cmd/user/handler"
	"user/middleware"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRoutes(router *gin.Engine, userHandler handler.UserHandler, jwtSecret string, requestTimeout time.Duration) {
	//Public API
	public := router.Group("/api")
	public.Use(middleware.RequestLogger(requestTimeout))
	public.Use(otelgin.Middleware("user"))
	public.GET("/ping", userHandler.Ping)
	public.POST("/v1/register", userHandler.Register)
	public.POST("/v1/login", userHandler.Login)

	// Private API
	authMiddleware := middleware.AuthMiddleware(jwtSecret)
	private := router.Group("/api")
	private.Use(middleware.RequestLogger(requestTimeout))
	private.Use(otelgin.Middleware("user"))
	private.Use(authMiddleware)
	private.GET("/v1/user_info", userHandler.GetUserInfo)
}
