package totp

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/pquerna/otp/totp"
)

var totpVerifiedCache = gcache.New()
var totpEnabledCache = gcache.New()

func (t *Totp) Verify(ctx context.Context, userId int, code string) (bool, error) {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, gerror.New("user not found")
	}

	if user.TwoFactorEnabled != 1 {
		return false, gerror.New("TOTP is not enabled")
	}

	if user.TwoFactorSecret == "" {
		return false, gerror.New("TOTP secret not found, please setup TOTP first")
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return false, nil
	}

	totpVerifiedCache.Set(ctx, userId, true, 3600)

	return true, nil
}

func (t *Totp) IsVerified(ctx context.Context, userId int) bool {
	v, err := totpVerifiedCache.Get(ctx, userId)
	if err != nil {
		return false
	}
	return v.Bool()
}

func (t *Totp) ClearVerification(ctx context.Context, userId int) {
	totpVerifiedCache.Remove(ctx, userId)
}

func (t *Totp) NeedTotp(ctx context.Context, userId int) (bool, error) {
	v, err := totpEnabledCache.Get(ctx, userId)
	if err == nil {
		return v.Bool(), nil
	}

	var user entity.Users
	err = dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, gerror.New("user not found")
	}

	enabled := user.TwoFactorEnabled == 1
	totpEnabledCache.Set(ctx, userId, enabled, 3600)

	return enabled, nil
}

func (t *Totp) ClearTotpEnabledCache(ctx context.Context, userId int) {
	totpEnabledCache.Remove(ctx, userId)
}
