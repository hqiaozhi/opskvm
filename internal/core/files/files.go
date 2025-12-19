package files

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProgressInfo 进度信息结构体
type ProgressInfo struct {
	FileName     string `json:"file_name"`     // 文件名
	TotalSize    int64  `json:"total_size"`    // 总大小
	Completed    int64  `json:"completed"`     // 已完成大小
	Progress     int    `json:"progress"`      // 进度百分比
	Status       string `json:"status"`        // 状态：uploading, downloading, completed, failed
	LastModified int64  `json:"last_modified"` // 最后更新时间戳
}

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

// NewLinuxFilesManager 创建文件管理器实例
// 参数：baseDir ISO文件存储的基础目录
// 返回：文件管理器实例，错误信息
func NewLinuxFilesManager(baseDir string) (*LinuxFilesManager, error) {
	// 验证基础目录是否存在
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		// 目录不存在则创建
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return nil, fmt.Errorf("创建基础目录失败: %w", err)
		}
	}

	return &LinuxFilesManager{
		baseDir:          baseDir,
		uploadProgress:   make(map[string]*ProgressInfo),
		downloadProgress: make(map[string]*ProgressInfo),
		chunkData:        make(map[string][]ChunkInfo),
		chunkProgress:    make(map[string]map[int]bool),
		totalChunkMap:    make(map[string]int),
	}, nil
}

// updateUploadProgress 更新上传进度
func (fm *LinuxFilesManager) updateUploadProgress(fileName string, completed, total int64, status string) {
	progress := int(float64(completed) / float64(total) * 100)
	if progress > 100 {
		progress = 100
	}

	fm.uploadProgress[fileName] = &ProgressInfo{
		FileName:     fileName,
		TotalSize:    total,
		Completed:    completed,
		Progress:     progress,
		Status:       status,
		LastModified: time.Now().Unix(),
	}
}

// updateDownloadProgress 更新下载进度
func (fm *LinuxFilesManager) updateDownloadProgress(fileName string, completed, total int64, status string) {
	progress := int(float64(completed) / float64(total) * 100)
	if progress > 100 {
		progress = 100
	}

	fm.downloadProgress[fileName] = &ProgressInfo{
		FileName:     fileName,
		TotalSize:    total,
		Completed:    completed,
		Progress:     progress,
		Status:       status,
		LastModified: time.Now().Unix(),
	}
}

// GetUploadProgress 获取上传进度
func (fm *LinuxFilesManager) GetUploadProgress(fileName string) (*ProgressInfo, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	progress, exists := fm.uploadProgress[fileName]
	if !exists {
		return nil, fmt.Errorf("未找到上传进度: %s", fileName)
	}

	return progress, nil
}

// GetDownloadProgress 获取下载进度
func (fm *LinuxFilesManager) GetDownloadProgress(fileName string) (*ProgressInfo, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	progress, exists := fm.downloadProgress[fileName]
	if !exists {
		return nil, fmt.Errorf("未找到下载进度: %s", fileName)
	}

	return progress, nil
}

