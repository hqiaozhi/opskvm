package users

import (
	"context"
	"opskvm/internal/dao"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (u *Users) DeleteUser(ctx context.Context, userId int) error {
	result, err := dao.Users.Ctx(ctx).Where("id", userId).Delete()
	if err != nil {
		return gerror.New("database delete error")
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return gerror.New("database delete error")
	}

	if affected == 0 {
		return gerror.New("user not found")
	}

	return nil
}
