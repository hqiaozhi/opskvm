package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type InitUploadReq struct {
	g.Meta    `path:"files/upload/init" method:"post" sm:"初始化分块上传" tags:"文件管理"`
	Files     []FileMeta `json:"files" v:"required" dc:"文件列表，支持多文件和目录"`
	ChunkSize int64      `json:"chunk_size" dc:"分块大小（字节），默认 5MB"`
}

type FileMeta struct {
	Path string `json:"path" v:"required" dc:"文件路径（含目录），如 folder/file.txt"`
	Size int64  `json:"size" v:"required" dc:"文件大小（字节）"`
	Md5  string `json:"md5" v:"required" dc:"文件MD5校验值"`
}

type UploadSessionInfo struct {
	UploadID   string `json:"upload_id" dc:"上传会话ID"`
	Path       string `json:"path" dc:"文件路径"`
	FileName   string `json:"file_name" dc:"文件名"`
	FileSize   int64  `json:"file_size" dc:"文件大小"`
	ChunkSize  int64  `json:"chunk_size" dc:"分块大小"`
	ChunkCount int    `json:"chunk_count" dc:"总分块数"`
	IsDir      bool   `json:"is_dir" dc:"是否为目录"`
	FileMd5    string `json:"file_md5" dc:"文件SHA256校验值"`
	SkipUpload bool   `json:"skip_upload" dc:"是否跳过上传（秒传）"`
}

type InitUploadRes struct {
	Uploads []UploadSessionInfo `json:"uploads" dc:"上传会话列表"`
}

type UploadChunkReq struct {
	g.Meta     `path:"files/upload/:uploadId/:chunkIndex" method:"put" sm:"上传分块" tags:"文件管理"`
	UploadId   string `json:"upload_id" dc:"上传会话ID"`
	ChunkIndex int    `json:"chunk_index" dc:"分块索引，从 0 开始"`
}

type UploadChunkRes struct {
	ChunkIndex int   `json:"chunk_index" dc:"已上传的分块索引"`
	Offset     int64 `json:"offset" dc:"当前已上传的偏移量"`
}

type CompleteUploadReq struct {
	g.Meta   `path:"files/upload/:uploadId/complete" method:"post" sm:"完成分块上传" tags:"文件管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type CompleteUploadRes struct {
	UploadId string `json:"upload_id" dc:"上传会话ID"`
	FileId   string `json:"file_id" dc:"文件ID"`
	FilePath string `json:"file_path" dc:"最终文件路径"`
	FileName string `json:"file_name" dc:"文件名"`
	FileSize int64  `json:"file_size" dc:"文件大小"`
}

type CompleteAllUploadsReq struct {
	g.Meta `path:"files/upload/complete-all" method:"post" sm:"完成所有分块上传" tags:"文件管理"`
}

type CompleteAllUploadsRes struct {
	CompletedCount int                 `json:"completed_count" dc:"成功完成数量"`
	Results        []CompleteUploadRes `json:"results" dc:"完成结果列表"`
}

type CancelUploadReq struct {
	g.Meta   `path:"files/upload/:uploadId" method:"delete" sm:"取消分块上传" tags:"文件管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type CancelUploadRes struct {
}

type GetUploadStatusReq struct {
	g.Meta   `path:"files/upload/:uploadId" method:"get" sm:"查询上传状态" tags:"文件管理"`
	UploadId string `json:"upload_id" dc:"上传会话ID"`
}

type GetUploadStatusRes struct {
	UploadID      string `json:"upload_id" dc:"上传会话ID"`
	Path          string `json:"path" dc:"文件路径"`
	FileName      string `json:"file_name" dc:"文件名"`
	FileSize      int64  `json:"file_size" dc:"文件大小"`
	ChunkSize     int64  `json:"chunk_size" dc:"分块大小"`
	ChunkCount    int    `json:"chunk_count" dc:"总分块数"`
	UploadedCount int    `json:"uploaded_count" dc:"已上传分块数"`
	UploadedSize  int64  `json:"uploaded_size" dc:"已上传大小"`
}

type ListFilesReq struct {
	g.Meta `path:"files" method:"get" sm:"获取文件列表" tags:"文件管理"`
	Path   string `json:"path" dc:"目录路径，默认为根目录"`
	Page   int    `json:"page" dc:"页码，默认 1"`
	Limit  int    `json:"limit" dc:"每页数量，默认 20"`
}

