// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id               int         `json:"id"               orm:"id"                 description:""` //
	Username         string      `json:"username"         orm:"username"           description:""` //
	Password         string      `json:"password"         orm:"password"           description:""` //
	Nickname         string      `json:"nickname"         orm:"nickname"           description:""` //
	Email            string      `json:"email"            orm:"email"              description:""` //
	IsAdmin          int         `json:"isAdmin"          orm:"is_admin"           description:""` //
	TwoFactorEnabled int         `json:"twoFactorEnabled" orm:"two_factor_enabled" description:""` //
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"         description:""` //
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"         description:""` //
}
