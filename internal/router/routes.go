package router

import (
	"net/http"
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

	// 视频路由
	videoHandler := video.NewVideoHandler(ctx)

	// WebSocket视频流
	engine.ginEngine.GET("/ws/video", videoHandler.WebSocket)

	// 版本：/api/v1
	v1 := engine.Group("/api/v1")
	{
		// 公共路由组
		_ = v1.Group("")
		{

		}

	}
}
