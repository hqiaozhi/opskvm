package v1

import "github.com/gogf/gf/v2/frame/g"

type SetupReq struct {
	g.Meta `path:"totp/setup" method:"post" sm:"生成TOTP密钥" tags:"两步验证"`
}

type SetupRes struct {
	Secret string `json:"secret" dc:"TOTP密钥(Base32编码)"`
	QRCode string `json:"qrCode" dc:"二维码内容(otpauth://...)"`
}

type EnableReq struct {
	g.Meta `path:"totp/enable" method:"post" sm:"确认启用TOTP" tags:"两步验证"`
	Code   string `json:"code" v:"required|size:6" dc:"6位验证码(从验证器获取)"`
}

type EnableRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"提示信息"`
}

type VerifyReq struct {
	g.Meta `path:"totp/verify" method:"post" sm:"验证TOTP" tags:"两步验证"`
	Code   string `json:"code" v:"required|size:6" dc:"6位验证码"`
}

type VerifyRes struct {
	Success bool   `json:"success" dc:"验证是否成功"`
	Message string `json:"message" dc:"提示信息"`
}

type DisableReq struct {
	g.Meta   `path:"totp/disable" method:"post" sm:"禁用TOTP" tags:"两步验证"`
	Code     string `json:"code" v:"required|size:6" dc:"6位验证码(验证通过后禁用)"`
	Password string `json:"password" v:"required" dc:"密码(二次确认)"`
}

type DisableRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"提示信息"`
}

type StatusReq struct {
	g.Meta `path:"totp/status" method:"get" sm:"查询状态" tags:"两步验证"`
}

type StatusRes struct {
	Enabled bool   `json:"enabled" dc:"是否已启用TOTP"`
	Message string `json:"message" dc:"提示信息"`
}
