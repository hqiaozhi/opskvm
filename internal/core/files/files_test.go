package files

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 创建一个简单的ISO 9660格式的测试文件
func createTestISOFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 填充一些数据
	buffer := make([]byte, 2048)
	if _, err := file.Write(buffer); err != nil {
		return err
	}

	// 写入ISO 9660签名（在偏移量0x8000处）
	if _, err := file.Seek(0x8000, 0); err != nil {
		return err
	}

	// 写入ISO 9660签名 "CD001"
	signature := []byte("CD001")
	if _, err := file.Write(signature); err != nil {
		return err
	}

	return nil
}

func TestNew(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试创建新的文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 验证基础目录是否正确
	if fm == nil {
		t.Fatal("文件管理器为nil")
	}
}

func TestUploadISOWithChunk(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 测试上传非ISO文件
	_, _, err = fm.UploadISOWithChunk("test.txt", 1, 1, []byte("test"))
	if err == nil || !strings.Contains(err.Error(), "仅支持ISO格式文件") {
		t.Errorf("上传非ISO文件应该失败，但实际结果: %v", err)
	}

	// 测试上传有效ISO文件（使用单个分片）
	isoContent := make([]byte, 0x8005) // 足够容纳ISO签名
	copy(isoContent[0x8000:0x8005], []byte("CD001"))

	filePath, completed, err := fm.UploadISOWithChunk("test.iso", 1, 1, isoContent)
	if err != nil {
		t.Fatalf("上传ISO文件失败: %v", err)
	}

	// 验证上传是否完成
	if !completed {
		t.Error("单分片上传应该完成")
	}

	// 验证文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("上传的文件不存在: %s", filePath)
	}

	// 验证文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if fileInfo.Size() != int64(len(isoContent)) {
		t.Errorf("文件大小不一致，预期: %d, 实际: %d", len(isoContent), fileInfo.Size())
	}

	// 测试多分片上传
	multiChunkContent := make([]byte, 0x10005)              // 更大的内容
	copy(multiChunkContent[0x8000:0x8005], []byte("CD001")) // ISO签名

	totalChunks := 3
	chunkSize := len(multiChunkContent) / totalChunks

	var lastFilePath string
	var lastCompleted bool

	for i := 1; i <= totalChunks; i++ {
		start := (i - 1) * chunkSize
		end := start + chunkSize

		// 最后一个分片包含剩余的所有数据
		if i == totalChunks {
			end = len(multiChunkContent)
		}

		chunkData := multiChunkContent[start:end]
		filePath, completed, err := fm.UploadISOWithChunk("multi.iso", i, totalChunks, chunkData)
		if err != nil {
			t.Fatalf("上传分片 %d 失败: %v", i, err)
		}

		lastFilePath = filePath
		lastCompleted = completed

		// 检查上传进度
		progress, err := fm.GetUploadProgress("multi.iso")
		if err != nil {
			t.Errorf("获取上传进度失败: %v", err)
		} else {
			// 验证进度信息
			if progress.FileName != "multi.iso" {
				t.Errorf("进度信息中的文件名不正确，预期: multi.iso, 实际: %s", progress.FileName)
			}
			if progress.Status != "uploading" && i < totalChunks {
				t.Errorf("上传中的文件状态应该是 'uploading'，但实际是: %s", progress.Status)
			}
		}
	}

	// 验证最后一个分片上传后应该完成
	if !lastCompleted {
		t.Error("多分片上传完成后应该返回completed=true")
	}

	// 验证文件是否存在
	if _, err := os.Stat(lastFilePath); os.IsNotExist(err) {
		t.Fatalf("多分片上传的文件不存在: %s", lastFilePath)
	}

	// 验证文件大小
	fileInfo, err = os.Stat(lastFilePath)
	if err != nil {
		t.Fatalf("获取多分片上传文件信息失败: %v", err)
	}

	if fileInfo.Size() != int64(len(multiChunkContent)) {
		t.Errorf("多分片上传文件大小不一致，预期: %d, 实际: %d", len(multiChunkContent), fileInfo.Size())
	}

	// 验证最终上传进度状态为completed
	progress, err := fm.GetUploadProgress("multi.iso")
	if err != nil {
		t.Errorf("获取上传进度失败: %v", err)
	} else {
		if progress.Status != "completed" {
			t.Errorf("上传完成后状态应该是 'completed'，但实际是: %s", progress.Status)
		}
		if progress.Progress != 100 {
			t.Errorf("上传完成后进度应该是 100%%，但实际是: %d%%", progress.Progress)
		}
	}
}

