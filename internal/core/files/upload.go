package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UploadISOWithChunk 分片上传ISO镜像文件
// 参数：fileName 文件名，chunkNumber 分片序号，totalChunks 总分片数，chunkData 分片数据
// 返回：文件绝对路径（当所有分片上传完成时），是否完成，错误信息
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