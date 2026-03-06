package webshell

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	v1 "opskvm/api/webshell/v1"
	"opskvm/internal/logic/webshell"

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

type Terminal struct {
	ID   string
	Cols uint16
	Rows uint16
	Pt   *webshell.Pty
	mu   sync.RWMutex
}

var (
	terminals     = make(map[string]*Terminal)
	terminalsLock sync.RWMutex
)

func (c *ControllerV1) Shell(ctx context.Context, req *v1.ShellReq) (res *v1.ShellRes, err error) {
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

	term := &Terminal{
		ID:   sessionId,
		Cols: 80,
		Rows: 24,
		Pt:   webshell.NewPty(),
	}
	term.Pt.Start()

	terminalsLock.Lock()
	terminals[sessionId] = term
	terminalsLock.Unlock()

	conn, err := upgrader.Upgrade(gReq.Response.Writer, gReq.Request, nil)
	if err != nil {
		terminalsLock.Lock()
		delete(terminals, sessionId)
		terminalsLock.Unlock()
		term.Pt.Close()
		gReq.Response.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	defer conn.Close()

	defer func() {
		terminalsLock.Lock()
		delete(terminals, sessionId)
		terminalsLock.Unlock()
		term.Pt.Close()
	}()

	go func() {
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if msgType == websocket.BinaryMessage {
				if _, err := term.Pt.Write(message); err != nil {
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-term.Pt.ReadCh():
			if !ok {
				return
			}
			if err = conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}
