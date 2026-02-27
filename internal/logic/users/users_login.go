package users

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/frame/g"
)

func (u *Users) Login(ctx context.Context, username, password string) error {
	user_name := g.Cfg().MustGetWithCmd(ctx, `username`, "admin")
	if user_name == nil {
		user_name = g.Cfg().MustGetWithCmd(ctx, `u`, "admin")
	}

	pass_word := g.Cfg().MustGetWithCmd(ctx, `password`, "admin123")
	if pass_word == nil {
		pass_word = g.Cfg().MustGetWithCmd(ctx, `p`, "admin123")
	}

	// 判断用户名或密码
	if user_name.String() != username || pass_word.String() != password {
		return errors.New("username or password error")
	}
	return nil
}