// UploadISOWithChunk 分片上传ISO镜像文件
func (fm *LinuxFilesManager) UploadISOWithChunk(fileName string, chunkNumber, totalChunks int, chunkData []byte) (string, bool, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 验证文件名是否以.iso结尾
	if !strings.HasSuffix(strings.ToLower(fileName), ".iso") {
		return "", false, errors.New("仅支持ISO格式文件")
	}

	// 初始化分片进度映射
	if _, exists := fm.chunkProgress[fileName]; !exists {
		fm.chunkProgress[fileName] = make(map[int]bool)
		fm.totalChunkMap[fileName] = totalChunks
		fm.updateUploadProgress(fileName, 0, 0, "uploading")
	}

	// 验证总分片数是否一致
	if storedTotal, exists := fm.totalChunkMap[fileName]; exists && storedTotal != totalChunks {
		return "", false, fmt.Errorf("总分片数不匹配，期望: %d, 实际: %d", storedTotal, totalChunks)
	}

	// 验证分片序号是否有效
	if chunkNumber < 1 || chunkNumber > totalChunks {
		return "", false, fmt.Errorf("无效的分片序号: %d", chunkNumber)
	}

	// 检查分片是否已上传
	if fm.chunkProgress[fileName][chunkNumber] {
		// 分片已上传，跳过
		return "", false, nil
	}

	// 保存分片数据
	chunk := ChunkInfo{
		ChunkNumber: chunkNumber,
		Data:        chunkData,
	}

	// 添加或更新分片数据
	found := false
	for i, c := range fm.chunkData[fileName] {
		if c.ChunkNumber == chunkNumber {
			fm.chunkData[fileName][i] = chunk
			found = true
			break
		}
	}

	if !found {
		fm.chunkData[fileName] = append(fm.chunkData[fileName], chunk)
	}

	// 标记分片为已上传
	fm.chunkProgress[fileName][chunkNumber] = true

	// 计算已完成的分片数
	completedChunks := 0
	for _, uploaded := range fm.chunkProgress[fileName] {
		if uploaded {
			completedChunks++
		}
	}

	// 计算总大小（预估）
	var totalSize int64
	for _, chunk := range fm.chunkData[fileName] {
		totalSize += int64(len(chunk.Data))
	}

	// 预估总大小（根据已上传分片平均大小计算）
	if completedChunks > 0 {
		averageChunkSize := totalSize / int64(completedChunks)
		totalSize = averageChunkSize * int64(totalChunks)
	}

	// 更新上传进度
	completedSize := int64(len(chunkData)) * int64(completedChunks)
	fm.updateUploadProgress(fileName, completedSize, totalSize, "uploading")

	// 检查是否所有分片都已上传
	if completedChunks == totalChunks {
		// 构建文件绝对路径
		filePath := filepath.Join(fm.baseDir, fileName)

		// 创建目标文件
		targetFile, err := os.Create(filePath)
		if err != nil {
			return "", false, fmt.Errorf("创建文件失败: %w", err)
		}
		defer targetFile.Close()

		// 按顺序合并分片
		for i := 1; i <= totalChunks; i++ {
			// 查找对应分片
			var chunkToWrite ChunkInfo
			found := false
			for _, chunk := range fm.chunkData[fileName] {
				if chunk.ChunkNumber == i {
					chunkToWrite = chunk
					found = true
					break
				}
			}

			if !found {
				// 清理已创建的文件和分片数据
				targetFile.Close()
				os.Remove(filePath)
				return "", false, fmt.Errorf("分片丢失: %d", i)
			}

			// 写入分片数据
			if _, err := targetFile.Write(chunkToWrite.Data); err != nil {
				// 清理已创建的文件和分片数据
				targetFile.Close()
				os.Remove(filePath)
				return "", false, fmt.Errorf("写入分片数据失败: %w", err)
			}
		}

		// 验证文件是否为有效的ISO 9660格式
		if !fm.isValidISO(filePath) {
			// 不是有效ISO则删除文件
			targetFile.Close()
			os.Remove(filePath)
			// 更新进度为失败
			fm.updateUploadProgress(fileName, completedSize, totalSize, "failed")
			return "", false, errors.New("不是有效的ISO 9660格式文件")
		}

		// 更新进度为完成
		fm.updateUploadProgress(fileName, totalSize, totalSize, "completed")

		// 清理分片数据
		delete(fm.chunkData, fileName)
		delete(fm.chunkProgress, fileName)
		delete(fm.totalChunkMap, fileName)

		return filePath, true, nil
	}

	return "", false, nil
}

// extractFileNameFromURL 从URL中提取文件名
func extractFileNameFromURL(urlStr string) string {
	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// URL解析失败，直接返回空
		return ""
	}

	// 获取路径中的文件名
	path := parsedURL.Path

	// 检查路径是否以/结尾
	isPathEndsWithSlash := strings.HasSuffix(path, "/")

	fileName := filepath.Base(path)

	// 如果文件名包含查询参数，移除查询参数
	if idx := strings.Index(fileName, "?"); idx != -1 {
		fileName = fileName[:idx]
	}

	// 如果文件名包含哈希值，移除哈希值
	if idx := strings.Index(fileName, "#"); idx != -1 {
		fileName = fileName[:idx]
	}

	// 如果文件名为空、路径以/结尾，或者文件名为"."（当路径为"/"时），返回默认名称
	if fileName == "" || isPathEndsWithSlash || fileName == "." {
		fileName = "default.iso"
	}

	// 确保文件名以.iso结尾
	if !strings.HasSuffix(strings.ToLower(fileName), ".iso") {
		fileName += ".iso"
	}

	return fileName
}

