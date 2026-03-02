package wake

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"
)

func (w *Wake) GetWol(ctx context.Context, id int) (*entity.Wol, error) {
	var wol entity.Wol
	err := dao.Wol.Ctx(ctx).Where("id", id).Scan(&wol)
	if err != nil {
		return nil, err
	}
	return &wol, nil
}

func (w *Wake) GetWolByMac(ctx context.Context, macAddr string) (*entity.Wol, error) {
	var wol entity.Wol
	err := dao.Wol.Ctx(ctx).Where("mac_addr", macAddr).Scan(&wol)
	if err != nil {
		return nil, err
	}
	return &wol, nil
}
