package video

import (
	"encoding/json"
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

// WSMessageType WebSocket消息类型
type WSMessageType int

const (
	WSMessageTypeVideo     WSMessageType = 0 // 视频流数据
	WSMessageTypeKeyboard  WSMessageType = 1 // 键盘事件
	WSMessageTypeMouse     WSMessageType = 2 // 鼠标事件
	WSMessageTypeMouseMode WSMessageType = 3 // 鼠标模式切换
)

// ASCIIToHIDMap ASCII码到HID扫描码的映射表
var ASCIIToHIDMap = map[byte]byte{
	// 字母
	'a': 0x04, 'b': 0x05, 'c': 0x06, 'd': 0x07, 'e': 0x08, 'f': 0x09, 'g': 0x0A,
	'h': 0x0B, 'i': 0x0C, 'j': 0x0D, 'k': 0x0E, 'l': 0x0F, 'm': 0x10, 'n': 0x11,
	'o': 0x12, 'p': 0x13, 'q': 0x14, 'r': 0x15, 's': 0x16, 't': 0x17, 'u': 0x18,
	'v': 0x19, 'w': 0x1A, 'x': 0x1B, 'y': 0x1C, 'z': 0x1D,
	'A': 0x04, 'B': 0x05, 'C': 0x06, 'D': 0x07, 'E': 0x08, 'F': 0x09, 'G': 0x0A,
	'H': 0x0B, 'I': 0x0C, 'J': 0x0D, 'K': 0x0E, 'L': 0x0F, 'M': 0x10, 'N': 0x11,
	'O': 0x12, 'P': 0x13, 'Q': 0x14, 'R': 0x15, 'S': 0x16, 'T': 0x17, 'U': 0x18,
	'V': 0x19, 'W': 0x1A, 'X': 0x1B, 'Y': 0x1C, 'Z': 0x1D,
	// 数字
	'1': 0x1E, '2': 0x1F, '3': 0x20, '4': 0x21, '5': 0x22, '6': 0x23, '7': 0x24,
	'8': 0x25, '9': 0x26, '0': 0x27,
	// 特殊字符
	'-': 0x2D, '=': 0x2E, '[': 0x2F, ']': 0x30, '\\': 0x31, ';': 0x33, '\'': 0x34,
	'`': 0x35, ',': 0x36, '.': 0x37, '/': 0x38,
	// 功能键
	' ': 0x2C,
}

// KeyCodeToHIDMap 前端按键码到HID扫描码的映射表（用于直接按键码映射）
var KeyCodeToHIDMap = map[byte]byte{
	// HID扫描码直接映射
	0x04: 0x04, 0x05: 0x05, 0x06: 0x06, 0x07: 0x07, 0x08: 0x08, 0x09: 0x09, 0x0A: 0x0A,
	0x0B: 0x0B, 0x0C: 0x0C, 0x0D: 0x0D, 0x0E: 0x0E, 0x0F: 0x0F, 0x10: 0x10, 0x11: 0x11,
	0x12: 0x12, 0x13: 0x13, 0x14: 0x14, 0x15: 0x15, 0x16: 0x16, 0x17: 0x17, 0x18: 0x18,
	0x19: 0x19, 0x1A: 0x1A, 0x1B: 0x1B, 0x1C: 0x1C, 0x1D: 0x1D,
}

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type WSMessageType `json:"type"` // 消息类型
	Data interface{}   `json:"data"` // 消息数据
}

// KeyboardData 键盘事件数据
type KeyboardData struct {
	Modifier byte   `json:"modifier"` // 修饰键
	Keys     []byte `json:"keys"`     // 按键列表
}

// MouseData 鼠标事件数据
type MouseData struct {
	Buttons byte `json:"buttons"` // 鼠标按键状态
	DX      int8 `json:"dx"`      // X方向增量
	DY      int8 `json:"dy"`      // Y方向增量
	Wheel   int8 `json:"wheel"`   // 滚轮增量
}

// MouseModeData 鼠标模式切换数据
type MouseModeData struct {
	Absolute bool `json:"absolute"` // 是否为绝对模式
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

	// 启动消息接收协程
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[%s] Read message error: %v", c.ClientIP(), err)
				return
			}

			// 处理接收到的消息
			s.handleWSMessage(message)
		}
	}()

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

