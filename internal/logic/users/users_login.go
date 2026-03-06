package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"
)

func (u *Users) Login(ctx context.Context, username, password string) (int, error) {

	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("username", username).Scan(&user)
	if err != nil {
		return 0, gerror.New("username or password error")
	}

	if user.Username != username {
		return 0, gerror.New("username do not exist")
	}

	if !u.CheckPassword(password, user.Password) {
		return 0, gerror.New("username or password error")
	}

	return user.Id, nil
}

func (u *Users) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u *Users) CheckPassword(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
