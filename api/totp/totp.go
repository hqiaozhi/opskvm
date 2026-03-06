// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package totp

import (
	"context"

	"opskvm/api/totp/v1"
)

type ITotpV1 interface {
	Setup(ctx context.Context, req *v1.SetupReq) (res *v1.SetupRes, err error)
	Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error)
	Verify(ctx context.Context, req *v1.VerifyReq) (res *v1.VerifyRes, err error)
	Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error)
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
}
