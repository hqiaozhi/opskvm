package mirrors

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/util/grand"
)

const (
	DefaultChunkSize = 5 * 1024 * 1024
	StorageDir       = "mirrors"
	StaticURL        = "/mirrors"
)

type FileMeta struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}

type UploadSession struct {
	UploadID    string        `json:"upload_id"`
	FileName    string        `json:"file_name" dc:"文件名"`
	FileSize    int64         `json:"file_size" dc:"文件大小"`
	ChunkSize   int64         `json:"chunk_size" dc:"分块大小"`
	ChunkCount  int           `json:"chunk_count" dc:"总分块数"`
	UploadedMap map[int]int64 `json:"uploaded_map" dc:"已上传分块"`
	FinalPath   string        `json:"final_path" dc:"最终文件路径"`
	File        *os.File      `json:"-" dc:"文件句柄"`
	CreatedAt   time.Time     `json:"created_at" dc:"创建时间"`
}

type IMirrorsChunkUploadService interface {
	InitUpload(fileMetas []FileMeta, chunkSize int64) ([]UploadSession, error)
	UploadChunk(ctx context.Context, uploadID string, chunkIndex int, chunkData []byte) (int64, error)
	CompleteUpload(ctx context.Context, uploadID string) (string, error)
	CancelUpload(ctx context.Context, uploadID string) error
	GetUploadStatus(uploadID string) (*UploadSession, error)
	GetSession(uploadID string) (*UploadSession, bool)
	RecordChunk(uploadID string, chunkIndex int, size int64)
	GetAllSessions() map[string]*UploadSession
}

type ChunkUploadService struct {
	sessions map[string]*UploadSession
	mu       sync.RWMutex
	rootPath string
	FinalDir string
}

var _ IMirrorsChunkUploadService = (*ChunkUploadService)(nil)
var chunkUploadService *ChunkUploadService
var onceChunk sync.Once

func GetMirrorsChunkUploadService(rootPath string) *ChunkUploadService {
	onceChunk.Do(func() {
		chunkUploadService = &ChunkUploadService{
			sessions: make(map[string]*UploadSession),
			rootPath: rootPath,
			FinalDir: filepath.Join(rootPath, StorageDir),
		}
		if err := os.MkdirAll(chunkUploadService.FinalDir, 0755); err != nil {
			fmt.Printf("创建目录失败: %s, err: %v\n", chunkUploadService.FinalDir, err)
		}
	})
	return chunkUploadService
}

func (s *ChunkUploadService) generateUploadID() string {
	return fmt.Sprintf("_%s_%d", grand.Letters(16), time.Now().Unix())
}

func (s *ChunkUploadService) validateFileName(fileName string) error {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return fmt.Errorf("文件名不能为空")
	}

	cleanName := filepath.Base(fileName)
	if strings.HasPrefix(cleanName, ".") {
		return fmt.Errorf("不支持隐藏文件")
	}

	if containsSpecialCharsAllowNoExt(cleanName) {
		return fmt.Errorf("文件名包含特殊字符: %s", cleanName)
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
		if err := s.validateFileName(meta.FileName); err != nil {
			continue
		}

		chunkCount := int((meta.Size + chunkSize - 1) / chunkSize)

		uploadID := s.generateUploadID()

		finalPath := filepath.Join(s.FinalDir, meta.FileName)
		dirPath := filepath.Dir(finalPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("创建父目录失败: %s, err: %v\n", dirPath, err)
		}

		finalFile, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("创建最终文件失败: %s, err: %v\n", finalPath, err)
			continue
		}

		session := UploadSession{
			UploadID:    uploadID,
			FileName:    meta.FileName,
			FileSize:    meta.Size,
			ChunkSize:   chunkSize,
			ChunkCount:  chunkCount,
			UploadedMap: make(map[int]int64),
			FinalPath:   finalPath,
			File:        finalFile,
			CreatedAt:   time.Now(),
		}

		s.sessions[uploadID] = &session

		fmt.Printf("[InitUpload] uploadId=%s, fileName=%s, size=%d, chunkCount=%d\n",
			uploadID, meta.FileName, meta.Size, chunkCount)

		sessions = append(sessions, session)
	}

	return sessions, nil
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

	if chunkIndex < 0 || chunkIndex >= session.ChunkCount {
		return 0, fmt.Errorf("分块索引无效: %d, 总分块数: %d", chunkIndex, session.ChunkCount)
	}

	offset := int64(chunkIndex) * session.ChunkSize

	s.mu.RLock()
	file := session.File
	s.mu.RUnlock()

	_, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf(" seek 失败: %v", err)
	}

	_, err = file.Write(chunkData)
	if err != nil {
		return 0, fmt.Errorf("写入分块数据失败: %v", err)
	}

	writtenSize := int64(len(chunkData))

	s.mu.Lock()
	session.UploadedMap[chunkIndex] = writtenSize
	s.mu.Unlock()

	fmt.Printf("[UploadChunk] uploadId=%s, chunkIndex=%d, size=%d, offset=%d\n", uploadID, chunkIndex, writtenSize, offset)

	return writtenSize, nil
}

func (s *ChunkUploadService) CompleteUpload(ctx context.Context, uploadID string) (string, error) {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	if len(session.UploadedMap) != session.ChunkCount {
		return "", fmt.Errorf("分块未全部上传: 已上传 %d/%d", len(session.UploadedMap), session.ChunkCount)
	}

	s.mu.Lock()
	session.File.Close()
	session.File = nil
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	fmt.Printf("[CompleteUpload] 完成, uploadId=%s, finalPath=%s\n", uploadID, session.FinalPath)

	return session.FinalPath, nil
}

func (s *ChunkUploadService) CancelUpload(ctx context.Context, uploadID string) error {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	if session.File != nil {
		session.File.Close()
	}

	s.mu.Lock()
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	if session.FinalPath != "" {
		os.Remove(session.FinalPath)
	}

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