func TestGetISOList(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 创建测试文件
	createTestISOFile(filepath.Join(tempDir, "test1.iso"))
	createTestISOFile(filepath.Join(tempDir, "test2.iso"))

	// 创建非ISO文件
	nonISOFile, _ := os.Create(filepath.Join(tempDir, "test.txt"))
	nonISOFile.Write([]byte("test"))
	nonISOFile.Close()

	// 获取ISO文件列表
	fileList, err := fm.GetISOList()
	if err != nil {
		t.Fatalf("获取ISO文件列表失败: %v", err)
	}

	// 验证列表长度（应该只有2个ISO文件）
	if len(fileList) != 2 {
		t.Errorf("ISO文件列表长度不正确，预期: 2, 实际: %d", len(fileList))
	}

	// 验证文件名
	fileNames := make(map[string]bool)
	for _, file := range fileList {
		fileNames[file.Name] = true
		// 验证IsISO字段
		if !file.IsISO {
			t.Errorf("文件 %s 的IsISO字段应该为true", file.Name)
		}
	}

	if !fileNames["test1.iso"] {
		t.Error("列表中应该包含test1.iso文件")
	}

	if !fileNames["test2.iso"] {
		t.Error("列表中应该包含test2.iso文件")
	}
}

func TestGetISOAbsolutePath(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 创建测试文件
	testISOFile := "test.iso"
	createTestISOFile(filepath.Join(tempDir, testISOFile))

	// 测试获取存在的ISO文件的绝对路径
	absPath, err := fm.GetISOAbsolutePath(testISOFile)
	if err != nil {
		t.Fatalf("获取ISO文件绝对路径失败: %v", err)
	}

	// 验证绝对路径
	expectedPath := filepath.Join(tempDir, testISOFile)
	if absPath != expectedPath {
		t.Errorf("绝对路径不正确，预期: %s, 实际: %s", expectedPath, absPath)
	}

	// 测试获取不存在的ISO文件的绝对路径
	_, err = fm.GetISOAbsolutePath("non_existent.iso")
	if err == nil {
		t.Error("获取不存在的ISO文件的绝对路径应该失败")
	}

	// 测试获取非ISO文件的绝对路径
	_, err = fm.GetISOAbsolutePath("test.txt")
	if err == nil || !strings.Contains(err.Error(), "仅支持ISO格式文件") {
		t.Errorf("获取非ISO文件的绝对路径应该失败，但实际结果: %v", err)
	}
}

func TestDeleteISO(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 创建测试文件
	testISOFile := "test.iso"
	createTestISOFile(filepath.Join(tempDir, testISOFile))

	// 验证文件是否存在
	if _, err := os.Stat(filepath.Join(tempDir, testISOFile)); os.IsNotExist(err) {
		t.Fatalf("测试文件不存在: %s", testISOFile)
	}

	// 测试删除ISO文件
	if err := fm.DeleteISO(testISOFile); err != nil {
		t.Fatalf("删除ISO文件失败: %v", err)
	}

	// 验证文件是否被删除
	if _, err := os.Stat(filepath.Join(tempDir, testISOFile)); !os.IsNotExist(err) {
		t.Fatalf("ISO文件应该被删除，但仍然存在")
	}

	// 测试删除不存在的ISO文件
	if err := fm.DeleteISO(testISOFile); err == nil {
		t.Error("删除不存在的ISO文件应该失败")
	}

	// 测试删除非ISO文件
	if err := fm.DeleteISO("test.txt"); err == nil || !strings.Contains(err.Error(), "仅支持删除ISO格式文件") {
		t.Errorf("删除非ISO文件应该失败，但实际结果: %v", err)
	}
}

