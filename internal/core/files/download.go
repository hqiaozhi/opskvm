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
)

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
// 参数：url 文件下载链接
// 返回：文件绝对路径，错误信息
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
