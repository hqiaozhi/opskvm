package cmd

import (
	"strings"

	"opskvm/internal/logic/totp"
	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/net/ghttp"
)

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func MiddlewareAuth(r *ghttp.Request) {
	noAuthPaths := []string{
		"/api/v1/users/login",
		"/api/v1/users/register",
	}
	isNoAuth := false
	for _, path := range noAuthPaths {
		if r.URL.Path == path {
			isNoAuth = true
			break
		}
	}
	if isNoAuth {
		r.Middleware.Next()
		return
	}

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

	userId := claims.UserID
	r.SetParam("userid", userId)
	r.SetParam("username", claims.Username)
	r.SetCtxVar("token", token)

	userIdInt := jwtService.GetUserIdInt(claims)
	needTotp, err := totp.TotpInstance.NeedTotp(r.Context(), userIdInt)
	if err == nil && needTotp {
		if !totp.TotpInstance.IsVerified(r.Context(), userIdInt) {
			r.Response.WriteStatus(403, "TOTP verification required")
			r.Exit()
		}
	}

	r.Middleware.Next()
}
