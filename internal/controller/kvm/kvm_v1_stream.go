package kvm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	v1 "opskvm/api/kvm/v1"
	"opskvm/internal/service/video"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WSMessageType WebSocket消息类型
const (
	WSMessageTypeVideo        WSMessageType = 0 // 视频流数据
	WSMessageTypeKeyboard     WSMessageType = 1 // 键盘事件
	WSMessageTypeMouse        WSMessageType = 2 // 鼠标事件
	WSMessageTypeMouseMode    WSMessageType = 3 // 鼠标模式切换
	WSMessageTypeVideoControl WSMessageType = 4 // 视频控制
)

// WSMessageType WebSocket消息类型
type WSMessageType int

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type interface{} `json:"type"` // 消息类型，可以是数字或字符串
	Data interface{} `json:"data"` // 消息数据
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

// VideoControlData 视频控制数据
type VideoControlData struct {
	Action string `json:"action"` // 操作: stop/start
}

func (c *ControllerV1) Stream(ctx context.Context, req *v1.StreamReq) (res *v1.StreamRes, err error) {
	gReq := ghttp.RequestFromCtx(ctx)

	sessionId := req.SessionId
	if sessionId == "" {
		gReq.Response.WriteHeader(http.StatusBadRequest)
		g.Log().Errorf(ctx, "sessionid is required")
		return nil, fmt.Errorf("sessionid is required")
	}

	sessionsLock.RLock()
	_, exists := sessions[sessionId]
	sessionsLock.RUnlock()

	if !exists {
		gReq.Response.WriteHeader(http.StatusForbidden)
		return nil, fmt.Errorf("invalid sessionid")
	}

	sessionsLock.Lock()
	delete(sessions, sessionId)
	sessionsLock.Unlock()

	// 连接的时候打开摄像头
	err = c.kvm.SVC.Camera.TurnOn()
	if err != nil {
		g.Log().Warningf(ctx, "TurnOn Device: %s", err)
	}

	conn, err := upgrader.Upgrade(gReq.Response.Writer, gReq.Request, nil)
	if err != nil {
		gReq.Response.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	defer conn.Close()

	clt := c.kvm.SVC.Streamer.AddClient()
	if clt == nil {
		g.Log().Debugf(ctx, "[%s] Failed to add client to streamer (streamer may be stopped)", gReq.GetClientIp())
		// 记录当前流状态
		g.Log().Debugf(ctx, "[%s] Current streamer status: stopped=%v", gReq.GetClientIp(), c.kvm.SVC.Streamer.IsStopped())
		return
	}
	defer c.kvm.SVC.Streamer.RemoveClient(clt)

	// 启动消息接收协程
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				g.Log().Warningf(ctx, "[%s] Read message error: %v", gReq.GetClientIp(), err)
				return
			}

			// 处理接收到的消息
			c.handleWSMessage(ctx, clt, message)
		}
	}()

	// 持续发送帧数据
	for {
		buf, ok := <-clt.Ch
		if !ok {
			g.Log().Debugf(ctx, "[%s] Connection closed", gReq.GetClientIp())
			return
		}

		// 检查客户端是否需要接收视频
		if !clt.SendVideo {
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(time.Second))

		// 发送帧数据
		if buf == nil {
			if err := conn.WriteMessage(websocket.BinaryMessage, c.kvm.SVC.Streamer.(*video.MJPEGStreamer).Blank); err != nil {
				g.Log().Printf(ctx, "[%s] Write blank frame error: %v", gReq.GetClientIp(), err)
			}
			g.Log().Printf(ctx, "[%s] Quitting", gReq.GetClientIp())
			return
		}

		if err = conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
			g.Log().Printf(ctx, "[%s] Write frame data error: %v", gReq.GetClientIp(), err)
			return
		}
	}
}

