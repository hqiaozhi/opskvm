package kvm

import (
	"context"
	"fmt"
	"time"

	v1 "opskvm/api/kvm/v1"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

func (c *ControllerV1) CreateSession(ctx context.Context, req *v1.CreateSessionReq) (res *v1.CreateSessionRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	sessionId := guid.S([]byte(r.RemoteAddr), []byte(fmt.Sprintf("%v-%d", r.Header, time.Now().UnixNano())))

	sessionsLock.Lock()
	sessions[sessionId] = true
	sessionsLock.Unlock()

	return &v1.CreateSessionRes{
		SessionId: sessionId,
	}, nil
}
