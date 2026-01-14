package video

import (
	"fmt"
	"log"
	"net/http"
	"time"

	v1 "opskvm/api/video/v1"
	"opskvm/internal/core/video"
	"opskvm/internal/svc"

	"github.com/korandiz/v4l/fmt/mjpeg"

	"github.com/gin-gonic/gin"
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

// ServeStream MJPEG流处理（兼容原生Hijack）
func (s *VideoHandler) ServeStream(c *gin.Context) {
	log.Printf("[%s] New stream connection", c.ClientIP())

	// 获取原生ResponseWriter
	w := c.Writer

	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("[%s] ResponseWriter is not a Hijacker", c.ClientIP())
		c.AbortWithStatusJSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		log.Printf("[%s] Hijack error: %v", c.ClientIP(), err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, v1.ConfigResponse{
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}
	defer conn.Close()

	clt := s.svcCtx.Streamer.AddClient()
	if clt == nil {
		return
	}
	defer s.svcCtx.Streamer.RemoveClient(clt)

	const boundary = "45c7pIy0cxa4vWtwGuVuAkbzKAQGpRjz9eyhyHTv"

	// 发送响应头
	respHeader := []byte(
		"HTTP/1.1 200 OK\r\n" +
			"Date: " + time.Now().UTC().Format(http.TimeFormat) + "\r\n" +
			"Content-Type: multipart/x-mixed-replace; boundary=" + boundary + "\r\n" +
			"Cache-Control: no-cache, no-store, max-age=0, must-revalidate\r\n" +
			"Pragma: no-cache\r\n" +
			"\r\n" +
			"--" + boundary + "\r\n",
	)
	if _, err := conn.Write(respHeader); err != nil {
		log.Printf("[%s] Write header error: %v", c.ClientIP(), err)
		return
	}

	// 持续发送帧数据
	for {
		buf, ok := <-clt.Ch
		if !ok {
			log.Printf("[%s] Connection closed", c.ClientIP())
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second))

		// 发送帧头
		if _, err := conn.Write([]byte("Content-Type: image/jpeg\r\n\r\n")); err != nil {
			log.Printf("[%s] Write frame header error: %v", c.ClientIP(), err)
			return
		}

		// 发送帧数据
		if buf == nil {
			if _, err := conn.Write(s.svcCtx.Streamer.(*video.MJPEGStreamer).Blank); err == nil {
				conn.Write([]byte("--" + boundary + "--\r\n"))
			} else {
				log.Printf("[%s] Write blank frame error: %v", c.ClientIP(), err)
			}
			log.Printf("[%s] Quitting", c.ClientIP())
			return
		}

		if _, err := conn.Write(buf); err != nil {
			log.Printf("[%s] Write frame data error: %v", c.ClientIP(), err)
			return
		}

		// 发送边界
		if _, err := conn.Write([]byte("--" + boundary + "\r\n")); err != nil {
			log.Printf("[%s] Write boundary error: %v", c.ClientIP(), err)
			return
		}
	}
}