// handleWSMessage 处理WebSocket消息
func (s *VideoHandler) handleWSMessage(message []byte) {
	// 解析JSON消息
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Parse WebSocket message error: %v", err)
		return
	}

	// 只在关键消息类型（如模式切换）时记录完整日志，减少重复日志
	if wsMsg.Type == WSMessageTypeMouseMode {
		log.Printf("Received WebSocket message, type: %d", wsMsg.Type)
	}

	// 根据消息类型处理
	switch wsMsg.Type {
	case WSMessageTypeKeyboard:
		// 处理键盘事件
		data, ok := wsMsg.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid keyboard message data format")
			return
		}

		// 安全获取modifier字段，默认值为0
		modifier := byte(0)
		if modifierVal, exists := data["modifier"]; exists && modifierVal != nil {
			if val, ok := modifierVal.(float64); ok {
				modifier = byte(val)
			}
		}

		// 安全获取keys字段，默认值为空切片
		keys := make([]byte, 0)
		if keysData, exists := data["keys"]; exists && keysData != nil {
			if keysSlice, ok := keysData.([]interface{}); ok {
				for _, k := range keysSlice {
					if k != nil {
						if keyVal, ok := k.(float64); ok {
							keys = append(keys, byte(keyVal))
						}
					}
				}
			}
		}

		// 只在有按键变化时记录日志
		if len(keys) > 0 || modifier > 0 {
			log.Printf("Keyboard event: modifier=0x%02x, keys=%v", modifier, keys)
		}

		// 直接使用接收到的按键码发送报告
		s.svcCtx.KMHID.SendKeyboardReport(modifier, keys)

	case WSMessageTypeMouse:
		// 处理鼠标事件
		data, ok := wsMsg.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid mouse message data format")
			return
		}

		// 安全获取buttons字段，默认值为0
		buttons := byte(0)
		if buttonsVal, exists := data["buttons"]; exists && buttonsVal != nil {
			if val, ok := buttonsVal.(float64); ok {
				buttons = byte(val)
			}
		}

		// 安全获取坐标字段
		dx := 0
		dy := 0

		// 直接从dx和dy字段获取坐标
		// 根据用户提示：绝对坐标和相对坐标使用的API都是一样的，通过另外的接口控制模式的切换
		if dxVal, exists := data["dx"]; exists && dxVal != nil {
			if val, ok := dxVal.(float64); ok {
				dx = int(val)
			}
		}
		if dyVal, exists := data["dy"]; exists && dyVal != nil {
			if val, ok := dyVal.(float64); ok {
				dy = int(val)
			}
		}

		// 安全获取wheel字段，默认值为0
		wheel := int8(0)
		if wheelVal, exists := data["wheel"]; exists && wheelVal != nil {
			if val, ok := wheelVal.(float64); ok {
				wheel = int8(val)
			}
		}

		// 记录当前鼠标模式和原始事件数据，用于诊断绝对模式问题
		absolute := s.svcCtx.KMHID.IsAbsoluteMouse()
		log.Printf("WS Mouse Data - Type: %d, Mode: %v, Raw: %+v", wsMsg.Type, absolute, data)

		// 记录所有鼠标事件，包括纯移动事件
		log.Printf("Mouse event: buttons=0x%02x, dx=%d, dy=%d, wheel=%d", buttons, dx, dy, wheel)
		s.svcCtx.KMHID.SendMouseReport(buttons, dx, dy, wheel)

	case WSMessageTypeMouseMode:
		// 处理鼠标模式切换
		data, ok := wsMsg.Data.(map[string]interface{})
		if !ok {
			log.Printf("Invalid mouse mode message data format")
			return
		}

		// 安全获取absolute字段，默认值为false
		absolute := false
		if absoluteVal, exists := data["absolute"]; exists && absoluteVal != nil {
			if val, ok := absoluteVal.(bool); ok {
				absolute = val
			}
		}

		log.Printf("Mouse mode change: absolute=%v", absolute)
		s.svcCtx.KMHID.SetAbsoluteMouse(absolute)
	default:
		log.Printf("Unknown WebSocket message type: %d", wsMsg.Type)
	}
}
