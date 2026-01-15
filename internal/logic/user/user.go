package user

import (
	"opskvm/internal/svc"
	"opskvm/internal/utils/jwt"
)

type User struct {
	svcCtx *svc.SvcContext
	jwt    *jwt.JwtService
}

func New(svcCtx *svc.SvcContext) *User {
	return &User{
		svcCtx: svcCtx,
		jwt:    svcCtx.JWT,
	}
}
