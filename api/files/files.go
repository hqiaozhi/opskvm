// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package files

import (
	"context"

	"opskvm/api/files/v1"
)

type IFilesV1 interface {
	InitUpload(ctx context.Context, req *v1.InitUploadReq) (res *v1.InitUploadRes, err error)
	UploadChunk(ctx context.Context, req *v1.UploadChunkReq) (res *v1.UploadChunkRes, err error)
	CompleteUpload(ctx context.Context, req *v1.CompleteUploadReq) (res *v1.CompleteUploadRes, err error)
	CompleteAllUploads(ctx context.Context, req *v1.CompleteAllUploadsReq) (res *v1.CompleteAllUploadsRes, err error)
	CancelUpload(ctx context.Context, req *v1.CancelUploadReq) (res *v1.CancelUploadRes, err error)
	GetUploadStatus(ctx context.Context, req *v1.GetUploadStatusReq) (res *v1.GetUploadStatusRes, err error)
	ListFiles(ctx context.Context, req *v1.ListFilesReq) (res *v1.ListFilesRes, err error)
	GetFileInfo(ctx context.Context, req *v1.GetFileInfoReq) (res *v1.GetFileInfoRes, err error)
	DeleteFile(ctx context.Context, req *v1.DeleteFileReq) (res *v1.DeleteFileRes, err error)
	DownloadFile(ctx context.Context, req *v1.DownloadFileReq) (res *v1.DownloadFileRes, err error)
	GetStorageInfo(ctx context.Context, req *v1.GetStorageInfoReq) (res *v1.GetStorageInfoRes, err error)
	CreateDirectory(ctx context.Context, req *v1.CreateDirectoryReq) (res *v1.CreateDirectoryRes, err error)
}