// handleWSMessage 处理WebSocket消息
func (c *ControllerV1) handleWSMessage(ctx context.Context, clt *video.Client, message []byte) {
	// 记录原始消息，便于调试
	g.Log().Printf(ctx, "Received WebSocket raw message: %s", string(message))

	// 解析JSON消息
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		g.Log().Printf(ctx, "Parse WebSocket message error: %v", err)
		return
	}

	// 统一处理消息类型（支持数字或字符串）
	var msgType WSMessageType
	switch v := wsMsg.Type.(type) {
	case float64:
		msgType = WSMessageType(int(v))
	case string:
		switch v {
		case "video":
			msgType = WSMessageTypeVideo
		case "keyboard":
			msgType = WSMessageTypeKeyboard
		case "mouse":
			msgType = WSMessageTypeMouse
		case "mouse_mode":
			msgType = WSMessageTypeMouseMode
		case "video_control":
			msgType = WSMessageTypeVideoControl
		default:
			g.Log().Printf(ctx, "Unknown WebSocket message type string: %s", v)
			return
		}
	default:
		g.Log().Printf(ctx, "Unknown WebSocket message type: %T", wsMsg.Type)
		return
	}

	// 记录消息类型
	g.Log().Printf(ctx, "Received WebSocket message, type: %d", msgType)

	// 根据消息类型处理
	switch msgType {
	case WSMessageTypeKeyboard:
		// 处理键盘事件
		var keyboardData KeyboardData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal keyboard data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &keyboardData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal keyboard data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal keyboard data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &keyboardData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal keyboard data: %v", err)
				return
			}
		}

		// 记录键盘事件
		g.Log().Printf(ctx, "Keyboard event: modifier=0x%02x, keys=%v", keyboardData.Modifier, keyboardData.Keys)

		// 直接使用接收到的按键码发送报告
		c.kvm.SVC.HID.SendKeyboardReport(keyboardData.Modifier, keyboardData.Keys)

	case WSMessageTypeMouse:
		// 处理鼠标事件
		var mouseData MouseData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal mouse data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &mouseData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal mouse data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal mouse data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &mouseData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal mouse data: %v", err)
				return
			}
		}

		// 记录当前鼠标模式
		absolute := c.kvm.SVC.HID.IsAbsoluteMouse()
		g.Log().Printf(ctx, "WS Mouse Data - Type: %d, Mode: %v, Data: %+v", wsMsg.Type, absolute, mouseData)

		// 记录所有鼠标事件，包括纯移动事件
		g.Log().Printf(ctx, "Mouse event: buttons=0x%02x, dx=%d, dy=%d, wheel=%d", mouseData.Buttons, mouseData.DX, mouseData.DY, mouseData.Wheel)
		c.kvm.SVC.HID.SendMouseReport(mouseData.Buttons, int(mouseData.DX), int(mouseData.DY), mouseData.Wheel)

	case WSMessageTypeMouseMode:
		// 处理鼠标模式切换
		var mouseModeData MouseModeData
		// 先尝试将数据转换为map，再转换为结构体
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			// 转换为JSON字节
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal mouse mode data: %v", err)
				return
			}
			// 解析为结构体
			if err := json.Unmarshal(dataBytes, &mouseModeData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal mouse mode data: %v", err)
				return
			}
		} else {
			// 直接解析
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal mouse mode data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &mouseModeData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal mouse mode data: %v", err)
				return
			}
		}

		g.Log().Printf(ctx, "Mouse mode change: absolute=%v", mouseModeData.Absolute)
		c.kvm.SVC.HID.SetAbsoluteMouse(mouseModeData.Absolute)
	case WSMessageTypeVideo:
		// 视频流数据，这里不需要处理，因为视频流是从服务器发送到客户端的
		g.Log().Printf(ctx, "Received video stream data, ignoring")
	case WSMessageTypeVideoControl:
		// 处理视频控制消息
		var videoControlData VideoControlData
		if dataMap, ok := wsMsg.Data.(map[string]interface{}); ok {
			dataBytes, err := json.Marshal(dataMap)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal video control data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &videoControlData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal video control data: %v", err)
				return
			}
		} else {
			dataBytes, err := json.Marshal(wsMsg.Data)
			if err != nil {
				g.Log().Printf(ctx, "Failed to marshal video control data: %v", err)
				return
			}
			if err := json.Unmarshal(dataBytes, &videoControlData); err != nil {
				g.Log().Printf(ctx, "Failed to unmarshal video control data: %v", err)
				return
			}
		}

		g.Log().Printf(ctx, "Video control action: %s", videoControlData.Action)
		switch videoControlData.Action {
		case "stop":
			clt.SendVideo = false
			g.Log().Printf(ctx, "Video streaming stopped for this client")
		case "start":
			clt.SendVideo = true
			g.Log().Printf(ctx, "Video streaming resumed for this client")
		default:
			g.Log().Printf(ctx, "Unknown video control action: %s", videoControlData.Action)
		}
	default:
		g.Log().Printf(ctx, "Unknown WebSocket message type: %d", wsMsg.Type)
	}
}
