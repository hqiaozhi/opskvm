package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
)

type UserInfo struct {
	Id               int
	Username         string
	Nickname         string
	Email            string
	IsAdmin          int
	TwoFactorEnabled int
	CreatedAt        string
}

func (u *Users) GetUserInfo(ctx context.Context, userId int) (*UserInfo, error) {
	var user entity.Users
	cls := dao.Users.Columns()
	err := dao.Users.Ctx(ctx).Where(cls.Id, userId).Scan(&user)
	if err != nil {
		return nil, gerror.New("database query error")
	}

	if user.Id == 0 {
		return nil, gerror.New("user not found")
	}

	return &UserInfo{
		Id:               user.Id,
		Username:         user.Username,
		Nickname:         user.Nickname,
		Email:            user.Email,
		IsAdmin:          user.IsAdmin,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt.Format("Y-m-d H:i:s"),
	}, nil
}

func (u *Users) GetUserById(ctx context.Context, userId int) (*entity.Users, error) {
	var user entity.Users
	cls := dao.Users.Columns()
	err := dao.Users.Ctx(ctx).Where(cls.Id, userId).Scan(&user)
	if err != nil {
		return nil, gerror.New("database query error")
	}

	if user.Id == 0 {
		return nil, gerror.New("user not found")
	}

	return &user, nil
}

func (u *Users) IsAdmin(ctx context.Context, userId int) (bool, error) {
	var user entity.Users
	cls := dao.Users.Columns()
	err := dao.Users.Ctx(ctx).Where(cls.Id, userId).Scan(&user)
	if err != nil {
		return false, nil
	}

	if user.IsAdmin == 0 {
		return false, nil
	}

	return user.IsAdmin == 1, nil
}
