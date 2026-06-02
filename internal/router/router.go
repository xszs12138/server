package router

import (
	"blog-server/internal/config"
	"blog-server/internal/controller"
	"blog-server/internal/dao"
	"blog-server/internal/ent"
	"blog-server/internal/imagebed"
	"blog-server/internal/livehub"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, client *ent.Client) *gin.Engine {
	engine := gin.Default()

	userDAO := dao.NewEntUserDAO(client)
	dictItemDAO := dao.NewEntDictItemDAO(client)
	operationLogDAO := dao.NewEntOperationLogDAO(client)
	postDAO := dao.NewEntPostDAO(client)
	categoryDAO := dao.NewEntCategoryDAO(client)
	tagDAO := dao.NewEntTagDAO(client)
	commentDAO := dao.NewEntCommentDAO(client)
	siteSettingDAO := dao.NewEntSiteSettingDAO(client)
	gameDAO := dao.NewEntGameDAO(client)
	musicTrackDAO := dao.NewEntMusicTrackDAO(client)

	authService := service.NewAuthService(
		userDAO,
		dictItemDAO,
		operationLogDAO,
		cfg.JWTSecret,
		cfg.TokenDuration,
	)
	dictService := service.NewDictService(dictItemDAO, siteSettingDAO, authService)
	postService := service.NewPostService(postDAO, categoryDAO, tagDAO, authService)
	categoryService := service.NewCategoryService(categoryDAO, postDAO, authService)
	tagService := service.NewTagService(tagDAO, postDAO, authService)
	commentService := service.NewCommentService(commentDAO, postDAO, authService)
	siteService := service.NewSiteService(siteSettingDAO, authService)
	gameService := service.NewGameService(gameDAO, dictItemDAO, authService, cfg)
	liveHub := livehub.NewHub()
	liveService := service.NewLiveService(siteSettingDAO, authService, liveHub)
	imageBedClient := imagebed.NewClient(cfg.ImageBedAPIURL, cfg.ImageBedToken)
	musicService := service.NewMusicService(musicTrackDAO, authService)
	galleryService := service.NewGalleryService(
		imageBedClient,
		cfg.ImageBedAlbumID,
		cfg.ImageBedOrder,
	)

	authController := controller.NewAuthController(authService)
	dictController := controller.NewDictController(dictService)
	postController := controller.NewPostController(postService)
	categoryController := controller.NewCategoryController(categoryService)
	tagController := controller.NewTagController(tagService)
	commentController := controller.NewCommentController(commentService)
	siteController := controller.NewSiteController(siteService)
	gameController := controller.NewGameController(gameService)
	liveController := controller.NewLiveController(liveService)
	liveWSController := controller.NewLiveWSController(liveService, liveHub)
	galleryController := controller.NewGalleryController(galleryService)
	musicController := controller.NewMusicController(musicService)

	web := engine.Group("/api/web")
	{
		web.GET("/site", siteController.WebGetSite)
		web.GET("/gallery/images", galleryController.WebListImages)
		web.GET("/posts", postController.WebList)
		web.GET("/archives", postController.WebArchives)
		web.GET("/categories", categoryController.WebList)
		web.GET("/tags", tagController.WebList)
		web.GET("/posts/:slug/comments", commentController.WebListByPostSlug)
		web.POST("/posts/:slug/comments", commentController.WebCreate)
		web.GET("/posts/:slug", postController.WebGetBySlug)
		web.GET("/games", gameController.WebList)
		web.GET("/games/genres", gameController.WebListGenres)
		web.GET("/dictionaries/:dictType/items", dictController.WebListItems)
		web.GET("/games/sidebar", gameController.WebSidebar)
		web.GET("/live", liveController.WebGetLive)
		web.GET("/live/ws", liveWSController.WebLiveWS)
		web.GET("/music/playlist", musicController.WebPlaylist)
	}

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

		admin.GET("/dictionaries", dictController.AdminListTypes)
		admin.POST("/dictionaries/:dictType/update", dictController.AdminUpdateType)
		admin.GET("/dictionaries/:dictType/items", dictController.AdminListItems)
		admin.POST("/dictionaries/:dictType/items", dictController.AdminCreateItem)
		admin.GET("/dict-items", dictController.AdminListItems)
		admin.POST("/dict-items", dictController.AdminCreateItem)
		admin.POST("/dict-items/:id/update", dictController.AdminUpdateItem)
		admin.POST("/dict-items/:id/delete", dictController.AdminDeleteItem)
		admin.GET("/operation-logs", authController.ListOperationLogs)

		admin.GET("/posts", postController.AdminList)
		admin.GET("/posts/:id/comments", commentController.AdminListByPostID)
		admin.GET("/posts/:id", postController.AdminGetByID)
		admin.POST("/posts", postController.Create)
		admin.POST("/posts/:id/update", postController.Update)
		admin.POST("/posts/:id/delete", postController.Delete)
		admin.POST("/posts/:id/status", postController.UpdateStatus)

		admin.GET("/categories", categoryController.AdminList)
		admin.GET("/categories/:id", categoryController.AdminGetByID)
		admin.POST("/categories", categoryController.Create)
		admin.POST("/categories/:id/update", categoryController.Update)
		admin.POST("/categories/:id/delete", categoryController.Delete)

		admin.GET("/tags", tagController.AdminList)
		admin.GET("/tags/:id", tagController.AdminGetByID)
		admin.POST("/tags", tagController.Create)
		admin.POST("/tags/:id/update", tagController.Update)
		admin.POST("/tags/:id/delete", tagController.Delete)

		admin.GET("/comments", commentController.AdminList)
		admin.POST("/comments/:id/approve", commentController.AdminApprove)
		admin.POST("/comments/:id/reject", commentController.AdminReject)
		admin.POST("/comments/:id/delete", commentController.AdminDelete)
		admin.POST("/comments/:id/reply", commentController.AdminReply)

		admin.GET("/games", gameController.AdminList)
		admin.GET("/games/:id", gameController.AdminGetByID)
		admin.POST("/games/sync", gameController.AdminSync)
		admin.POST("/games/:id/update", gameController.AdminUpdate)

		admin.GET("/live", liveController.AdminGetLive)
		admin.POST("/live/update", liveController.AdminUpdateLive)

		admin.GET("/site-settings", siteController.AdminGetSiteSettings)
		admin.POST("/site-settings/update", siteController.AdminUpdateSiteSettings)

		admin.GET("/music/tracks", musicController.AdminList)
		admin.GET("/music/tracks/:id", musicController.AdminGetByID)
		admin.POST("/music/tracks", musicController.Create)
		admin.POST("/music/tracks/:id/update", musicController.Update)
		admin.POST("/music/tracks/:id/delete", musicController.Delete)
	}

	return engine
}
