package video

import (
	"fmt"
	"log"
	"net/http"
	"time"

	v1 "opskvm/api/video/v1"
	"opskvm/internal/core/video"
	"opskvm/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/korandiz/v4l/fmt/mjpeg"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	svcCtx *svc.SvcContext
}

// NewVideoHandler 创建视频处理器实例
func NewVideoHandler(svcCtx *svc.SvcContext) *VideoHandler {
	return &VideoHandler{
		svcCtx: svcCtx,
	}
}

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

// ServeStream WebSocket流处理
func (s *VideoHandler) ServeStream(c *gin.Context) {
	// 自动打开摄像头
	err := s.svcCtx.Camera.TurnOn()
	if err != nil {
		log.Printf("[%s] TurnOn error: %v", c.ClientIP(), err)
	}

	log.Printf("[%s] New websocket stream connection", c.ClientIP())

	// 升级HTTP连接为WebSocket连接
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源的WebSocket连接
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[%s] WebSocket upgrade error: %v", c.ClientIP(), err)
		return
	}
	defer conn.Close()

	clt := s.svcCtx.Streamer.AddClient()
	if clt == nil {
		return
	}
	defer s.svcCtx.Streamer.RemoveClient(clt)

	// 持续发送帧数据
	for {
		buf, ok := <-clt.Ch
		if !ok {
			log.Printf("[%s] Connection closed", c.ClientIP())
			return
		}

		conn.SetWriteDeadline(time.Now().Add(time.Second))

		// 发送帧数据
		if buf == nil {
			if err := conn.WriteMessage(websocket.BinaryMessage, s.svcCtx.Streamer.(*video.MJPEGStreamer).Blank); err != nil {
				log.Printf("[%s] Write blank frame error: %v", c.ClientIP(), err)
			}
			log.Printf("[%s] Quitting", c.ClientIP())
			return
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
			log.Printf("[%s] Write frame data error: %v", c.ClientIP(), err)
			return
		}
	}
}
