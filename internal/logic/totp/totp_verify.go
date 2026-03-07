package totp

import (
	"context"
	"fmt"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/pquerna/otp/totp"
)

var totpVerifiedCache = gcache.New()
var totpVerifiedCacheCtx = gctx.New()

var totpEnabledCache = gcache.New()
var totpEnabledCacheCtx = gctx.New()

func (t *Totp) Verify(ctx context.Context, userId int, code string) (bool, error) {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, gerror.New("user not found")
	}

	if user.TwoFactorEnabled != 1 {
		return false, gerror.New("TOTP is not enabled")
	}

	if user.TotpSecret == "" {
		return false, gerror.New("TOTP secret not found, please setup TOTP first")
	}

	valid := totp.Validate(code, user.TotpSecret)
	if !valid {
		return false, nil
	}

	cacheKey := fmt.Sprintf("verified_%d", userId)
	totpVerifiedCache.Set(totpVerifiedCacheCtx, cacheKey, true, 3600*time.Second)

	return true, nil
}

func (t *Totp) IsVerified(ctx context.Context, userId int) bool {
	cacheKey := fmt.Sprintf("verified_%d", userId)
	v, err := totpVerifiedCache.Get(totpVerifiedCacheCtx, cacheKey)
	if err != nil {
		return false
	}
	return v.Bool()
}

func (t *Totp) ClearVerification(ctx context.Context, userId int) {
	cacheKey := fmt.Sprintf("verified_%d", userId)
	totpVerifiedCache.Remove(totpVerifiedCacheCtx, cacheKey)
}

func (t *Totp) NeedTotp(ctx context.Context, userId int) (bool, error) {
	cacheKey := fmt.Sprintf("enabled_%d", userId)
	v, err := totpEnabledCache.Get(totpEnabledCacheCtx, cacheKey)
	if err == nil {
		return v.Bool(), nil
	}

	var user entity.Users
	err = dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return false, gerror.New("user not found")
	}

	enabled := user.TwoFactorEnabled == 1
	totpEnabledCache.Set(totpEnabledCacheCtx, cacheKey, enabled, 3600*time.Second)

	return enabled, nil
}

func (t *Totp) ClearTotpEnabledCache(ctx context.Context, userId int) {
	cacheKey := fmt.Sprintf("enabled_%d", userId)
	totpEnabledCache.Remove(totpEnabledCacheCtx, cacheKey)
}
