package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/do"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

func (u *Users) Register(ctx context.Context, username, password, nickname, email string) (int, error) {
	count, err := dao.Users.Ctx(ctx).Where("username", username).Count()
	if err != nil {
		g.Log().Errorf(ctx, "database query error: %s", err.Error())
		return 0, gerror.Newf("database query error: %s", err.Error())
	}

	if count > 0 {
		return 0, gerror.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, gerror.New("password encryption failed")
	}

	if nickname == "" {
		nickname = username
	}

	userId, err := dao.Users.Ctx(ctx).InsertAndGetId(do.Users{
		Username: username,
		Password: string(hashedPassword),
		Nickname: nickname,
		Email:    email,
		IsAdmin:  0,
	})
	if err != nil {
		g.Log().Errorf(ctx, "failed to create user: %s", err.Error())
		return 0, gerror.New("failed to create user")
	}

	return int(userId), nil
}
