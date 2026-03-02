// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package wake

import (
	"context"

	"opskvm/api/wake/v1"
)

type IWakeV1 interface {
	CreateWol(ctx context.Context, req *v1.CreateWolReq) (res *v1.CreateWolRes, err error)
	UpdateWol(ctx context.Context, req *v1.UpdateWolReq) (res *v1.UpdateWolRes, err error)
	DeleteWol(ctx context.Context, req *v1.DeleteWolReq) (res *v1.DeleteWolRes, err error)
	DeleteWolBatch(ctx context.Context, req *v1.DeleteWolBatchReq) (res *v1.DeleteWolBatchRes, err error)
	GetWol(ctx context.Context, req *v1.GetWolReq) (res *v1.GetWolRes, err error)
	ListWol(ctx context.Context, req *v1.ListWolReq) (res *v1.ListWolRes, err error)
	WakeOnLan(ctx context.Context, req *v1.WakeOnLanReq) (res *v1.WakeOnLanRes, err error)
}
