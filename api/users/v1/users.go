package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"users/login" method:"post" sm:"登录" tags:"用户管理"`
	Username string `json:"username" v:"required|length:3,12" dc:"账户"`
	Password string `json:"password" v:"required|length:6,16" dc:"密码"`
}

type LoginRes struct {
	Token        string `json:"token" dc:"访问令牌"`
	TotpRequired bool   `json:"totpRequired" dc:"是否需要TOTP验证"`
	UserId       int    `json:"userId" dc:"用户ID"`
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
	UserId      int    `json:"userId" dc:"用户ID(管理员可传,否则修改自己)"`
	OldPassword string `json:"oldPassword" dc:"旧密码(管理员重置时不需要)"`
	NewPassword string `json:"newPassword" v:"required|length:6,16" dc:"新密码"`
}

type ChangePasswordRes struct {
}

type UpdateProfileReq struct {
	g.Meta   `path:"users/updateProfile" method:"post" sm:"修改信息" tags:"用户管理"`
	UserId   int    `json:"userId" dc:"用户ID(管理员可传,否则修改自己)"`
	Nickname string `json:"nickname" v:"length:0,50" dc:"昵称"`
	Email    string `json:"email" v:"email|length:0,100" dc:"邮箱"`
}

type UpdateProfileRes struct {
}

type GetUserInfoReq struct {
	g.Meta `path:"users/getUserInfo" method:"get" sm:"获取用户信息" tags:"用户管理"`
	UserId int `json:"userId" dc:"用户ID"`
}

type GetUserInfoRes struct {
	Id               int    `json:"id" dc:"用户ID"`
	Username         string `json:"username" dc:"账户"`
	Nickname         string `json:"nickname" dc:"昵称"`
	Email            string `json:"email" dc:"邮箱"`
	IsAdmin          int    `json:"isAdmin" dc:"是否管理员"`
	TwoFactorEnabled int    `json:"twoFactorEnabled" dc:"是否启用TOTP"`
	CreatedAt        string `json:"createdAt" dc:"创建时间"`
}

type GetUserListReq struct {
	g.Meta `path:"users/getUserList" method:"get" sm:"获取用户列表" tags:"用户管理"`
	Page   int `json:"page" dc:"页码"`
	Limit  int `json:"limit" dc:"每页数量"`
}

type GetUserListRes struct {
	Total int              `json:"total" dc:"总数"`
	List  []GetUserInfoRes `json:"list" dc:"用户列表"`
}

type DeleteUserReq struct {
	g.Meta `path:"users/deleteUser" method:"delete" sm:"删除用户" tags:"用户管理"`
	UserId int `json:"userId" dc:"用户ID"`
}

type DeleteUserRes struct {
}

type ResetPasswordReq struct {
	g.Meta      `path:"users/resetPassword" method:"post" sm:"重置密码" tags:"用户管理"`
	UserId      int    `json:"userId" v:"required" dc:"用户ID"`
	NewPassword string `json:"newPassword" v:"required|length:6,16" dc:"新密码"`
}

type ResetPasswordRes struct {
}
