package wake

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/do"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (w *Wake) CreateWol(ctx context.Context, deviceName, macAddr, broadcastIp string, port int, remark string) (int64, error) {
	count, err := dao.Wol.Ctx(ctx).Where("mac_addr", macAddr).Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.Newf("MAC地址 %s 已存在", macAddr)
	}

	data := do.Wol{
		DeviceName:  deviceName,
		MacAddr:     macAddr,
		BroadcastIp: broadcastIp,
		Port:        port,
		Remark:      remark,
	}
	if data.BroadcastIp == "" {
		data.BroadcastIp = "192.168.1.255"
	}
	if data.Port == 0 {
		data.Port = 9
	}
	result, err := dao.Wol.Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
