package video

import (
	"fmt"
	"net/http"
	"time"

	v1 "opskvm/api/video/v1"
	"opskvm/internal/core/video"

	"github.com/gin-gonic/gin"

	"github.com/korandiz/v4l/fmt/mjpeg"
)

// GetConfigHandler 处理GET /api/config（查询配置）
func (s *VideoHandler) GetConfigHandler(c *gin.Context) {
	cfg, err := s.svcCtx.Camera.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get config: %v", err),
		})
		return
	}

	// 构建响应
	resp := v1.ConfigResponse{
		Success: true,
		Config: &video.Config{
			Width:  cfg.Width,
			Height: cfg.Height,
			FPS:    float64(cfg.FPS.N) / float64(cfg.FPS.D),
		},
	}
	c.JSON(http.StatusOK, resp)
}

// GetSupportedConfigsHandler 处理GET /api/configs（查询支持的配置列表）
func (s *VideoHandler) GetSupportedConfigsHandler(c *gin.Context) {
	// 获取设备支持的配置列表
	supportedCfgs, err := s.svcCtx.Camera.ListConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get supported configs: %v", err),
		})
		return
	}

	// 过滤并转换为简化的Config格式（只返回MJPEG格式的配置）
	var configs []*video.Config
	for _, cfg := range supportedCfgs {
		if cfg.Format == mjpeg.FourCC {
			configs = append(configs, &video.Config{
				Width:  cfg.Width,
				Height: cfg.Height,
				FPS:    float64(cfg.FPS.N) / float64(cfg.FPS.D),
			})
		}
	}

	// 构建响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"configs": configs,
	})
}

// UpdateConfigHandler 处理POST /api/config（修改配置）
func (s *VideoHandler) UpdateConfigHandler(c *gin.Context) {
	// 绑定并校验请求体（Gin自动校验binding标签）
	var req v1.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// 暂停流分发，避免配置更新冲突（延长暂停时间，适配设备重启）
	s.svcCtx.Streamer.Pause()
	defer func() {
		// 恢复流分发前短暂延迟，确保设备完全启动
		time.Sleep(200 * time.Millisecond)
		s.svcCtx.Streamer.Resume()
	}()

	// 更新摄像头配置
	actualCfg, err := s.svcCtx.Camera.UpdateConfig(req.Width, req.Height, uint32(req.FPS))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update config: %v", err),
		})
		return
	}

	// 构建成功响应
	c.JSON(http.StatusOK, v1.ConfigResponse{
		Success: true,
		Message: "Config updated successfully (device restarted)",
		Config: &video.Config{
			Width:  actualCfg.Width,
			Height: actualCfg.Height,
			FPS:    float64(actualCfg.FPS.N) / float64(actualCfg.FPS.D),
		},
	})
}
