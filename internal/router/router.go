package router

import (
	"blog-server/internal/config"
	"blog-server/internal/controller"
	"blog-server/internal/dao"
	"blog-server/internal/ent"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, client *ent.Client) *gin.Engine {
	engine := gin.Default()

	userDAO := dao.NewEntUserDAO(client)
	dictItemDAO := dao.NewEntDictItemDAO(client)
	operationLogDAO := dao.NewEntOperationLogDAO(client)
	authService := service.NewAuthService(
		userDAO,
		dictItemDAO,
		operationLogDAO,
		cfg.JWTSecret,
		cfg.TokenDuration,
	)
	authController := controller.NewAuthController(authService)

	admin := engine.Group("/api/admin")
	{
		auth := admin.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/logout", authController.Logout)
			auth.POST("/refresh", authController.Refresh)
			auth.GET("/me", authController.Me)
		}

		admin.GET("/dict-items", authController.ListDictItems)
		admin.POST("/dict-items", authController.CreateDictItem)
		admin.POST("/dict-items/:id/update", authController.UpdateDictItem)
		admin.POST("/dict-items/:id/delete", authController.DeleteDictItem)
		admin.GET("/operation-logs", authController.ListOperationLogs)
	}

	return engine
}
