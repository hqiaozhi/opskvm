package wake

import (
	"context"
	"opskvm/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

func (w *Wake) UpdateWol(ctx context.Context, id int, deviceName, macAddr, broadcastIp string, port int, remark string) error {
	data := g.Map{}
	if deviceName != "" {
		data["device_name"] = deviceName
	}
	if macAddr != "" {
		data["mac_addr"] = macAddr
	}
	if broadcastIp != "" {
		data["broadcast_ip"] = broadcastIp
	}
	if port > 0 {
		data["port"] = port
	}
	if remark != "" {
		data["remark"] = remark
	}
	if len(data) == 0 {
		return nil
	}
	_, err := dao.Wol.Ctx(ctx).Where("id", id).Data(data).Update()
	return err
}