func TestExtractFileNameFromURL(t *testing.T) {
	// 测试URL包含明确的文件名
	url1 := "https://example.com/files/test.iso"
	name1 := extractFileNameFromURL(url1)
	if name1 != "test.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.iso, 实际提取到: %s", url1, name1)
	}

	// 测试URL包含查询参数
	url2 := "https://example.com/files/test.iso?param=value"
	name2 := extractFileNameFromURL(url2)
	if name2 != "test.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.iso, 实际提取到: %s", url2, name2)
	}

	// 测试URL包含哈希值
	url3 := "https://example.com/files/test.iso#fragment"
	name3 := extractFileNameFromURL(url3)
	if name3 != "test.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.iso, 实际提取到: %s", url3, name3)
	}

	// 测试URL包含查询参数和哈希值
	url4 := "https://example.com/files/test.iso?param=value#fragment"
	name4 := extractFileNameFromURL(url4)
	if name4 != "test.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.iso, 实际提取到: %s", url4, name4)
	}

	// 测试非ISO文件名（应该自动添加.iso扩展名）
	url5 := "https://example.com/files/test"
	name5 := extractFileNameFromURL(url5)
	if name5 != "test.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.iso, 实际提取到: %s", url5, name5)
	}

	// 测试非ISO文件名但包含扩展名（应该替换为.iso）
	url6 := "https://example.com/files/test.img"
	name6 := extractFileNameFromURL(url6)
	if name6 != "test.img.iso" {
		t.Errorf("URL %s 应该提取到文件名 test.img.iso, 实际提取到: %s", url6, name6)
	}

	// 测试空文件名（应该返回default.iso）
	url7 := "https://example.com/files/"
	name7 := extractFileNameFromURL(url7)
	if name7 != "default.iso" {
		t.Errorf("URL %s 应该提取到文件名 default.iso, 实际提取到: %s", url7, name7)
	}

	// 测试复杂路径
	url8 := "https://example.com/path/to/files/image.iso"
	name8 := extractFileNameFromURL(url8)
	if name8 != "image.iso" {
		t.Errorf("URL %s 应该提取到文件名 image.iso, 实际提取到: %s", url8, name8)
	}

	// 测试URL编码的文件名
	url9 := "https://example.com/files/test%20file.iso"
	name9 := extractFileNameFromURL(url9)
	if name9 != "test file.iso" {
		t.Errorf("URL %s 应该提取到文件名 test file.iso, 实际提取到: %s", url9, name9)
	}
}

func TestIsValidISO(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 创建有效的ISO文件
	validISOFile := filepath.Join(tempDir, "valid.iso")
	if err := createTestISOFile(validISOFile); err != nil {
		t.Fatalf("创建有效ISO文件失败: %v", err)
	}

	// 测试有效ISO文件
	if !fm.isValidISO(validISOFile) {
		t.Error("有效的ISO文件应该返回true")
	}

	// 创建无效的ISO文件
	invalidISOFile := filepath.Join(tempDir, "invalid.iso")
	invalidFile, _ := os.Create(invalidISOFile)
	invalidFile.Write([]byte("test"))
	invalidFile.Close()

	// 测试无效ISO文件
	if fm.isValidISO(invalidISOFile) {
		t.Error("无效的ISO文件应该返回false")
	}

	// 测试不存在的文件
	if fm.isValidISO("non_existent.iso") {
		t.Error("不存在的文件应该返回false")
	}
}

// TestDownloadISOByURLWithProgress 测试根据URL下载ISO文件并跟踪进度功能
func TestDownloadISOByURLWithProgress(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件管理器
	fm, err := New(tempDir)
	if err != nil {
		t.Fatalf("创建文件管理器失败: %v", err)
	}

	// 启动模拟HTTP服务器
	// 创建一个测试用的ISO内容
	isoContent := make([]byte, 0x8005) // 足够容纳ISO签名
	copy(isoContent[0x8000:0x8005], []byte("CD001"))

	// 创建模拟服务器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置响应头
		w.Header().Set("Content-Type", "application/x-iso9660-image")
		w.Header().Set("Content-Disposition", "attachment; filename=test.iso")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(isoContent)))
		w.WriteHeader(http.StatusOK)
		// 写入模拟ISO内容
		w.Write(isoContent)
	})

	// 启动服务器
	server := httptest.NewServer(handler)
	defer server.Close()

	// 测试下载功能
	filePath, err := fm.DownloadISOByURLWithProgress(server.URL)
	if err != nil {
		t.Fatalf("下载ISO文件失败: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("下载的文件不存在: %s", filePath)
	}

	// 验证文件名是否正确
	expectedFileName := "test.iso"
	actualFileName := filepath.Base(filePath)
	if actualFileName != expectedFileName {
		t.Errorf("文件名不正确，预期: %s, 实际: %s", expectedFileName, actualFileName)
	}

	// 验证文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if fileInfo.Size() != int64(len(isoContent)) {
		t.Errorf("文件大小不一致，预期: %d, 实际: %d", len(isoContent), fileInfo.Size())
	}

	// 验证下载进度
	progress, err := fm.GetDownloadProgress("test.iso")
	if err != nil {
		t.Errorf("获取下载进度失败: %v", err)
	} else {
		// 验证进度信息
		if progress.FileName != "test.iso" {
			t.Errorf("进度信息中的文件名不正确，预期: test.iso, 实际: %s", progress.FileName)
		}
		if progress.Status != "completed" {
			t.Errorf("下载完成后状态应该是 'completed'，但实际是: %s", progress.Status)
		}
		if progress.Progress != 100 {
			t.Errorf("下载完成后进度应该是 100%%，但实际是: %d%%", progress.Progress)
		}
	}

	// 测试下载不存在的URL
	_, err = fm.DownloadISOByURLWithProgress("http://localhost:12345/non_existent.iso")
	if err == nil {
		t.Error("下载不存在的URL应该失败")
	}
}
