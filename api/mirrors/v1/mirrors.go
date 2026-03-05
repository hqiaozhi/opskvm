package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type FileMeta struct {
	FileName string `json:"file_name" v:"required" dc:"文件名，如 ubuntu-22.04."`
	Size     int64  `json:"size" v:"required" dc:"文件大小（字节）"`
	Md5      string `json:"md5" v:"required" dc:"文件MD5校验值"`
}

type InitUploadReq struct {
	g.Meta    `path:"mirrors/init" method:"post" sm:"初始化分块上传" tags:"镜像管理"`
	Files     []FileMeta `json:"files" v:"required" dc:"文件列表"`
	ChunkSize int64      `json:"chunk_size" dc:"分块大小（字节），默认 5MB"`
}

type UploadSessionInfo struct {
	UploadID   string `json:"upload_id" dc:"上传会话ID"`
	FileName   string `json:"file_name" dc:"文件名"`
	FileSize   int64  `json:"file_size" dc:"文件大小"`
	ChunkSize  int64  `json:"chunk_size" dc:"分块大小"`
	ChunkCount int    `json:"chunk_count" dc:"总分块数"`
	FileMd5    string `json:"file_md5" dc:"文件SHA256校验值"`
	SkipUpload bool   `json:"skip_upload" dc:"是否跳过上传（秒传）"`
}

type InitUploadRes struct {
	Uploads []UploadSessionInfo `json:"uploads" dc:"上传会话列表"`
}

type UploadChunkReq struct {
	g.Meta     `path:"mirrors/upload/:uploadId/:chunkIndex" method:"put" sm:"上传分块" tags:"镜像管理"`
	UploadId   string `json:"upload_id" dc:"上传会话ID"`
	ChunkIndex int    `json:"chunk_index" dc:"分块索引，从 0 开始"`
}

type UploadChunkRes struct {
	ChunkIndex int   `json:"chunk_index" dc:"已上传的分块索引"`
	Offset     int64 `json:"offset" dc:"当前已上传的偏移量"`
}

type CompleteUploadReq struct {
	g.Meta   `path:"mirrors/upload/:uploadId/complete" method:"post" sm:"完成分块上传" tags:"镜像管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type CompleteUploadRes struct {
	UploadId  string `json:"upload_id" dc:"上传会话ID"`
	Id        string `json:"id" dc:"镜像ID"`
	FileName  string `json:"file_name" dc:"文件名"`
	FileSize  int64  `json:"file_size" dc:"文件大小"`
	LocalPath string `json:"local_path" dc:"本地存储路径"`
	URL       string `json:"url" dc:"镜像下载地址"`
}

type CancelUploadReq struct {
	g.Meta   `path:"mirrors/upload/:uploadId" method:"delete" sm:"取消分块上传" tags:"镜像管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type CancelUploadRes struct {
}

type GetUploadStatusReq struct {
	g.Meta   `path:"mirrors/upload/:uploadId" method:"get" sm:"查询上传状态" tags:"镜像管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type GetUploadStatusRes struct {
	UploadID      string `json:"upload_id" dc:"上传会话ID"`
	FileName      string `json:"file_name" dc:"文件名"`
	FileSize      int64  `json:"file_size" dc:"文件大小"`
	ChunkSize     int64  `json:"chunk_size" dc:"分块大小"`
	ChunkCount    int    `json:"chunk_count" dc:"总分块数"`
	UploadedCount int    `json:"uploaded_count" dc:"已上传分块数"`
	UploadedSize  int64  `json:"uploaded_size" dc:"已上传大小"`
}

type UploadByUrlReq struct {
	g.Meta   `path:"mirrors/url" method:"post" sm:"通过链接上传镜像" tags:"镜像管理"`
	Url      string `json:"url" v:"required" dc:"镜像下载地址URL"`
	FileName string `json:"file_name" dc:"指定保存的文件名，默认从URL中提取"`
}

type UploadByUrlRes struct {
	UploadId  string `json:"upload_id" dc:"上传任务ID"`
	FileName  string `json:"file_name" dc:"文件名"`
	FileSize  int64  `json:"file_size" dc:"文件大小"`
	Status    string `json:"status" dc:"上传状态：downloading/completed/failed"`
	LocalPath string `json:"local_path" dc:"本地存储路径"`
	URL       string `json:"url" dc:"镜像下载地址"`
}

type GetUrlUploadStatusReq struct {
	g.Meta   `path:"mirrors/url/:uploadId" method:"get" sm:"查询URL上传进度" tags:"镜像管理"`
	UploadId string `json:"upload_id" dc:"上传任务ID"`
}

type GetUrlUploadStatusRes struct {
	UploadId  string `json:"upload_id" dc:"上传任务ID"`
	FileName  string `json:"file_name" dc:"文件名"`
	FileSize  int64  `json:"file_size" dc:"文件大小"`
	Status    string `json:"status" dc:"上传状态：downloading/completed/failed"`
	LocalPath string `json:"local_path" dc:"本地存储路径"`
	URL       string `json:"url" dc:"镜像下载地址"`
	Error     string `json:"error" dc:"错误信息"`
}

type ListReq struct {
	g.Meta `path:"mirrors" method:"get" sm:"获取镜像列表" tags:"镜像管理"`
	Path   string `json:"path" dc:"目录路径，默认为根目录"`
	Page   int    `json:"page" dc:"页码，默认 1"`
	Limit  int    `json:"limit" dc:"每页数量，默认 20"`
}

type Info struct {
	Id        string `json:"id" dc:"镜像ID"`
	Name      string `json:"name" dc:"文件名"`
	Path      string `json:"path" dc:"完整路径"`
	Size      int64  `json:"size" dc:"文件大小（字节）"`
	URL       string `json:"url" dc:"镜像下载地址"`
	CreatedAt string `json:"created_at" dc:"创建时间"`
	UpdatedAt string `json:"updated_at" dc:"更新时间"`
}

type ListRes struct {
	Total int    `json:"total" dc:"镜像总数"`
	Page  int    `json:"page" dc:"当前页码"`
	Limit int    `json:"limit" dc:"每页数量"`
	List  []Info `json:"_list" dc:"镜像列表"`
}

type DeleteReq struct {
	g.Meta `path:"mirrors" method:"delete" sm:"删除镜像" tags:"镜像管理"`
	Ids    string `json:"ids" dc:"镜像ID或路径列表，逗号分隔"`
}

type DeleteRes struct {
	DeletedCount int      `json:"deleted_count" dc:"成功删除数量"`
	FailedIds    []string `json:"failed_ids" dc:"删除失败的ID或路径"`
}

type MsdManagerReq struct {
	g.Meta `path:"mirrors/msd" method:"post" sm:"虚拟介质" tags:"镜像管理"`
	Path   string `json:"path" dc:"镜像路径"`
	Mode   string `json:"mode" v:"required" dc:"挂载模式: 0(Flash)/1(CD/DVD)/2(取消挂载)"`
}
type MsdManagerRes struct {
}