// DownloadISOByURLWithProgress 根据URL下载ISO文件并跟踪进度
func (fm *LinuxFilesManager) DownloadISOByURLWithProgress(urlStr string) (string, error) {
	// 从URL中提取文件名作为默认值
	fileName := extractFileNameFromURL(urlStr)
	if fileName == "" {
		fileName = "default.iso"
	}

	// 验证文件名是否以.iso结尾
	if !strings.HasSuffix(strings.ToLower(fileName), ".iso") {
		return "", errors.New("仅支持ISO格式文件")
	}

	// 创建HTTP请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 发送请求获取响应头
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 从Content-Disposition头中提取文件名（如果有）
	disposition := resp.Header.Get("Content-Disposition")
	if disposition != "" {
		// 查找filename="xxx"或filename=xxx
		filenameStart := strings.Index(disposition, "filename=")
		if filenameStart != -1 {
			filenameStart += len("filename=")
			var extractedName string
			if disposition[filenameStart] == '"' {
				// 处理带引号的文件名
				filenameEnd := strings.Index(disposition[filenameStart+1:], "\"")
				if filenameEnd != -1 {
					extractedName = disposition[filenameStart+1 : filenameStart+1+filenameEnd]
				}
			} else {
				// 处理不带引号的文件名
				filenameEnd := strings.Index(disposition[filenameStart:], ";")
				if filenameEnd != -1 {
					extractedName = disposition[filenameStart : filenameStart+filenameEnd]
				} else {
					extractedName = disposition[filenameStart:]
				}
			}
			// 确保提取的文件名是有效的ISO文件名
			if extractedName != "" && strings.HasSuffix(strings.ToLower(extractedName), ".iso") {
				fileName = extractedName
			}
		}
	}

	// 构建文件绝对路径
	filePath := filepath.Join(fm.baseDir, fileName)

	// 获取文件大小
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		// 如果无法获取Content-Length，使用一个合理的默认值
		contentLength = 100 * 1024 * 1024 // 100MB
	}

	// 初始化下载进度（需要加锁）
	fm.mutex.Lock()
	fm.updateDownloadProgress(fileName, 0, contentLength, "downloading")
	fm.mutex.Unlock()

	// 创建目标文件
	targetFile, err := os.Create(filePath)
	if err != nil {
		// 更新进度为失败（需要加锁）
		fm.mutex.Lock()
		fm.updateDownloadProgress(fileName, 0, contentLength, "failed")
		fm.mutex.Unlock()
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer targetFile.Close()

	// 复制文件内容（支持大文件下载）
	const chunkSize = 8 * 1024 * 1024 // 8MB 分片大小
	buffer := make([]byte, chunkSize)
	var totalWritten int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// 写入文件内容
			written, writeErr := targetFile.Write(buffer[:n])
			if writeErr != nil {
				// 更新进度为失败（需要加锁）
				fm.mutex.Lock()
				fm.updateDownloadProgress(fileName, totalWritten, contentLength, "failed")
				fm.mutex.Unlock()
				// 删除已创建的文件
				os.Remove(filePath)
				return "", fmt.Errorf("写入文件内容失败: %w", writeErr)
			}
			totalWritten += int64(written)

			// 更新下载进度（需要加锁）
			fm.mutex.Lock()
			fm.updateDownloadProgress(fileName, totalWritten, contentLength, "downloading")
			fm.mutex.Unlock()
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			// 更新进度为失败（需要加锁）
			fm.mutex.Lock()
			fm.updateDownloadProgress(fileName, totalWritten, contentLength, "failed")
			fm.mutex.Unlock()
			// 删除已创建的文件
			os.Remove(filePath)
			return "", fmt.Errorf("读取响应内容失败: %w", err)
		}
	}

	// 更新实际文件大小（如果Content-Length不准确）
	if fi, err := targetFile.Stat(); err == nil {
		totalWritten = fi.Size()
	}

	// 验证文件大小是否有效（大于0字节）
	if totalWritten <= 0 {
		// 删除文件
		os.Remove(filePath)
		// 更新进度为失败（需要加锁）
		fm.mutex.Lock()
		fm.updateDownloadProgress(fileName, 0, contentLength, "failed")
		fm.mutex.Unlock()
		return "", errors.New("文件大小无效")
	}

	// 验证文件是否为有效的ISO 9660格式
	if !fm.isValidISO(filePath) {
		// 不是有效ISO则删除文件
		os.Remove(filePath)
		// 更新进度为失败（需要加锁）
		fm.mutex.Lock()
		fm.updateDownloadProgress(fileName, totalWritten, contentLength, "failed")
		fm.mutex.Unlock()
		return "", errors.New("不是有效的ISO 9660格式文件")
	}

	// 更新进度为完成（需要加锁）
	fm.mutex.Lock()
	fm.updateDownloadProgress(fileName, totalWritten, totalWritten, "completed")
	fm.mutex.Unlock()

	return filePath, nil
}

