package video

import (
	"fmt"
	"net/http"

	v1 "opskvm/api/video/v1"

	"github.com/gin-gonic/gin"
)

// TurnOnHandler 处理POST /api/on（打开摄像头）
func (s *VideoHandler) TurnOnHandler(c *gin.Context) {
	// 打开摄像头
	err := s.svcCtx.Camera.TurnOn()
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to turn on camera: %v", err),
		})
		return
	}

	// 构建成功响应
	c.JSON(http.StatusOK, v1.ConfigResponse{
		Success: true,
		Message: "Camera turned on successfully",
	})
}

// TurnOffHandler 处理POST /api/off（关闭摄像头）
func (s *VideoHandler) TurnOffHandler(c *gin.Context) {
	// 关闭摄像头
	err := s.svcCtx.Camera.TurnOff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to turn off camera: %v", err),
		})
		return
	}

	// 构建成功响应
	c.JSON(http.StatusOK, v1.ConfigResponse{
		Success: true,
		Message: "Camera turned off successfully",
	})
}
