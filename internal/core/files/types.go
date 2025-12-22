package files

import "sync"

// ProgressInfo 进度信息结构体
type ProgressInfo struct {
	FileName     string `json:"file_name"`     // 文件名
	TotalSize    int64  `json:"total_size"`    // 总大小
	Completed    int64  `json:"completed"`     // 已完成大小
	Progress     int    `json:"progress"`      // 进度百分比
	Status       string `json:"status"`        // 状态：uploading, downloading, completed, failed
	LastModified int64  `json:"last_modified"` // 最后更新时间戳
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Name         string `json:"name"`          // 文件名
	Path         string `json:"path"`          // 文件路径
	Size         int64  `json:"size"`          // 文件大小（字节）
	IsISO        bool   `json:"is_iso"`        // 是否为ISO文件
	LastModified string `json:"last_modified"` // 最后修改时间
}

// ChunkInfo 分片信息结构体
type ChunkInfo struct {
	ChunkNumber int    `json:"chunk_number"`
	Data        []byte `json:"data"`
}

// LinuxFilesManager Linux平台的文件管理器实现
type LinuxFilesManager struct {
	baseDir          string                   // ISO文件存储的基础目录
	mutex            sync.Mutex               // 并发访问保护锁
	uploadProgress   map[string]*ProgressInfo // 上传进度映射
	downloadProgress map[string]*ProgressInfo // 下载进度映射
	chunkData        map[string][]ChunkInfo   // 分片数据映射
	chunkProgress    map[string]map[int]bool  // 分片上传进度映射
	totalChunkMap    map[string]int           // 总分片数映射
}
