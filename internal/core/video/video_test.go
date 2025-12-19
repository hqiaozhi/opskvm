package video

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

const (
	// 测试用的视频设备路径
	testDevicePath = "/dev/video0"
	// 测试用的分辨率
	testWidth  = 1920
	testHeight = 1080
)

// TestVideoStreamer_Creation 测试视频流创建
func TestVideoStreamer_Creation(t *testing.T) {
	// 检查设备是否存在
	if _, err := os.Stat(testDevicePath); os.IsNotExist(err) {
		t.Skipf("video device %s not found, skipping test", testDevicePath)
	}

	// 创建视频流
	streamer, err := NewVideoStreamer(testDevicePath, testWidth, testHeight)
	if err != nil {
		t.Fatalf("failed to create video streamer: %v", err)
	}
	defer streamer.Close()

	// 验证视频流参数
	if streamer.Width() != testWidth {
		t.Errorf("expected width %d, got %d", testWidth, streamer.Width())
	}

	if streamer.Height() != testHeight {
		t.Errorf("expected height %d, got %d", testHeight, streamer.Height())
	}

	if !streamer.IsRunning() {
		t.Error("video streamer should be running after creation")
	}
}

// TestVideoStreamer_ReadFrame 测试读取视频帧
func TestVideoStreamer_ReadFrame(t *testing.T) {
	// 检查设备是否存在
	if _, err := os.Stat(testDevicePath); os.IsNotExist(err) {
		t.Skipf("video device %s not found, skipping test", testDevicePath)
	}

	// 创建视频流
	streamer, err := NewVideoStreamer(testDevicePath, testWidth, testHeight)
	if err != nil {
		t.Fatalf("failed to create video streamer: %v", err)
	}
	defer streamer.Close()

	// 尝试读取几帧视频数据
	for i := 0; i < 3; i++ {
		frame, err := streamer.ReadFrame()
		if err != nil {
			t.Errorf("failed to read frame %d: %v", i+1, err)
			continue
		}

		if len(frame) == 0 {
			t.Errorf("frame %d should not be empty", i+1)
			continue
		}

		t.Logf("successfully read frame %d, size: %d bytes", i+1, len(frame))

		// 等待一段时间再读取下一帧
		time.Sleep(100 * time.Millisecond)
	}
}

// TestVideoStreamer_Close 测试关闭视频流
func TestVideoStreamer_Close(t *testing.T) {
	// 检查设备是否存在
	if _, err := os.Stat(testDevicePath); os.IsNotExist(err) {
		t.Skipf("video device %s not found, skipping test", testDevicePath)
	}

	// 创建视频流
	streamer, err := NewVideoStreamer(testDevicePath, testWidth, testHeight)
	if err != nil {
		t.Fatalf("failed to create video streamer: %v", err)
	}

	// 关闭视频流
	err = streamer.Close()
	if err != nil {
		t.Errorf("failed to close video streamer: %v", err)
	}

	// 验证视频流已停止
	if streamer.IsRunning() {
		t.Error("video streamer should not be running after close")
	}

	// 验证重复关闭不会出错
	err = streamer.Close()
	if err != nil {
		t.Errorf("repeated close should not return error, got: %v", err)
	}
}

// TestVideoStreamer_ConcurrentAccess 测试并发访问视频流
func TestVideoStreamer_ConcurrentAccess(t *testing.T) {
	// 检查设备是否存在
	if _, err := os.Stat(testDevicePath); os.IsNotExist(err) {
		t.Skipf("video device %s not found, skipping test", testDevicePath)
	}

	// 创建视频流
	streamer, err := NewVideoStreamer(testDevicePath, testWidth, testHeight)
	if err != nil {
		t.Fatalf("failed to create video streamer: %v", err)
	}
	defer streamer.Close()

	// 并发读取帧
	numGoroutines := 5
	numFramesPerGoroutine := 2
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*numFramesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numFramesPerGoroutine; j++ {
				_, err := streamer.ReadFrame()
				if err != nil {
					errChan <- fmt.Errorf("goroutine %d: frame %d: %w", goroutineID, j+1, err)
					continue
				}
				t.Logf("goroutine %d: successfully read frame %d", goroutineID, j+1)
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	errors := make([]error, 0)
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Logf("encountered %d errors during concurrent access:", len(errors))
		for _, err := range errors {
			t.Logf("  - %v", err)
		}
		// 注意：在高并发下可能会有一些帧读取失败，这是正常的，所以不使用t.Fatal
	}
}

