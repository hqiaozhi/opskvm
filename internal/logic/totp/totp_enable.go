package totp

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/pquerna/otp/totp"
)

func (t *Totp) Enable(ctx context.Context, userId int, secret, code string) error {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return gerror.New("user not found")
	}

	if user.TwoFactorEnabled == 1 {
		return gerror.New("TOTP is already enabled")
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return gerror.New("invalid TOTP code, please ensure your device time is correct")
	}

	_, err = dao.Users.Ctx(ctx).Where("id", userId).Data(map[string]interface{}{
		"two_factor_secret":  secret,
		"two_factor_enabled": 1,
		"updated_at":         gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.New("failed to enable TOTP")
	}

	t.ClearTotpEnabledCache(ctx, userId)

	return nil
}
