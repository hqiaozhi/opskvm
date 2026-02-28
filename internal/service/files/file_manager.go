package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/os/gfile"
	"golang.org/x/sys/unix"
)

const (
	FileStorageDir = "/data/opskvm/uploads"
)

type FileInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	RelPath   string `json:"rel_path"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type StorageInfo struct {
	TotalSpace     int64 `json:"total_space"`
	UsedSpace      int64 `json:"used_space"`
	AvailableSpace int64 `json:"available_space"`
	UsedPercent    int   `json:"used_percent"`
}

type IFileManagerService interface {
	ListFiles(path string, page, limit int) (int, []FileInfo, error)
	GetFileInfo(fileId string) (*FileInfo, error)
	DeleteFiles(paths string) (int, []string, error)
	DownloadFile(fileId string, w http.ResponseWriter) error
	GetFilePath(fileId string) string
	GetStorageDir() string
	GetStorageInfo() (*StorageInfo, error)
	CreateDirectory(path string) error
}

type FileManagerService struct {
	storageDir string
}

var _ IFileManagerService = (*FileManagerService)(nil)
var fileManagerService *FileManagerService
var onceFile sync.Once

func GetFileManagerService() *FileManagerService {
	onceFile.Do(func() {
		fileManagerService = &FileManagerService{
			storageDir: FileStorageDir,
		}
		fileManagerService.initDir()
	})
	return fileManagerService
}

func (s *FileManagerService) initDir() {
	if !gfile.Exists(s.storageDir) {
		if err := os.MkdirAll(s.storageDir, 0755); err != nil {
			fmt.Printf("创建文件存储目录失败: %s, err: %v\n", s.storageDir, err)
		}
	}
}

func (s *FileManagerService) normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	path = strings.TrimLeft(path, "/")
	return path
}

func (s *FileManagerService) validateAndCleanPath(targetPath string) (string, error) {
	targetPath = s.normalizePath(targetPath)
	if targetPath == "" {
		return "", nil
	}

	cleanPath := filepath.Clean(targetPath)
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("无效的路径")
	}

	parts := strings.Split(cleanPath, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return "", fmt.Errorf("不支持隐藏文件或目录")
		}
	}

	fullPath := filepath.Join(s.storageDir, cleanPath)
	absStorageDir, _ := filepath.Abs(s.storageDir)
	absFullPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFullPath, absStorageDir) {
		return "", fmt.Errorf("无效的路径")
	}

	return cleanPath, nil
}

func (s *FileManagerService) ListFiles(targetPath string, page, limit int) (int, []FileInfo, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	validatedPath, err := s.validateAndCleanPath(targetPath)
	if err != nil {
		return 0, nil, err
	}

	scanDir := s.storageDir
	if validatedPath != "" {
		scanDir = filepath.Join(s.storageDir, validatedPath)
		if !gfile.Exists(scanDir) {
			return 0, nil, fmt.Errorf("目录不存在: %s", validatedPath)
		}
	}

	var files []FileInfo
	err = filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(s.storageDir, path)
		if relPath == "." {
			return nil
		}

		if strings.HasPrefix(relPath, ".tmp") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		isRoot := validatedPath == ""
		if !isRoot && path == scanDir {
			return nil
		}

		files = append(files, FileInfo{
			Id:        info.Name(),
			Name:      info.Name(),
			Path:      path,
			RelPath:   relPath,
			Size:      info.Size(),
			IsDir:     info.IsDir(),
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			UpdatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		})

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	total := len(files)

	start := (page - 1) * limit
	end := start + limit
	if start > total {
		files = []FileInfo{}
	} else {
		if end > total {
			end = total
		}
		files = files[start:end]
	}

	return total, files, nil
}

func (s *FileManagerService) GetFileInfo(targetPath string) (*FileInfo, error) {
	validatedPath, err := s.validateAndCleanPath(targetPath)
	if err != nil {
		return nil, err
	}

	if validatedPath == "" {
		return &FileInfo{
			Name:      "root",
			Path:      s.storageDir,
			RelPath:   "",
			Size:      0,
			IsDir:     true,
			CreatedAt: "",
			UpdatedAt: "",
		}, nil
	}

	filePath := filepath.Join(s.storageDir, validatedPath)

	if !gfile.Exists(filePath) {
		return nil, fmt.Errorf("文件或目录不存在: %s", validatedPath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Id:        info.Name(),
		Name:      info.Name(),
		Path:      filePath,
		RelPath:   targetPath,
		Size:      info.Size(),
		IsDir:     info.IsDir(),
		CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		UpdatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *FileManagerService) DeleteFiles(paths string) (int, []string, error) {
	if paths == "" {
		return 0, nil, fmt.Errorf("路径列表不能为空")
	}

	pathList := strings.Split(paths, ",")
	var validatedPaths []string

	for _, targetPath := range pathList {
		targetPath = strings.TrimSpace(targetPath)
		validatedPath, err := s.validateAndCleanPath(targetPath)
		if err != nil {
			continue
		}

		if validatedPath != "" {
			validatedPaths = append(validatedPaths, validatedPath)
		}
	}

	if len(validatedPaths) == 0 {
		return 0, nil, fmt.Errorf("没有有效的路径")
	}

	deletedPaths := make(map[string]bool)
	deletedCount := 0
	var failedPaths []string

	for _, validatedPath := range validatedPaths {
		filePath := filepath.Join(s.storageDir, validatedPath)

		affected := false
		for deletedPath := range deletedPaths {
			if strings.HasPrefix(validatedPath, deletedPath+"/") || deletedPath == validatedPath {
				affected = true
				break
			}
		}
		if affected {
			continue
		}

		if !gfile.Exists(filePath) {
			continue
		}

		err := os.RemoveAll(filePath)
		if err != nil {
			fmt.Printf("[DeleteFiles] 删除失败: %s, err: %v\n", validatedPath, err)
			failedPaths = append(failedPaths, validatedPath)
			continue
		}

		deletedPaths[validatedPath] = true
		deletedCount++
		fmt.Printf("[DeleteFiles] path=%s\n", validatedPath)
	}

	return deletedCount, failedPaths, nil
}

func (s *FileManagerService) DownloadFile(targetPath string, w http.ResponseWriter) error {
	validatedPath, err := s.validateAndCleanPath(targetPath)
	if err != nil {
		return err
	}

	if validatedPath == "" {
		return fmt.Errorf("不能下载根目录")
	}

	filePath := filepath.Join(s.storageDir, validatedPath)

	if !gfile.Exists(filePath) {
		return fmt.Errorf("文件不存在: %s", validatedPath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("不能下载目录: %s", validatedPath)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(validatedPath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}

func (s *FileManagerService) GetFilePath(targetPath string) string {
	return filepath.Join(s.storageDir, targetPath)
}

func (s *FileManagerService) GetStorageDir() string {
	return s.storageDir
}

func (s *FileManagerService) GetStorageInfo() (*StorageInfo, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(s.storageDir, &stat)
	if err != nil {
		return nil, fmt.Errorf("获取存储信息失败: %v", err)
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available

	var usedPercent int
	if total > 0 {
		usedPercent = int(float64(used) / float64(total) * 100)
	}

	return &StorageInfo{
		TotalSpace:     total,
		UsedSpace:      used,
		AvailableSpace: available,
		UsedPercent:    usedPercent,
	}, nil
}

func (s *FileManagerService) CreateDirectory(targetPath string) error {
	validatedPath, err := s.validateAndCleanPath(targetPath)
	if err != nil {
		return err
	}

	if validatedPath == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	dirPath := filepath.Join(s.storageDir, validatedPath)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	fmt.Printf("[CreateDirectory] path=%s\n", targetPath)
	return nil
}
