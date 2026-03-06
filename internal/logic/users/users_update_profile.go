package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

func (u *Users) UpdateProfile(ctx context.Context, userId int, nickname, email string) error {
	var user entity.Users
	err := dao.Users.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return gerror.New("database query error")
	}

	if user.Id == 0 {
		return gerror.New("user not found")
	}

	if email != "" && email != user.Email {
		var existEmail entity.Users
		err := dao.Users.Ctx(ctx).Where("email", email).Scan(&existEmail)
		if err != nil {
			return gerror.New("database query error")
		}
		if existEmail.Email == email && existEmail.Id != userId {
			return gerror.New("email already exists")
		}
	}

	updateData := entity.Users{
		UpdatedAt: gtime.Now(),
	}
	if nickname != "" {
		updateData.Nickname = nickname
	}
	if email != "" {
		updateData.Email = email
	}

	_, err = dao.Users.Ctx(ctx).Where("id", userId).Data(updateData).Update()
	if err != nil {
		return gerror.New("failed to update profile")
	}

	return nil
}
