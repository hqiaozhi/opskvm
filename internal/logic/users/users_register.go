package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/crypto/bcrypt"
)

func (u *Users) Register(ctx context.Context, username, password, nickname, email string) (int, error) {
	var existUser entity.Users
	err := dao.Users.Ctx(ctx).Where("username", username).Scan(&existUser)
	if err != nil {
		return 0, gerror.New("database query error")
	}

	if existUser.Username == username {
		return 0, gerror.New("username already exists")
	}

	if email != "" {
		var existEmail entity.Users
		err := dao.Users.Ctx(ctx).Where("email", email).Scan(&existEmail)
		if err != nil {
			return 0, gerror.New("database query error")
		}
		if existEmail.Email == email {
			return 0, gerror.New("email already exists")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, gerror.New("password encryption failed")
	}

	if nickname == "" {
		nickname = username
	}

	userId, err := dao.Users.Ctx(ctx).InsertAndGetId(entity.Users{
		Username:  username,
		Password:  string(hashedPassword),
		Nickname:  nickname,
		Email:     email,
		IsAdmin:   0,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	})
	if err != nil {
		return 0, gerror.New("failed to create user")
	}

	return int(userId), nil
}
