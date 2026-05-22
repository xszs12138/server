package router

import (
	"blog-server/internal/config"
	"blog-server/internal/controller"
	"blog-server/internal/dao"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) *gin.Engine {
	engine := gin.Default()

	userDAO := dao.NewMemoryUserDAO(cfg.Admin)
	authService := service.NewAuthService(userDAO, cfg.JWTSecret, cfg.TokenDuration)
	authController := controller.NewAuthController(authService)

	admin := engine.Group("/api/admin")
	{
		auth := admin.Group("/auth")
		{
			auth.POST("/login", authController.Login)
		}
	}

	return engine
}
