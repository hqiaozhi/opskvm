package totp

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (t *Totp) Status(ctx context.Context, userId int) (bool, string, error) {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, "", gerror.New("user not found")
	}

	if user.TwoFactorEnabled == 1 {
		return true, "TOTP is enabled", nil
	}
	return false, "TOTP is not enabled", nil
}
