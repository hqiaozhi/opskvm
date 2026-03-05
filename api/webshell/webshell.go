// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package webshell

import (
	"context"

	"opskvm/api/webshell/v1"
)

type IWebshellV1 interface {
	Shell(ctx context.Context, req *v1.ShellReq) (res *v1.ShellRes, err error)
	CreateShell(ctx context.Context, req *v1.CreateShellReq) (res *v1.CreateShellRes, err error)
	TerminalResize(ctx context.Context, req *v1.TerminalResizeReq) (res *v1.TerminalResizeRes, err error)
	TerminalInfo(ctx context.Context, req *v1.TerminalInfoReq) (res *v1.TerminalInfoRes, err error)
}
