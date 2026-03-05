package webshell

import (
	"context"

	v1 "opskvm/api/webshell/v1"
)

func (c *ControllerV1) TerminalInfo(ctx context.Context, req *v1.TerminalInfoReq) (res *v1.TerminalInfoRes, err error) {
	sessionId := req.SessionId

	terminalsLock.RLock()
	term, ok := terminals[sessionId]
	terminalsLock.RUnlock()

	if !ok {
		return &v1.TerminalInfoRes{
			Cols: 80,
			Rows: 24,
		}, nil
	}

	term.mu.RLock()
	defer term.mu.RUnlock()

	return &v1.TerminalInfoRes{
		Cols: term.Cols,
		Rows: term.Rows,
	}, nil
}
