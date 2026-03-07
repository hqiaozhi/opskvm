package files

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"
)

const (
	DefaultChunkSize = 5 * 1024 * 1024
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
	FinalPath   string        `json:"final_path" dc:"最终文件路径"`
	File        *os.File      `json:"-" dc:"文件句柄"`
	CreatedAt   time.Time     `json:"created_at" dc:"创建时间"`
	FileMd5     string        `json:"file_md5" dc:"文件SHA256校验值"`
	SkipUpload  bool          `json:"skip_upload" dc:"是否跳过上传（秒传）"`
}

type IChunkUploadService interface {
	InitUpload(fileMetas []FileMeta, chunkSize int64) ([]UploadSession, error)
	UploadChunk(ctx context.Context, uploadID string, chunkIndex int, chunkData []byte) (int64, error)
	CompleteUpload(ctx context.Context, uploadID string) (string, error)
	CompleteAllUploads(ctx context.Context) (int, []UploadSession, error)
	CancelUpload(ctx context.Context, uploadID string) error
	GetUploadStatus(uploadID string) (*UploadSession, error)
	GetSession(uploadID string) (*UploadSession, bool)
	RecordChunk(uploadID string, chunkIndex int, size int64)
	GetAllSessions() map[string]*UploadSession
}

type FileMeta struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Md5  string `json:"md5"`
}

type ChunkUploadService struct {
	sessions       map[string]*UploadSession
	mu             sync.RWMutex
	rootPath       string
	UploadFinalDir string
}

var _ IChunkUploadService = (*ChunkUploadService)(nil)
var chunkUploadService *ChunkUploadService
var onceChunk sync.Once

func GetChunkUploadService(rootPath string) *ChunkUploadService {
	onceChunk.Do(func() {
		chunkUploadService = &ChunkUploadService{
			sessions:       make(map[string]*UploadSession),
			rootPath:       rootPath,
			UploadFinalDir: filepath.Join(rootPath, "uploads"),
		}
		if err := os.MkdirAll(chunkUploadService.UploadFinalDir, 0755); err != nil {
			g.Log().Errorf(context.Background(), "创建目录失败: %s, err: %v", chunkUploadService.UploadFinalDir, err)
		}
	})
	return chunkUploadService
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
		if containsSpecialChars(part) {
			return fmt.Errorf("文件名包含特殊字符: %s", part)
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

		finalPath := filepath.Join(s.UploadFinalDir, meta.Path)
		dirPath := filepath.Dir(finalPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			g.Log().Errorf(context.Background(), "创建父目录失败: %s, err: %v", dirPath, err)
		}

		skipUpload := false
		var finalFile *os.File
		if !isDir {
			if meta.Md5 != "" {
				if existsInfo, err := os.Stat(finalPath); err == nil {
					if existsInfo.Size() == meta.Size {
						existingSha256, err := s.calculateFileSha256(finalPath)
						if err == nil && existingSha256 == meta.Md5 {
							skipUpload = true
							g.Log().Infof(context.Background(), "[InitUpload] 秒传: uploadId=%s, path=%s, size=%d, sha256=%s",
								uploadID, meta.Path, meta.Size, meta.Md5)
						}
					} else {
						os.Remove(finalPath)
						g.Log().Infof(context.Background(), "[InitUpload] 文件大小不匹配，删除旧文件: uploadId=%s, path=%s, oldSize=%d, newSize=%d",
							uploadID, meta.Path, existsInfo.Size(), meta.Size)
					}
				}
			}

			if !skipUpload {
				f, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					g.Log().Errorf(context.Background(), "创建最终文件失败: %s, err: %v", finalPath, err)
					continue
				}
				finalFile = f
			}
		}

		session := UploadSession{
			UploadID:    uploadID,
			Path:        meta.Path,
			FileName:    fileName,
			FileSize:    meta.Size,
			ChunkSize:   chunkSize,
			ChunkCount:  chunkCount,
			IsDir:       isDir,
			UploadedMap: make(map[int]int64),
			FinalPath:   finalPath,
			File:        finalFile,
			CreatedAt:   time.Now(),
			FileMd5:     meta.Md5,
			SkipUpload:  skipUpload,
		}

		s.sessions[uploadID] = &session
		sessions = append(sessions, session)

		if !skipUpload {
			g.Log().Infof(context.Background(), "[InitUpload] uploadId=%s, path=%s, size=%d, chunkCount=%d, isDir=%v",
				uploadID, meta.Path, meta.Size, chunkCount, isDir)
		}
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

	if session.IsDir {
		return 0, fmt.Errorf("目录不需要上传分块: %s", uploadID)
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
		return 0, fmt.Errorf("seek 失败: %v", err)
	}

	_, err = file.Write(chunkData)
	if err != nil {
		return 0, fmt.Errorf("写入分块数据失败: %v", err)
	}

	writtenSize := int64(len(chunkData))

	s.RecordChunk(uploadID, chunkIndex, writtenSize)

	g.Log().Debugf(context.Background(), "[UploadChunk] uploadId=%s, chunkIndex=%d, size=%d, offset=%d", uploadID, chunkIndex, writtenSize, offset)

	return writtenSize, nil
}

