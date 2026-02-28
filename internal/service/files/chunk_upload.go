package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/grand"
)

const (
	DefaultChunkSize = 5 * 1024 * 1024
	UploadTempDir    = "/data/opskvm/uploads/.tmp"
	UploadFinalDir   = "/data/opskvm/uploads"
)

type UploadSession struct {
	UploadID    string        `json:"upload_id"`
	Path        string        `json:"path" dc:"文件路径（含目录）"`
	FileName    string        `json:"file_name" dc:"文件名"`
	FileSize    int64         `json:"file_size" dc:"文件大小"`
	ChunkSize   int64         `json:"chunk_size" dc:"分块大小"`
	ChunkCount  int           `json:"chunk_count" dc:"总分块数"`
	IsDir       bool          `json:"is_dir" dc:"是否为目录"`
	UploadedMap map[int]int64 `json:"uploaded_map" dc:"已上传分块"`
	CreatedAt   time.Time     `json:"created_at" dc:"创建时间"`
}

type IChunkUploadService interface {
	InitUpload(fileMetas []FileMeta, chunkSize int64) ([]UploadSession, error)
	UploadChunk(ctx context.Context, uploadID string, chunkIndex int, chunkData []byte) (int64, error)
	CompleteUpload(ctx context.Context, uploadID string) (string, error)
	CompleteAllUploads(ctx context.Context) (int, []UploadSession, error)
	CancelUpload(ctx context.Context, uploadID string) error
	GetUploadStatus(uploadID string) (*UploadSession, error)
	GetSession(uploadID string) (*UploadSession, bool)
	GetUploadDir(uploadID string) string
	RecordChunk(uploadID string, chunkIndex int, size int64)
	GetAllSessions() map[string]*UploadSession
}

type FileMeta struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type ChunkUploadService struct {
	sessions map[string]*UploadSession
	mu       sync.RWMutex
}

var _ IChunkUploadService = (*ChunkUploadService)(nil)
var chunkUploadService *ChunkUploadService
var onceChunk sync.Once

func GetChunkUploadService() *ChunkUploadService {
	onceChunk.Do(func() {
		chunkUploadService = &ChunkUploadService{
			sessions: make(map[string]*UploadSession),
		}
		chunkUploadService.initDirs()
	})
	return chunkUploadService
}

func (s *ChunkUploadService) initDirs() {
	for _, dir := range []string{UploadTempDir, UploadFinalDir} {
		if !gfile.Exists(dir) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("创建上传目录失败: %s, err: %v\n", dir, err)
			}
		}
	}
}

func (s *ChunkUploadService) generateUploadID() string {
	return fmt.Sprintf("%s_%d", grand.Letters(16), time.Now().Unix())
}

func (s *ChunkUploadService) validatePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("无效的路径")
	}

	parts := strings.Split(cleanPath, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return fmt.Errorf("不支持隐藏文件或目录")
		}
	}

	return nil
}

func (s *ChunkUploadService) InitUpload(fileMetas []FileMeta, chunkSize int64) ([]UploadSession, error) {
	if len(fileMetas) == 0 {
		return nil, fmt.Errorf("文件列表不能为空")
	}

	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	var sessions []UploadSession

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, meta := range fileMetas {
		if err := s.validatePath(meta.Path); err != nil {
			continue
		}

		var chunkCount int
		var isDir bool

		if meta.Size == 0 {
			isDir = true
			chunkCount = 0
		} else {
			chunkCount = int((meta.Size + chunkSize - 1) / chunkSize)
		}

		uploadID := s.generateUploadID()
		fileName := filepath.Base(meta.Path)

		session := UploadSession{
			UploadID:    uploadID,
			Path:        meta.Path,
			FileName:    fileName,
			FileSize:    meta.Size,
			ChunkSize:   chunkSize,
			ChunkCount:  chunkCount,
			IsDir:       isDir,
			UploadedMap: make(map[int]int64),
			CreatedAt:   time.Now(),
		}

		s.sessions[uploadID] = &session
		sessions = append(sessions, session)

		if isDir {
			dirPath := filepath.Join(UploadFinalDir, meta.Path)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				fmt.Printf("创建目录失败: %s, err: %v\n", dirPath, err)
			}
		} else {
			dirPath := filepath.Dir(filepath.Join(UploadFinalDir, meta.Path))
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				fmt.Printf("创建父目录失败: %s, err: %v\n", dirPath, err)
			}
		}

		fmt.Printf("[InitUpload] uploadId=%s, path=%s, size=%d, chunkCount=%d, isDir=%v\n",
			uploadID, meta.Path, meta.Size, chunkCount, isDir)
	}

	return sessions, nil
}

func (s *ChunkUploadService) GetUploadDir(uploadID string) string {
	return filepath.Join(UploadTempDir, uploadID)
}

