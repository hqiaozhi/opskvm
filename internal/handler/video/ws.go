package video

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"opskvm/internal/core/video"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSMessageType WebSocket消息类型
const (
	WSMessageTypeVideo     WSMessageType = 0 // 视频流数据
	WSMessageTypeKeyboard  WSMessageType = 1 // 键盘事件
	WSMessageTypeMouse     WSMessageType = 2 // 鼠标事件
	WSMessageTypeMouseMode WSMessageType = 3 // 鼠标模式切换
)

// WSMessageType WebSocket消息类型
type WSMessageType int

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
	Buttons  byte `json:"buttons"`  // 鼠标按键状态
	DX       int  `json:"dx"`       // X方向增量
	DY       int  `json:"dy"`       // Y方向增量
	Wheel    int8 `json:"wheel"`    // 滚轮增量
	Absolute bool `json:"absolute"` // 是否为绝对坐标
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
		log.Printf("[%s] Failed to add client to streamer (streamer may be stopped)", c.ClientIP())
		// 记录当前流状态
		log.Printf("[%s] Current streamer status: stopped=%v", c.ClientIP(), s.svcCtx.Streamer.IsStopped())
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
	// 记录原始消息，便于调试
	log.Printf("Received WebSocket raw message: %s", string(message))

	// 解析JSON消息
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Parse WebSocket message error: %v", err)
		return
	}

	// 记录消息类型
	log.Printf("Received WebSocket message, type: %d", wsMsg.Type)

	// 根据消息类型处理
	switch wsMsg.Type {
	case WSMessageTypeKeyboard:
		// 处理键盘事件
		var keyboardData KeyboardData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				log.Printf("Failed to marshal keyboard data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &keyboardData); err != nil {
				log.Printf("Failed to unmarshal keyboard data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				log.Printf("Failed to marshal keyboard data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &keyboardData); err != nil {
				log.Printf("Failed to unmarshal keyboard data: %v", err)
				return
			}
		}

		// 记录键盘事件
		log.Printf("Keyboard event: modifier=0x%02x, keys=%v", keyboardData.Modifier, keyboardData.Keys)

		// 直接使用接收到的按键码发送报告
		s.svcCtx.KMHID.SendKeyboardReport(keyboardData.Modifier, keyboardData.Keys)

	case WSMessageTypeMouse:
		// 处理鼠标事件
		var mouseData MouseData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				log.Printf("Failed to marshal mouse data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &mouseData); err != nil {
				log.Printf("Failed to unmarshal mouse data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				log.Printf("Failed to marshal mouse data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &mouseData); err != nil {
				log.Printf("Failed to unmarshal mouse data: %v", err)
				return
			}
		}

		// 记录当前鼠标模式
		absolute := s.svcCtx.KMHID.IsAbsoluteMouse()
		log.Printf("WS Mouse Data - Type: %d, Mode: %v, Data: %+v", wsMsg.Type, absolute, mouseData)

		// 记录所有鼠标事件，包括纯移动事件
		log.Printf("Mouse event: buttons=0x%02x, dx=%d, dy=%d, wheel=%d", mouseData.Buttons, mouseData.DX, mouseData.DY, mouseData.Wheel)
		s.svcCtx.KMHID.SendMouseReport(mouseData.Buttons, int(mouseData.DX), int(mouseData.DY), mouseData.Wheel)

	case WSMessageTypeMouseMode:
		// 处理鼠标模式切换
		var mouseModeData MouseModeData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				log.Printf("Failed to marshal mouse mode data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &mouseModeData); err != nil {
				log.Printf("Failed to unmarshal mouse mode data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				log.Printf("Failed to marshal mouse mode data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &mouseModeData); err != nil {
				log.Printf("Failed to unmarshal mouse mode data: %v", err)
				return
			}
		}

		log.Printf("Mouse mode change: absolute=%v", mouseModeData.Absolute)
		s.svcCtx.KMHID.SetAbsoluteMouse(mouseModeData.Absolute)
	case WSMessageTypeVideo:
		// 视频流数据，这里不需要处理，因为视频流是从服务器发送到客户端的
		log.Printf("Received video stream data, ignoring")
	default:
		log.Printf("Unknown WebSocket message type: %d", wsMsg.Type)
	}
}
