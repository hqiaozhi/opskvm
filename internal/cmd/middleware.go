package cmd

import (
	"strings"

	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/net/ghttp"
)

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func MiddlewareAuth(r *ghttp.Request) {
	// 登录接口不校验token
	if r.URL.Path == "/api/v1/users/login" {
		r.Middleware.Next()
		return
	}
	// WebSocket 连接：通过 HTTP API 先获取 sessionId（已验证 token），WebSocket 跳过 token 验证
	if r.Header.Get("Upgrade") == "websocket" {
		r.Middleware.Next()
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteStatus(401, "Missing authorization header")
		r.Exit()
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.WriteStatus(401, "Invalid authorization header format")
		r.Exit()
	}

	token := parts[1]
	jwtService := users.JwtInstance
	claims, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		r.Response.WriteStatus(401, "Invalid token: "+err.Error())
		r.Exit()
	}

	r.SetParam("userid", claims.UserID)
	r.SetParam("username", claims.Username)
	r.SetCtxVar("token", token)
	r.Middleware.Next()
}
