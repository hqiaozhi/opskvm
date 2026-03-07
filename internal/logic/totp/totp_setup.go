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
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var tempSecretCache = gcache.New()
var fixedCtx = gctx.New()

func (t *Totp) Setup(ctx context.Context, userId int) (secret, qrCode string, err error) {
	var user entity.Users
	err = dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return "", "", gerror.New("user not found")
	}

	if user.TwoFactorEnabled == 1 {
		return "", "", gerror.New("TOTP is already enabled")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OpsKVM",
		AccountName: user.Username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", gerror.New("failed to generate TOTP key")
	}

	secret = key.Secret()
	qrCode = key.URL()

	cacheKey := fmt.Sprintf("temp_secret_%d", userId)
	tempSecretCache.Set(fixedCtx, cacheKey, secret, 300*time.Second)

	return secret, qrCode, nil
}

func (t *Totp) GetTempSecret(ctx context.Context, userId int) (string, error) {
	cacheKey := fmt.Sprintf("temp_secret_%d", userId)
	v, err := tempSecretCache.Get(fixedCtx, cacheKey)
	if err != nil {
		return "", gerror.New("no pending TOTP setup, please call setup first")
	}
	return v.String(), nil
}

func (t *Totp) ClearTempSecret(ctx context.Context, userId int) {
	cacheKey := fmt.Sprintf("temp_secret_%d", userId)
	tempSecretCache.Remove(fixedCtx, cacheKey)
}
