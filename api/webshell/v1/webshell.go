package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ShellReq WebSocket Shell请求
type ShellReq struct {
	g.Meta    `path:"shell" method:"get" sm:"WebSocket SSH" tags:"WebShell"`
	SessionId string `json:"ssid" dc:"Session ID，可选，如果不传则创建新的"`
}

// ShellRes WebSocket Shell响应
type ShellRes struct {
	SessionId string `json:"ssid" dc:"Session ID，用于后续请求"`
}

// CreateShellReq 创建Shell会话请求
type CreateShellReq struct {
	g.Meta `path:"shell/create" method:"post" sm:"创建Shell会话" tags:"WebShell"`
}

// CreateShellRes 创建Shell会话响应
type CreateShellRes struct {
	SessionId string `json:"ssid" dc:"Session ID，用于后续请求"`
}

// TerminalResizeReq 终端调整大小请求
type TerminalResizeReq struct {
	g.Meta    `path:"shell/resize" method:"post" sm:"终端调整大小" tags:"WebShell"`
	SessionId string `json:"ssid" dc:"Session ID" v:"required|regex:^[a-zA-Z0-9-_]+$"`
	Cols      uint16 `json:"cols" dc:"终端列数" v:"required|min:1|max:500"`
	Rows      uint16 `json:"rows" dc:"终端行数" v:"required|min:1|max:500"`
}

// TerminalResizeRes 终端调整大小响应
type TerminalResizeRes struct {
}

// TerminalInfoReq 获取终端信息请求
type TerminalInfoReq struct {
	g.Meta    `path:"shell/info" method:"get" sm:"终端信息" tags:"WebShell"`
	SessionId string `json:"ssid" dc:"Session ID" v:"required|regex:^[a-zA-Z0-9-_]+$"`
}

// TerminalInfoRes 获取终端信息响应
type TerminalInfoRes struct {
	Cols uint16 `json:"cols" dc:"终端列数"`
	Rows uint16 `json:"rows" dc:"终端行数"`
}
