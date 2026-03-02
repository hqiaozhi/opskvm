// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Wol is the golang structure of table wol for DAO operations like Where/Data.
type Wol struct {
	g.Meta      `orm:"table:wol, do:true"`
	Id          any         //
	DeviceName  any         //
	MacAddr     any         //
	BroadcastIp any         //
	Port        any         //
	Remark      any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