func (s *ChunkUploadService) calculateFileSha256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s *ChunkUploadService) CompleteUpload(ctx context.Context, uploadID string) (string, error) {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	if session.SkipUpload {
		s.mu.Lock()
		delete(s.sessions, uploadID)
		s.mu.Unlock()
		g.Log().Infof(context.Background(), "[CompleteUpload] 秒传完成, uploadId=%s, finalPath=%s", uploadID, session.FinalPath)
		return session.FinalPath, nil
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

	filePath := session.FinalPath

	calculatedSha256, err := s.calculateFileSha256(filePath)
	if err != nil {
		session.File.Close()
		session.File = nil
		os.Remove(filePath)
		s.mu.Lock()
		delete(s.sessions, uploadID)
		s.mu.Unlock()
		return "", fmt.Errorf("计算文件SHA256失败: %v", err)
	}

	if calculatedSha256 != session.FileMd5 {
		session.File.Close()
		session.File = nil
		os.Remove(filePath)
		s.mu.Lock()
		delete(s.sessions, uploadID)
		s.mu.Unlock()
		g.Log().Errorf(context.Background(), "[CompleteUpload] SHA256不一致，上传失败！uploadId=%s, expectedSha256=%s, calculatedSha256=%s",
			uploadID, session.FileMd5, calculatedSha256)
		return "", fmt.Errorf("SHA256不一致，上传失败！")
	}

	s.mu.Lock()
	session.File.Close()
	session.File = nil
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	g.Log().Infof(context.Background(), "[CompleteUpload] 完成, uploadId=%s, finalPath=%s, sha256=%s", uploadID, filePath, calculatedSha256)

	return filePath, nil
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
			g.Log().Errorf(context.Background(), "[CompleteAllUploads] failed: uploadId=%s, err=%v", session.UploadID, err)
			continue
		}
		completedSessions = append(completedSessions, session)
		completedCount++
	}

	return completedCount, completedSessions, nil
}

func (s *ChunkUploadService) CancelUpload(ctx context.Context, uploadID string) error {
	s.mu.RLock()
	session, ok := s.sessions[uploadID]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("上传会话不存在: %s", uploadID)
	}

	g.Log().Infof(context.Background(), "[CancelUpload] 开始取消上传: uploadId=%s, finalPath=%s, skipUpload=%v, file=nil(%v)",
		uploadID, session.FinalPath, session.SkipUpload, session.File == nil)

	if session.File != nil {
		session.File.Close()
	}

	s.mu.Lock()
	delete(s.sessions, uploadID)
	s.mu.Unlock()

	if session.FinalPath != "" {
		if err := os.Remove(session.FinalPath); err == nil {
			g.Log().Infof(context.Background(), "[CancelUpload] 已删除文件: uploadId=%s, path=%s", uploadID, session.FinalPath)
		} else {
			g.Log().Errorf(context.Background(), "[CancelUpload] 删除文件失败: uploadId=%s, path=%s, err=%v", uploadID, session.FinalPath, err)
		}
	}

	g.Log().Infof(context.Background(), "[CancelUpload] 完成: uploadId=%s", uploadID)

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