func (s *ChunkUploadService) RecordChunk(uploadID string, chunkIndex int, size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.sessions[uploadID]; ok {
		session.UploadedMap[chunkIndex] = size
	}
}

func (s *ChunkUploadService) UploadChunk(ctx context.Context, uploadID string, chunkIndex int, chunkData []byte) (int64, error) {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	if session.IsDir {
		return 0, fmt.Errorf("目录不需要上传分块: %s", uploadID)
	}

	if chunkIndex < 0 || chunkIndex >= session.ChunkCount {
		return 0, fmt.Errorf("分块索引无效: %d, 总分块数: %d", chunkIndex, session.ChunkCount)
	}

	uploadDir := s.GetUploadDir(uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return 0, fmt.Errorf("创建上传目录失败: %v", err)
	}

	chunkPath := filepath.Join(uploadDir, fmt.Sprintf("chunk_%05d", chunkIndex))

	file, err := os.Create(chunkPath)
	if err != nil {
		return 0, fmt.Errorf("创建分块文件失败: %v", err)
	}
	defer file.Close()

	_, err = file.Write(chunkData)
	if err != nil {
		return 0, fmt.Errorf("写入分块数据失败: %v", err)
	}

	writtenSize := int64(len(chunkData))

	s.mu.Lock()
	session.UploadedMap[chunkIndex] = writtenSize
	s.mu.Unlock()

	fmt.Printf("[UploadChunk] uploadId=%s, chunkIndex=%d, size=%d\n", uploadID, chunkIndex, writtenSize)

	return writtenSize, nil
}

func (s *ChunkUploadService) CompleteUpload(ctx context.Context, uploadID string) (string, error) {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	if session.IsDir {
		s.mu.Lock()
		delete(s.sessions, uploadID)
		s.mu.Unlock()
		return session.Path, nil
	}

	if len(session.UploadedMap) != session.ChunkCount {
		return "", fmt.Errorf("分块未全部上传: 已上传 %d/%d", len(session.UploadedMap), session.ChunkCount)
	}

	uploadDir := s.GetUploadDir(uploadID)
	finalDir := filepath.Dir(filepath.Join(UploadFinalDir, session.Path))
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", fmt.Errorf("创建最终目录失败: %v", err)
	}

	finalPath := filepath.Join(UploadFinalDir, session.Path)

	if gfile.Exists(finalPath) {
		info, err := os.Stat(finalPath)
		if err == nil && info.IsDir() {
			if err := os.RemoveAll(finalPath); err != nil {
				return "", fmt.Errorf("删除已存在的目录失败: %v", err)
			}
		}
	}

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("创建最终文件失败: %v", err)
	}
	defer finalFile.Close()

	for i := 0; i < session.ChunkCount; i++ {
		chunkPath := filepath.Join(uploadDir, fmt.Sprintf("chunk_%05d", i))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return "", fmt.Errorf("打开分块文件失败: %v", err)
		}

		_, err = io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return "", fmt.Errorf("合并分块文件失败: %v", err)
		}
	}

	s.mu.Lock()
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	os.RemoveAll(uploadDir)

	fmt.Printf("[CompleteUpload] uploadId=%s, finalPath=%s\n", uploadID, finalPath)

	return finalPath, nil
}

func (s *ChunkUploadService) CompleteAllUploads(ctx context.Context) (int, []UploadSession, error) {
	s.mu.RLock()
	var pendingSessions []UploadSession
	for _, session := range s.sessions {
		pendingSessions = append(pendingSessions, *session)
	}
	s.mu.RUnlock()

	if len(pendingSessions) == 0 {
		return 0, nil, nil
	}

	var completedSessions []UploadSession
	completedCount := 0

	for _, session := range pendingSessions {
		_, err := s.CompleteUpload(ctx, session.UploadID)
		if err != nil {
			fmt.Printf("[CompleteAllUploads] failed: uploadId=%s, err=%v\n", session.UploadID, err)
			continue
		}
		completedSessions = append(completedSessions, session)
		completedCount++
	}

	return completedCount, completedSessions, nil
}

func (s *ChunkUploadService) CancelUpload(ctx context.Context, uploadID string) error {
	s.mu.RLock()
	_, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	uploadDir := s.GetUploadDir(uploadID)
	os.RemoveAll(uploadDir)

	s.mu.Lock()
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	fmt.Printf("[CancelUpload] uploadId=%s\n", uploadID)

	return nil
}

func (s *ChunkUploadService) GetUploadStatus(uploadID string) (*UploadSession, error) {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	return session, nil
}

func (s *ChunkUploadService) GetSession(uploadID string) (*UploadSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[uploadID]
	return session, ok
}

func (s *ChunkUploadService) GetAllSessions() map[string]*UploadSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions
}
