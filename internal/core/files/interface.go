package files

// FilesManager 文件管理接口
// 负责ISO镜像文件的上传、下载、列表、查询和删除等操作
type FilesManager interface {
	// GetISOList 获取ISO文件列表
	// 返回：文件信息列表，错误信息
	GetISOList() ([]FileInfo, error)

	// GetISOAbsolutePath 获取ISO文件的绝对路径
	// 参数：fileName 文件名
	// 返回：文件绝对路径，错误信息
	GetISOAbsolutePath(fileName string) (string, error)

	// DeleteISO 删除ISO文件
	// 参数：fileName 文件名
	// 返回：错误信息
	DeleteISO(fileName string) error

	// GetUploadProgress 获取上传进度
	// 参数：fileName 文件名
	// 返回：进度信息，错误信息
	GetUploadProgress(fileName string) (*ProgressInfo, error)

	// GetDownloadProgress 获取下载进度
	// 参数：fileName 文件名
	// 返回：进度信息，错误信息
	GetDownloadProgress(fileName string) (*ProgressInfo, error)

	// UploadISOWithChunk 分片上传ISO镜像文件
	// 参数：fileName 文件名，chunkNumber 分片序号，totalChunks 总分片数，chunkData 分片数据
	// 返回：文件绝对路径（当所有分片上传完成时），是否完成，错误信息
	UploadISOWithChunk(fileName string, chunkNumber, totalChunks int, chunkData []byte) (string, bool, error)

	// DownloadISOByURLWithProgress 根据URL下载ISO文件并跟踪进度
	// 参数：url 文件下载链接
	// 返回：文件绝对路径，错误信息
	DownloadISOByURLWithProgress(url string) (string, error)
}