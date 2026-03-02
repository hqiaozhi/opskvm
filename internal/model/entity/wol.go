// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Wol is the golang structure for table wol.
type Wol struct {
	Id          int         `json:"id"          orm:"id"           description:""` //
	DeviceName  string      `json:"deviceName"  orm:"device_name"  description:""` //
	MacAddr     string      `json:"macAddr"     orm:"mac_addr"     description:""` //
	BroadcastIp string      `json:"broadcastIp" orm:"broadcast_ip" description:""` //
	Port        int         `json:"port"        orm:"port"         description:""` //
	Remark      string      `json:"remark"      orm:"remark"       description:""` //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""` //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""` //
}
