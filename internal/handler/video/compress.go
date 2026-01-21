package video

import (
	"fmt"
	"net/http"

	v1 "opskvm/api/video/v1"
	"opskvm/internal/core/video"

	"github.com/gin-gonic/gin"
)

// GetCompressHandler 获取当前压缩状态
func (s *VideoHandler) GetCompressHandler(c *gin.Context) {
	// 获取当前流分发器实例
	streamer, ok := s.svcCtx.Streamer.(*video.MJPEGStreamer)
	if !ok {
		c.JSON(http.StatusInternalServerError, v1.CompressResponse{
			Success: false,
			Error:   "Invalid streamer type",
		})
		return
	}

	// 获取当前压缩配置
	enabled, quality := streamer.GetCompressConfig()

	// 构建成功响应
	c.JSON(http.StatusOK, v1.CompressResponse{
		Success: true,
		Enabled: enabled,
		Quality: quality,
	})
}

// CompressHandler 处理压缩控制请求
func (s *VideoHandler) CompressHandler(c *gin.Context) {
	// 绑定并校验请求体
	var req v1.CompressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, v1.CompressResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// 获取当前流分发器实例
	streamer, ok := s.svcCtx.Streamer.(*video.MJPEGStreamer)
	if !ok {
		c.JSON(http.StatusInternalServerError, v1.CompressResponse{
			Success: false,
			Error:   "Invalid streamer type",
		})
		return
	}

	// 根据请求启用或禁用压缩
	if req.Enabled {
		streamer.EnableCompression()
		// 如果请求中包含质量参数，则设置压缩质量
		if req.Quality > 0 {
			streamer.SetCompressQuality(req.Quality)
		}
	} else {
		streamer.DisableCompression()
	}

	// 获取当前压缩配置
	enabled, quality := streamer.GetCompressConfig()

	// 构建成功响应
	c.JSON(http.StatusOK, v1.CompressResponse{
		Success: true,
		Message: "Compress config updated successfully",
		Enabled: enabled,
		Quality: quality,
	})
}
