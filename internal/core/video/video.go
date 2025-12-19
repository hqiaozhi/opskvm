package video

import (
	"fmt"
	"sync"

	"github.com/blackjack/webcam"
)

// VideoStreamer 管理视频流的生命周期
type VideoStreamer struct {
	cam         *webcam.Webcam
	pixelFormat webcam.PixelFormat
	width       int
	height      int
	running     bool
	mu          sync.Mutex
}

// NewVideoStreamer 创建一个新的视频流管理器
func NewVideoStreamer(devicePath string, width, height int) (*VideoStreamer, error) {
	// 打开视频设备
	cam, err := webcam.Open(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video device: %w", err)
	}

	// 获取支持的像素格式
	formats := cam.GetSupportedFormats()
	if len(formats) == 0 {
		cam.Close()
		return nil, fmt.Errorf("no supported pixel formats found")
	}

	// 优先选择MJPEG像素格式（更高质量），如果没有则选择第一个可用格式
	var pixelFormat webcam.PixelFormat
	mjpegFound := false
	for pf := range formats {
		// MJPEG格式的4CC代码是'MJPG'
		if uint32(pf) == 1196444237 { // 1196444237 is the uint32 value for 'MJPG'
			pixelFormat = pf
			mjpegFound = true
			break
		}
	}

	// 如果没有找到MJPEG，则选择第一个可用格式
	if !mjpegFound {
		for pf := range formats {
			pixelFormat = pf
			break
		}
	}

	// 设置视频格式
	_, _, _, err = cam.SetImageFormat(pixelFormat, uint32(width), uint32(height))
	if err != nil {
		cam.Close()
		return nil, fmt.Errorf("failed to set image format: %w", err)
	}

	// 初始化缓冲区
	err = cam.StartStreaming()
	if err != nil {
		cam.Close()
		return nil, fmt.Errorf("failed to start streaming: %w", err)
	}

	return &VideoStreamer{
		cam:         cam,
		pixelFormat: pixelFormat,
		width:       width,
		height:      height,
		running:     true,
	}, nil
}

// ReadFrame 读取一帧视频数据
func (vs *VideoStreamer) ReadFrame() ([]byte, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if !vs.running {
		return nil, fmt.Errorf("video streamer is not running")
	}

	// 等待帧可用，设置1秒超时
	err := vs.cam.WaitForFrame(1)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for frame: %w", err)
	}

	// 读取帧数据
	frame, err := vs.cam.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf("failed to read frame: %w", err)
	}

	return frame, nil
}

// Close 关闭视频流
func (vs *VideoStreamer) Close() error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if !vs.running {
		return nil
	}

	vs.running = false

	// 停止流并关闭设备
	vs.cam.StopStreaming()
	return vs.cam.Close()
}

// Width 返回视频宽度
func (vs *VideoStreamer) Width() int {
	return vs.width
}

// Height 返回视频高度
func (vs *VideoStreamer) Height() int {
	return vs.height
}

// PixelFormat 返回像素格式
func (vs *VideoStreamer) PixelFormat() webcam.PixelFormat {
	return vs.pixelFormat
}

// IsRunning 检查视频流是否正在运行
func (vs *VideoStreamer) IsRunning() bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	return vs.running
}
