package wake

import (
	"context"
	"opskvm/internal/dao"
	"strconv"
)

func (w *Wake) DeleteWol(ctx context.Context, id int) error {
	_, err := dao.Wol.Ctx(ctx).Where("id", id).Delete()
	return err
}

func (w *Wake) DeleteWolBatch(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.Itoa(id)
	}
	result, err := dao.Wol.Ctx(ctx).WhereIn("id", idStrs).Delete()
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}
