package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"users/login" method:"post" sm:"登录" tags:"用户管理"`
	Username string `json:"username" v:"required|length:3,12" dc:"账户"`
	Password string `json:"password" v:"required|length:6,16" dc:"密码"`
}

type LoginRes struct {
	Token string `json:"token" dc:"在需要鉴权的接口中header加入Authorization: token"`
}

type LogoutReq struct {
	g.Meta `path:"users/logout" method:"post" sm:"退出" tags:"用户管理"`
}

type LogoutRes struct {
}

type RegisterReq struct {
	g.Meta   `path:"users/register" method:"post" sm:"注册" tags:"用户管理"`
	Username string `json:"username" v:"required|length:3,12" dc:"账户"`
	Password string `json:"password" v:"required|length:6,16" dc:"密码"`
	Nickname string `json:"nickname" v:"length:0,50" dc:"昵称"`
	Email    string `json:"email" v:"email|length:0,100" dc:"邮箱"`
}

type RegisterRes struct {
	Id int `json:"id" dc:"用户ID"`
}

type ChangePasswordReq struct {
	g.Meta      `path:"users/changePassword" method:"post" sm:"修改密码" tags:"用户管理"`
	OldPassword string `json:"oldPassword" v:"required|length:6,16" dc:"旧密码"`
	NewPassword string `json:"newPassword" v:"required|length:6,16" dc:"新密码"`
}

type ChangePasswordRes struct {
}

type UpdateProfileReq struct {
	g.Meta   `path:"users/updateProfile" method:"post" sm:"修改信息" tags:"用户管理"`
	Nickname string `json:"nickname" v:"length:0,50" dc:"昵称"`
	Email    string `json:"email" v:"email|length:0,100" dc:"邮箱"`
}

type UpdateProfileRes struct {
}
