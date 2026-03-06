package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/crypto/bcrypt"
)

func (u *Users) ChangePassword(ctx context.Context, userId int, oldPassword, newPassword string) error {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return gerror.New("database query error")
	}

	if user.Id == 0 {
		return gerror.New("user not found")
	}

	if !u.CheckPassword(oldPassword, user.Password) {
		return gerror.New("old password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return gerror.New("password encryption failed")
	}

	_, err = dao.Users.Ctx(ctx).Where("id", userId).Data(entity.Users{
		Password:  string(hashedPassword),
		UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.New("failed to update password")
	}

	return nil
}
