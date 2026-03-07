package users

import (
	"context"
	"opskvm/internal/dao"
	"opskvm/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (u *Users) GetUserList(ctx context.Context, page, limit int) (int, []*UserInfo, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	var total int
	total, err := dao.Users.Ctx(ctx).Count()
	if err != nil {
		return 0, nil, gerror.New("database query error")
	}

	var users []*entity.Users
	err = dao.Users.Ctx(ctx).Page(page, limit).Order("id DESC").Scan(&users)
	if err != nil {
		return 0, nil, gerror.New("database query error")
	}

	var result []*UserInfo
	for _, user := range users {
		result = append(result, &UserInfo{
			Id:               user.Id,
			Username:         user.Username,
			Nickname:         user.Nickname,
			Email:            user.Email,
			IsAdmin:          user.IsAdmin,
			TwoFactorEnabled: user.TwoFactorEnabled,
			CreatedAt:        user.CreatedAt.Format("Y-m-d H:i:s"),
		})
	}

	return total, result, nil
}
