package users

import (
	"context"
	v1 "opskvm/api/users/v1"
	"strconv"

	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	userIdStr := r.GetParam("userid")
	userId, err := strconv.Atoi(userIdStr.String())
	if err != nil {
		return nil, err
	}

	err = c.users.UpdateProfile(ctx, userId, req.Nickname, req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateProfileRes{}, nil
}
