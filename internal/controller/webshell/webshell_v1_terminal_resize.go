package webshell

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/webshell/v1"
)

func (c *ControllerV1) TerminalResize(ctx context.Context, req *v1.TerminalResizeReq) (res *v1.TerminalResizeRes, err error) {
	sessionId := req.SessionId
	if sessionId == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidRequest, "sessionId is required")
	}

	terminalsLock.RLock()
	term, ok := terminals[sessionId]
	terminalsLock.RUnlock()

	if !ok {
		return nil, gerror.NewCode(gcode.CodeNotFound, "terminal not found")
	}

	term.mu.Lock()
	term.Cols = req.Cols
	term.Rows = req.Rows
	term.mu.Unlock()

	term.Pt.Resize(req.Cols, req.Rows)

	return nil, nil
}
