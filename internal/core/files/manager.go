package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewLinuxFilesManager 创建文件管理器实例
// 参数：baseDir ISO文件存储的基础目录
// 返回：文件管理器实例，错误信息
func New(baseDir string) (*LinuxFilesManager, error) {
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
