// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package mirrors

import (
	"context"

	"opskvm/api/mirrors/v1"
)

type IMirrorsV1 interface {
	InitUpload(ctx context.Context, req *v1.InitUploadReq) (res *v1.InitUploadRes, err error)
	UploadChunk(ctx context.Context, req *v1.UploadChunkReq) (res *v1.UploadChunkRes, err error)
	CompleteUpload(ctx context.Context, req *v1.CompleteUploadReq) (res *v1.CompleteUploadRes, err error)
	CancelUpload(ctx context.Context, req *v1.CancelUploadReq) (res *v1.CancelUploadRes, err error)
	GetUploadStatus(ctx context.Context, req *v1.GetUploadStatusReq) (res *v1.GetUploadStatusRes, err error)
	UploadByUrl(ctx context.Context, req *v1.UploadByUrlReq) (res *v1.UploadByUrlRes, err error)
	GetUrlUploadStatus(ctx context.Context, req *v1.GetUrlUploadStatusReq) (res *v1.GetUrlUploadStatusRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
}
