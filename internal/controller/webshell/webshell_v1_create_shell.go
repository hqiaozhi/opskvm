package webshell

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "opskvm/api/webshell/v1"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

var (
	sessions     = make(map[string]bool)
	sessionsLock sync.RWMutex
)

func (c *ControllerV1) CreateShell(ctx context.Context, req *v1.CreateShellReq) (res *v1.CreateShellRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	sessionId := guid.S([]byte(r.RemoteAddr), []byte(fmt.Sprintf("%v-%d", r.Header, time.Now().UnixNano())))

	sessionsLock.Lock()
	sessions[sessionId] = true
	sessionsLock.Unlock()

	return &v1.CreateShellRes{
		SessionId: sessionId,
	}, nil
}
