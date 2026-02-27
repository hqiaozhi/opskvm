package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"users/login" method:"post" sm:"登录" tags:"用户"`
	Username string `json:"username" v:"required|length:3,12" dc:"账户"`
	Password string `json:"password" v:"required|length:6,16" dc:"密码"`
}

type LoginRes struct {
	Token string `json:"token" dc:"在需要鉴权的接口中header加入Authorization: token"`
}

type LogoutReq struct {
	g.Meta `path:"users/logout" method:"post" sm:"退出" tags:"用户"`
}

type LogoutRes struct {
}