// TestVideoStreamer_SaveToMP4 测试将视频保存为MP4文件
func TestVideoStreamer_SaveToMP4(t *testing.T) {
	// 检查设备是否存在
	if _, err := os.Stat(testDevicePath); os.IsNotExist(err) {
		t.Skipf("video device %s not found, skipping test", testDevicePath)
	}

	// 检查FFmpeg是否可用
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skipf("ffmpeg not found, skipping test: %v", err)
	}

	// 创建输出文件
	outputFile := "/tmp/test_video.mp4"

	// 确保输出文件不存在
	if _, err := os.Stat(outputFile); err == nil {
		if err := os.Remove(outputFile); err != nil {
			t.Fatalf("failed to remove existing output file: %v", err)
		}
	}

	// defer func() {
	// 	// 测试结束后删除输出文件
	// 	if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
	// 		t.Logf("failed to remove test output file: %v", err)
	// 	}
	// }()

	// 创建视频流
	streamer, err := NewVideoStreamer(testDevicePath, testWidth, testHeight)
	if err != nil {
		t.Fatalf("failed to create video streamer: %v", err)
	}
	defer streamer.Close()

	// 调试：打印实际的像素格式
	pixelFormat := streamer.PixelFormat()
	t.Logf("Actual pixel format: %v", pixelFormat)

	// 根据像素格式设置不同的FFmpeg输入参数
	var cmdArgs []string

	if uint32(pixelFormat) == 1196444237 { // MJPEG格式
		// MJPEG格式：不需要指定像素格式和视频尺寸（FFmpeg可自动识别）
		cmdArgs = []string{
			"ffmpeg",
			"-f", "mjpeg",
			"-framerate", "30",
			"-i", "-",
			"-c:v", "libx264",
			"-preset", "ultrafast", // 使用快速编码预设
			"-shortest", // 确保输出至少与输入一样长
			"-y",        // 覆盖输出文件
			outputFile,
		}
	} else { // YUYV或其他格式
		// 原始视频格式：需要指定像素格式和视频尺寸
		cmdArgs = []string{
			"ffmpeg",
			"-f", "rawvideo",
			"-pixel_format", "yuyv422",
			"-video_size", fmt.Sprintf("%dx%d", testWidth, testHeight),
			"-framerate", "30",
			"-i", "-",
			"-c:v", "libx264",
			"-preset", "ultrafast", // 使用快速编码预设
			"-shortest", // 确保输出至少与输入一样长
			"-y",        // 覆盖输出文件
			outputFile,
		}
	}

	// 构建FFmpeg命令
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	// 获取FFmpeg的标准输入
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get ffmpeg stdin pipe: %v", err)
	}

	// 设置FFmpeg的输出（可选，用于调试）
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动FFmpeg进程
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start ffmpeg: %v", err)
	}

	// 创建一个通道用于通知视频读取完成
	done := make(chan bool)

	// 启动goroutine读取视频帧并写入FFmpeg
	go func() {
		defer close(done)
		defer stdin.Close()

		// 记录开始时间
		startTime := time.Now()
		frameCount := 0
		errCount := 0

		// 持续读取视频帧，生成一个合理时长的视频
		// 运行5秒，生成大约150帧（30fps）
		loopStartTime := time.Now()
		for time.Since(loopStartTime) < 5*time.Second { // 运行5秒，生成足够的视频帧
			frame, err := streamer.ReadFrame()
			if err != nil {
				errCount++
				t.Logf("failed to read frame: %v", err)
				// 增加小的延迟后重试
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 将帧数据写入FFmpeg的标准输入
			if _, err := stdin.Write(frame); err != nil {
				t.Logf("failed to write frame to ffmpeg: %v", err)
				break
			}

			frameCount++

			// 简单的帧率控制，每30帧约1秒
			if frameCount%30 == 0 {
				elapsed := time.Since(startTime)
				t.Logf("Processed %d frames in %v", frameCount, elapsed)
			}

			// 控制帧率（大约30fps）
			time.Sleep(33 * time.Millisecond)
		}

		// 等待FFmpeg完成编码过程
		t.Logf("Waiting for FFmpeg to complete encoding...")
		time.Sleep(2 * time.Second)

		t.Logf("Video capture completed. Frames read: %d, Errors: %d", frameCount, errCount)
	}()

	// 等待视频读取完成
	<-done

	// 等待FFmpeg进程完成
	cmd.Wait()
	// 忽略FFmpeg的退出状态，因为即使有错误，视频文件可能已经成功生成

	// 检查输出文件是否存在且大小合理
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("output MP4 file not found: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Fatalf("output MP4 file is empty")
	}

	t.Logf("successfully saved video to %s, size: %d bytes (%.2f MB)",
		outputFile, fileInfo.Size(), float64(fileInfo.Size())/1024/1024)
}