type FileInfo struct {
	Id        string `json:"id" dc:"文件ID"`
	Name      string `json:"name" dc:"文件名或目录名"`
	Path      string `json:"path" dc:"完整路径"`
	RelPath   string `json:"rel_path" dc:"相对路径"`
	Size      int64  `json:"size" dc:"文件大小（字节）"`
	IsDir     bool   `json:"is_dir" dc:"是否为目录"`
	URL       string `json:"url" dc:"文件下载地址"`
	CreatedAt string `json:"created_at" dc:"创建时间"`
	UpdatedAt string `json:"updated_at" dc:"更新时间"`
}

type ListFilesRes struct {
	Total    int        `json:"total" dc:"文件总数（含目录）"`
	Page     int        `json:"page" dc:"当前页码"`
	Limit    int        `json:"limit" dc:"每页数量"`
	FileList []FileInfo `json:"file_list" dc:"文件列表"`
}

type GetFileInfoReq struct {
	g.Meta `path:"files/info/:fileId" method:"head" sm:"获取文件信息" tags:"文件管理"`
	FileId string `json:"file_id" dc:"文件ID或文件路径"`
}

type GetFileInfoRes struct {
	Name      string `json:"name" dc:"文件名"`
	Path      string `json:"path" dc:"完整路径"`
	Size      int64  `json:"size" dc:"文件大小"`
	IsDir     bool   `json:"is_dir" dc:"是否为目录"`
	CreatedAt string `json:"created_at" dc:"创建时间"`
}

type DeleteFileReq struct {
	g.Meta  `path:"files" method:"delete" sm:"删除文件或目录" tags:"文件管理"`
	FileIds string `json:"file_ids" dc:"文件ID或路径列表，逗号分隔，支持多级目录如 folder/subfolder"`
}

type DeleteFileRes struct {
	DeletedCount int      `json:"deleted_count" dc:"成功删除数量"`
	FailedIds    []string `json:"failed_ids" dc:"删除失败的路径"`
}

type GetStorageInfoReq struct {
	g.Meta `path:"files/storage" method:"get" sm:"获取存储空间信息" tags:"文件管理"`
}

type GetStorageInfoRes struct {
	TotalSpace     int64 `json:"total_space" dc:"总空间（字节）"`
	UsedSpace      int64 `json:"used_space" dc:"已使用空间（字节）"`
	AvailableSpace int64 `json:"available_space" dc:"可用空间（字节）"`
	UsedPercent    int   `json:"used_percent" dc:"使用百分比"`
}

type CreateDirectoryReq struct {
	g.Meta `path:"files/dir" method:"post" sm:"创建目录" tags:"文件管理"`
	Path   string `json:"path" v:"required" dc:"目录路径，如 folder/subfolder"`
}

type CreateDirectoryRes struct {
	Path string `json:"path" dc:"创建的目录路径"`
}

type DeviceInfo struct {
	DevPath     string `json:"dev_path" dc:"设备路径，如/dev/sda"`
	MountPoint  string `json:"mount_point" dc:"挂载点，如/mnt/usb"`
	Name        string `json:"name" dc:"设备名称，如sda"`
	Model       string `json:"model" dc:"设备型号，硬盘/优盘特有"`
	Capacity    string `json:"capacity" dc:"设备容量，硬盘/优盘特有"`
	Partition   string `json:"partition" dc:"分区号，如1，分区特有"`
	FsType      string `json:"fs_type" dc:"文件系统类型，如ext4/vfat"`
	Size        int64  `json:"size" dc:"分区大小（字节）"`
	Used        int64  `json:"used" dc:"已使用空间（字节）"`
	Available   int64  `json:"available" dc:"可用空间（字节）"`
	UsedPercent int    `json:"used_percent" dc:"使用百分比"`
	IsPartition bool   `json:"is_partition" dc:"是否为分区"`
	IsMounted   bool   `json:"is_mounted" dc:"是否已挂载"`
}

type ListDevicesReq struct {
	g.Meta `path:"files/devices" method:"get" sm:"获取设备列表" tags:"文件管理"`
}

type ListDevicesRes struct {
	Devices []DeviceInfo `json:"devices" dc:"设备列表"`
}

type MountDeviceReq struct {
	g.Meta  `path:"files/mount" method:"post" sm:"挂载/卸载设备" tags:"文件管理"`
	DevPath string `json:"dev_path" dc:"设备路径，如/dev/sda1，留空则执行卸载"`
}

type MountDeviceRes struct {
	DevPath    string `json:"dev_path" dc:"设备路径"`
	MountPoint string `json:"mount_point" dc:"挂载点"`
}
