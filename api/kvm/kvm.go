// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package kvm

import (
	"context"

	"opskvm/api/kvm/v1"
)

type IKvmV1 interface {
	Stream(ctx context.Context, req *v1.StreamReq) (res *v1.StreamRes, err error)
	TurnOn(ctx context.Context, req *v1.TurnOnReq) (res *v1.TurnOnRes, err error)
	TurnOff(ctx context.Context, req *v1.TurnOffReq) (res *v1.TurnOffRes, err error)
	GetConfig(ctx context.Context, req *v1.GetConfigReq) (res *v1.GetConfigRes, err error)
	GetConfigs(ctx context.Context, req *v1.GetConfigsReq) (res *v1.GetConfigsRes, err error)
	UpdateConfig(ctx context.Context, req *v1.UpdateConfigReq) (res *v1.UpdateConfigRes, err error)
	GetCompress(ctx context.Context, req *v1.GetCompressReq) (res *v1.GetCompressRes, err error)
	SetCompress(ctx context.Context, req *v1.SetCompressReq) (res *v1.SetCompressRes, err error)
}
