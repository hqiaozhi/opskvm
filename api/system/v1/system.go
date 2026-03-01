package v1

import "github.com/gogf/gf/v2/frame/g"

type RestartSysReq struct {
	g.Meta `path:"system/restart" method:"post" sm:"重启系统" tags:"系统管理"`
}
type RestartSysRes struct{}
