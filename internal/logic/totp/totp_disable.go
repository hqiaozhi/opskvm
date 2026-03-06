package totp

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func (t *Totp) Disable(ctx context.Context, userId int, code, password string) (bool, error) {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, gerror.New("user not found")
	}

	if user.TwoFactorEnabled != 1 {
		return false, gerror.New("TOTP is not enabled")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, gerror.New("password is incorrect")
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return false, gerror.New("invalid TOTP code")
	}

	_, err = dao.Users.Ctx(ctx).Where("id", userId).Data(map[string]interface{}{
		"two_factor_secret":   "",
		"two_factor_enabled": 0,
		"updated_at":         gtime.Now(),
	}).Update()
	if err != nil {
		return false, gerror.New("failed to disable TOTP")
	}

	t.ClearTotpEnabledCache(ctx, userId)

	return true, nil
}
