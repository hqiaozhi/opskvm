package webshell

import (
	"context"
	"fmt"

	v1 "opskvm/api/webshell/v1"
	webshelllogic "opskvm/internal/logic/webshell"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

func (c *ControllerV1) CreateShell(ctx context.Context, req *v1.CreateShellReq) (res *v1.CreateShellRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	sessionId := guid.S([]byte(r.RemoteAddr), []byte(fmt.Sprintf("%v", r.Header)))

	terminalsLock.RLock()
	existingTerm, exists := terminals[sessionId]
	terminalsLock.RUnlock()

	if exists {
		terminalsLock.Lock()
		delete(terminals, sessionId)
		terminalsLock.Unlock()
		existingTerm.Pt.Close()
	}

	term := &Terminal{
		ID:   sessionId,
		Cols: 80,
		Rows: 24,
		Pt:   webshelllogic.NewPty(),
	}
	term.Pt.Start()

	terminalsLock.Lock()
	terminals[sessionId] = term
	terminalsLock.Unlock()

	return &v1.CreateShellRes{
		SessionId: sessionId,
	}, nil
}
