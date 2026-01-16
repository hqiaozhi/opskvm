package router

import (
	"net/http"
	"opskvm/internal/handler/user"
	"opskvm/internal/handler/video"
	"opskvm/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterBusinessRoutes(engine *GinEngine) {
	engine.Use(middleware.Cors())
	ctx := engine.svcCtx

	// 健康检查路由
	engine.ginEngine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 认证中间件
	engine.Use(middleware.Auth(ctx))

	// 版本：/api/v1
	v1 := engine.Group("/api/v1")
	{
		// 用户路由组
		userHandler := user.New(ctx)
		userGroup := v1.Group("/user")
		{
			userGroup.POST("/login", userHandler.Login)
			userGroup.POST("/logout", userHandler.Logout)
		}

		// 视频路由
		videoHandler := video.NewVideoHandler(ctx)
		videoGroup := v1.Group("/video")
		{
			videoGroup.GET("", videoHandler.ServeStream)
			videoGroup.GET("/config", videoHandler.GetConfigHandler)
			videoGroup.POST("/config", videoHandler.UpdateConfigHandler)
			videoGroup.GET("/configs", videoHandler.GetSupportedConfigsHandler)
			videoGroup.POST("/on", videoHandler.TurnOnHandler)
			videoGroup.POST("/off", videoHandler.TurnOffHandler)
			videoGroup.GET("/compress", videoHandler.GetCompressHandler)
			videoGroup.POST("/compress", videoHandler.CompressHandler)
		}

	}
}
