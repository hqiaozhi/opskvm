package video

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	v1 "opskvm/api/video/v1"
	"opskvm/internal/logic/video"
	"opskvm/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	logic  *video.Logic
	svcCtx *svc.SvcContext
}

// NewVideoHandler 创建视频处理器实例
func NewVideoHandler(svcCtx *svc.SvcContext) *VideoHandler {
	return &VideoHandler{
		logic:  video.NewLogic(svcCtx),
		svcCtx: svcCtx,
	}
}

// WebSocket WebSocket视频流处理
func (h *VideoHandler) WebSocket(c *gin.Context) {
	log.Printf("New WebSocket connection from %s", c.ClientIP())

	// 升级HTTP连接为WebSocket连接
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
		WriteBufferSize: 1024 * 1024, // 1MB写入缓冲区
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	log.Println("New WebSocket connection established")

	// 设置连接参数
	conn.SetReadLimit(512)                                // 限制客户端消息大小
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // 读取超时
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		return nil
	})

	// 读取客户端请求参数
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Read WebSocket message error: %v", err)
		return
	}

	var req v1.VideoRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Printf("Parse WebSocket message error: %v", err)
		// 发送错误响应
		if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid request"}`)); err != nil {
			log.Printf("Write error response error: %v", err)
		}
		return
	}

	log.Printf("Received WebSocket request: %+v", req)

	// 创建上下文用于控制视频流
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动视频流
	clt, err := h.logic.StreamClient(ctx)
	if err != nil {
		log.Printf("Video stream error: %v", err)
		return
	}
	defer clt.Remove()
	log.Println("视频流已启动")
	// 发送视频帧
	for {
		buf := <-clt.CH
		if buf == nil {
			log.Printf("[%s] Received nil frame, trying to reconnect", c.ClientIP())
			// 尝试重新获取视频流
			cancel()
			newCtx, newCancel := context.WithCancel(context.Background())
			newClt, err := h.logic.StreamClient(newCtx)
			if err != nil {
				log.Printf("[%s] Reconnect error: %v", c.ClientIP(), err)
				newCancel() // 确保新的 cancel 函数被调用
				return
			}
			clt.Remove()
			clt = newClt
			ctx = newCtx
			cancel = newCancel
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
			log.Printf("[%s] Write frame error: %v", c.ClientIP(), err)
			return
		}
	}
}