// GetISOList 获取ISO文件列表
func (fm *LinuxFilesManager) GetISOList() ([]FileInfo, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 读取目录内容
	entries, err := os.ReadDir(fm.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var fileList []FileInfo

	// 遍历目录内容
	for _, entry := range entries {
		// 跳过目录
		if entry.IsDir() {
			continue
		}

		// 获取文件信息
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 检查是否为ISO文件
		isISO := strings.HasSuffix(strings.ToLower(entry.Name()), ".iso")
		if isISO {
			// 构建文件信息
			fileInfo := FileInfo{
				Name:         entry.Name(),
				Path:         filepath.Join(fm.baseDir, entry.Name()),
				Size:         info.Size(),
				IsISO:        true,
				LastModified: info.ModTime().Format("2006-01-02 15:04:05"),
			}
			fileList = append(fileList, fileInfo)
		}
	}

	return fileList, nil
}

// GetISOAbsolutePath 获取ISO文件的绝对路径
func (fm *LinuxFilesManager) GetISOAbsolutePath(fileName string) (string, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 验证文件名是否以.iso结尾
	if !strings.HasSuffix(strings.ToLower(fileName), ".iso") {
		return "", errors.New("仅支持ISO格式文件")
	}

	// 构建文件绝对路径
	filePath := filepath.Join(fm.baseDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", fileName)
	}

	// 验证文件是否为有效的ISO 9660格式
	if !fm.isValidISO(filePath) {
		return "", errors.New("不是有效的ISO 9660格式文件")
	}

	return filePath, nil
}

// DeleteISO 删除ISO文件
func (fm *LinuxFilesManager) DeleteISO(fileName string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 验证文件名是否以.iso结尾
	if !strings.HasSuffix(strings.ToLower(fileName), ".iso") {
		return errors.New("仅支持删除ISO格式文件")
	}

	// 构建文件绝对路径
	filePath := filepath.Join(fm.baseDir, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", fileName)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// isValidISO 验证文件是否为有效的ISO 9660格式
// 简单验证：检查文件开头是否有ISO 9660签名
func (fm *LinuxFilesManager) isValidISO(filePath string) bool {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// 读取文件开头的2048字节（ISO 9660格式的第一个扇区）
	buffer := make([]byte, 2048)
	n, err := file.Read(buffer)
	if err != nil || n < 2048 {
		return false
	}

	// 检查ISO 9660签名
	// ISO 9660格式在偏移量32769（0x8001）处有"CD001"签名
	// 注意：在Go中，文件偏移是从0开始的，所以我们需要调整偏移量
	// 由于我们读取了前2048字节（0-2047），而签名在32769（0x8001）处，
	// 所以我们需要再读取后续内容或者使用Seek定位

	// 定位到ISO 9660签名位置
	if _, err := file.Seek(0x8000, 0); err != nil { // 0x8000 = 32768
		return false
	}

	// 读取6个字节，寻找"CD001"签名
	isoSignature := make([]byte, 6)
	if _, err := file.Read(isoSignature); err != nil {
		return false
	}

	// 检查是否包含ISO 9660签名
	return string(isoSignature[:5]) == "CD001"
}
