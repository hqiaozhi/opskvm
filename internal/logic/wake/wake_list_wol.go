package wake

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"
)

func (w *Wake) ListWol(ctx context.Context, page, limit int, deviceName, macAddr string) ([]entity.Wol, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	model := dao.Wol.Ctx(ctx).Page(page, limit).Order("id DESC")
	if deviceName != "" {
		model = model.Where("device_name LIKE ?", "%"+deviceName+"%")
	}
	if macAddr != "" {
		model = model.Where("mac_addr LIKE ?", "%"+macAddr+"%")
	}
	var list []entity.Wol
	err := model.Scan(&list)
	if err != nil {
		return nil, 0, err
	}

	total := 0
	if deviceName != "" && macAddr != "" {
		total, _ = dao.Wol.Ctx(ctx).Where("device_name LIKE ? AND mac_addr LIKE ?", "%"+deviceName+"%", "%"+macAddr+"%").Count()
	} else if deviceName != "" {
		total, _ = dao.Wol.Ctx(ctx).Where("device_name LIKE ?", "%"+deviceName+"%").Count()
	} else if macAddr != "" {
		total, _ = dao.Wol.Ctx(ctx).Where("mac_addr LIKE ?", "%"+macAddr+"%").Count()
	} else {
		total, _ = dao.Wol.Ctx(ctx).Count()
	}

	return list, total, nil
}
