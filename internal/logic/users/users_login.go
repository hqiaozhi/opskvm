package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"
)

type LoginResult struct {
	UserId           int
	Username         string
	TwoFactorEnabled bool
}

func (u *Users) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("username", username).Scan(&user)
	if err != nil {
		return nil, gerror.New("username or password error")
	}

	if user.Username == "" {
		return nil, gerror.New("username do not exist")
	}

	if !u.CheckPassword(password, user.Password) {
		return nil, gerror.New("username or password error")
	}

	return &LoginResult{
		UserId:           user.Id,
		Username:         user.Username,
		TwoFactorEnabled: user.TwoFactorEnabled == 1,
	}, nil
}

func (u *Users) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u *Users) CheckPassword(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
