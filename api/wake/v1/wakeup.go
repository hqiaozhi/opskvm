package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type CreateWolReq struct {
	g.Meta      `path:"wol" method:"post" sm:"创建WoL配置" tags:"WoL管理"`
	DeviceName  string `json:"deviceName" v:"required" dc:"设备名称"`
	MacAddr     string `json:"macAddr" v:"required|mac" dc:"MAC地址"`
	BroadcastIp string `json:"broadcastIp" dc:"广播IP，默认192.168.1.255"`
	Port        int    `json:"port" dc:"唤醒端口，默认9"`
	Remark      string `json:"remark" dc:"备注"`
}

type CreateWolRes struct {
	Id int `json:"id" dc:"创建的WoL配置ID"`
}

type UpdateWolReq struct {
	g.Meta      `path:"wol/:id" method:"put" sm:"更新WoL配置" tags:"WoL管理"`
	Id          int    `json:"id" v:"required" dc:"WoL配置ID"`
	DeviceName  string `json:"deviceName" dc:"设备名称"`
	MacAddr     string `json:"macAddr" v:"mac" dc:"MAC地址"`
	BroadcastIp string `json:"broadcastIp" dc:"广播IP"`
	Port        int    `json:"port" dc:"唤醒端口"`
	Remark      string `json:"remark" dc:"备注"`
}

type UpdateWolRes struct {
}

type DeleteWolReq struct {
	g.Meta `path:"wol/:id" method:"delete" sm:"删除WoL配置" tags:"WoL管理"`
	Id     int `json:"id" v:"required" dc:"WoL配置ID"`
}

type DeleteWolRes struct {
}

type DeleteWolBatchReq struct {
	g.Meta `path:"wol/batch" method:"delete" sm:"批量删除WoL配置" tags:"WoL管理"`
	Ids    string `json:"ids" v:"required" dc:"WoL配置ID列表，逗号分隔"`
}

type DeleteWolBatchRes struct {
	DeletedCount int `json:"deleted_count" dc:"成功删除数量"`
}

type GetWolReq struct {
	g.Meta `path:"wol/:id" method:"get" sm:"获取WoL配置详情" tags:"WoL管理"`
	Id     int `json:"id" v:"required" dc:"WoL配置ID"`
}

type WolInfo struct {
	Id          int    `json:"id" dc:"WoL配置ID"`
	DeviceName  string `json:"deviceName" dc:"设备名称"`
	MacAddr     string `json:"macAddr" dc:"MAC地址"`
	BroadcastIp string `json:"broadcastIp" dc:"广播IP"`
	Port        int    `json:"port" dc:"唤醒端口"`
	Remark      string `json:"remark" dc:"备注"`
	CreatedAt   string `json:"createdAt" dc:"创建时间"`
	UpdatedAt   string `json:"updatedAt" dc:"更新时间"`
}

type GetWolRes struct {
	Wol WolInfo `json:"wol" dc:"WoL配置信息"`
}

type ListWolReq struct {
	g.Meta     `path:"wol" method:"get" sm:"获取WoL配置列表" tags:"WoL管理"`
	Page       int    `json:"page" dc:"页码，默认1"`
	Limit      int    `json:"limit" dc:"每页数量，默认20"`
	DeviceName string `json:"deviceName" dc:"设备名称（模糊搜索）"`
	MacAddr    string `json:"macAddr" dc:"MAC地址（模糊搜索）"`
}

type ListWolRes struct {
	Total   int       `json:"total" dc:"总数量"`
	Page    int       `json:"page" dc:"当前页码"`
	Limit   int       `json:"limit" dc:"每页数量"`
	WolList []WolInfo `json:"wolList" dc:"WoL配置列表"`
}

type WakeOnLanReq struct {
	g.Meta      `path:"wol/wake" method:"post" sm:"唤醒设备" tags:"WoL管理"`
	BroadcastIp string `json:"broadcastIp" v:"required" dc:"广播IP"`
	Port        int    `json:"port" v:"required" dc:"唤醒端口"`
	MacAddr     string `json:"macAddr" v:"required|mac" dc:"MAC地址"`
}

type WakeOnLanRes struct {
	Success bool   `json:"success" dc:"是否发送成功"`
	Message string `json:"message" dc:"结果信息"`
}
