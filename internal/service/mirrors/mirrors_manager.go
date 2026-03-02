package mirrors

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gfile"
)

const (
	StatusDownloading = "downloading"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
)

type Info struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UrlUploadTask struct {
	UploadId  string    `json:"upload_id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	LocalPath string    `json:"local_path"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Error     string    `json:"error"`
}

type IMirrorsManagerService interface {
	List(path string, page, limit int) (int, []Info, error)
	Delete(ids string) (int, []string, error)
	UploadByUrl(url string, fileName string) (*UrlUploadTask, error)
	GetUrlUploadStatus(uploadId string) (*UrlUploadTask, error)
	GetStorageDir() string
}

type MirrorsManagerService struct {
	StorageDir string
	staticURL  string
	urlTasks   map[string]*UrlUploadTask
	mu         sync.RWMutex
}

var _ IMirrorsManagerService = (*MirrorsManagerService)(nil)
var mirrorsManagerService *MirrorsManagerService
var onceMirrorsManager sync.Once

func GetMirrorsManagerService(rootPath string) *MirrorsManagerService {
	onceMirrorsManager.Do(func() {
		mirrorsManagerService = &MirrorsManagerService{
			StorageDir: filepath.Join(rootPath, StorageDir),
			staticURL:  StaticURL,
			urlTasks:   make(map[string]*UrlUploadTask),
		}
		mirrorsManagerService.initDir()
	})
	return mirrorsManagerService
}

func (s *MirrorsManagerService) initDir() {
	if !gfile.Exists(s.StorageDir) {
		if err := os.MkdirAll(s.StorageDir, 0755); err != nil {
			fmt.Printf("创建存储目录失败: %s, err: %v\n", s.StorageDir, err)
		}
	}
}

func (s *MirrorsManagerService) normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	path = strings.TrimLeft(path, "/")
	return path
}

func (s *MirrorsManagerService) validateAndCleanPath(targetPath string) (string, error) {
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
		if containsSpecialCharsAllowNoExt(part) {
			return "", fmt.Errorf("文件名包含特殊字符: %s", part)
		}
	}

	fullPath := filepath.Join(s.StorageDir, cleanPath)
	absStorageDir, _ := filepath.Abs(s.StorageDir)
	absFullPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFullPath, absStorageDir) {
		return "", fmt.Errorf("无效的路径")
	}

	return cleanPath, nil
}

func containsSpecialChars(s string) bool {
	invalidChars := `!@#$%^&*()+={}[]|\\:;"'<>,?`
	for _, c := range s {
		if strings.ContainsRune(invalidChars, c) {
			return true
		}
	}
	return false
}

func containsSpecialCharsAllowNoExt(s string) bool {
	invalidChars := `!#$%^+={}[]|\\:;"'<>,?/`
	for _, c := range s {
		if strings.ContainsRune(invalidChars, c) {
			return true
		}
	}
	return false
}

func (s *MirrorsManagerService) extractFileNameFromURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		fileName := parts[len(parts)-1]
		if strings.Contains(fileName, ".") {
			return fileName
		}
	}
	return ""
}

func (s *MirrorsManagerService) List(targetPath string, page, limit int) (int, []Info, error) {
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

	scanDir := s.StorageDir
	if validatedPath != "" {
		scanDir = filepath.Join(s.StorageDir, validatedPath)
		if !gfile.Exists(scanDir) {
			return 0, nil, fmt.Errorf("目录不存在: %s", validatedPath)
		}
	}

	var List []Info
	err = filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(s.StorageDir, path)
		if relPath == "." {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		isRoot := validatedPath == ""
		if !isRoot && path == scanDir {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		fileURL := s.staticURL + "/" + relPath

		List = append(List, Info{
			Id:        info.Name(),
			Name:      info.Name(),
			Path:      path,
			Size:      info.Size(),
			URL:       fileURL,
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			UpdatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		})

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	total := len(List)

	start := (page - 1) * limit
	end := start + limit
	if start > total {
		List = []Info{}
	} else {
		if end > total {
			end = total
		}
		List = List[start:end]
	}

	return total, List, nil
}

func (s *MirrorsManagerService) Delete(ids string) (int, []string, error) {
	if ids == "" {
		return 0, nil, fmt.Errorf("ID列表不能为空")
	}

	idList := strings.Split(ids, ",")
	var validatedPaths []string

	for _, id := range idList {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		validatedPath, err := s.validateAndCleanPath(id)
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
		filePath := filepath.Join(s.StorageDir, validatedPath)

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
			failedPaths = append(failedPaths, validatedPath)
			continue
		}

		deletedPaths[validatedPath] = true
		deletedCount++
	}

	return deletedCount, failedPaths, nil
}

func (s *MirrorsManagerService) UploadByUrl(url string, fileName string) (*UrlUploadTask, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	if fileName == "" {
		fileName = s.extractFileNameFromURL(url)
		if fileName == "" {
			return nil, fmt.Errorf("无法从URL中提取文件名")
		}
	}

	task := &UrlUploadTask{
		UploadId:  fmt.Sprintf("url_%d", time.Now().Unix()),
		FileName:  fileName,
		Status:    StatusDownloading,
		URL:       url,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.urlTasks[task.UploadId] = task
	s.mu.Unlock()

	go s.downloadFromUrl(task)

	return task, nil
}

func (s *MirrorsManagerService) downloadFromUrl(task *UrlUploadTask) {
	defer func() {
		if r := recover(); r != nil {
			task.Status = StatusFailed
			task.Error = fmt.Sprintf("下载失败: %v", r)
			fmt.Printf("[UploadByUrl] failed: uploadId=%s, error=%v\n", task.UploadId, r)
		}
	}()

	req, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("创建请求失败: %v", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.google.com/")

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("下载失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("HTTP错误: %d", resp.StatusCode)
		return
	}

	task.FileSize = resp.ContentLength

	finalPath := filepath.Join(s.StorageDir, task.FileName)
	if gfile.Exists(finalPath) {
		if err := os.Remove(finalPath); err != nil {
			task.Status = StatusFailed
			task.Error = fmt.Sprintf("删除已存在文件失败: %v", err)
			return
		}
	}

	file, err := os.Create(finalPath)
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("创建文件失败: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		task.Status = StatusFailed
		task.Error = fmt.Sprintf("写入文件失败: %v", err)
		os.Remove(finalPath)
		return
	}

	info, _ := os.Stat(finalPath)
	task.FileSize = info.Size()
	task.LocalPath = finalPath
	task.Status = StatusCompleted

	fmt.Printf("[UploadByUrl] completed: uploadId=%s, fileName=%s, size=%d\n",
		task.UploadId, task.FileName, task.FileSize)
}

func (s *MirrorsManagerService) GetUrlUploadStatus(uploadId string) (*UrlUploadTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.urlTasks[uploadId]
	if !ok {
		return nil, fmt.Errorf("上传任务不存在: %s", uploadId)
	}

	return task, nil
}

func (s *MirrorsManagerService) GetStorageDir() string {
	return s.StorageDir
}
