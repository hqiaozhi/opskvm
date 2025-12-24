package video

import (
	"context"
	"errors"
	"fmt"

	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"
)

func New() VideoManager {
	v := &Config{
		Path:   "/dev/video0",
		Width:  1920,
		Height: 1080,
		FPS:    10,
	}
	err := v.Open()
	if err != nil {
		panic(err)
	}
	err = v.SetConfig()
	if err != nil {
		panic(err)
	}
	err = v.cam.TurnOn()
	if err != nil {
		panic(err)
	}
	cfg, err := v.cam.GetConfig()
	if err != nil {
		panic(err)
	}
	fmt.Printf("FPS: %d, Format: %d, Width: %d, Height: %d\n", cfg.FPS.N, cfg.Format, cfg.Width, cfg.Height)

	ctrls, err := v.cam.ListControls()
	if err != nil {
		panic(err)
	}
	for _, ctrl := range ctrls {
		v.cam.SetControl(ctrl.CID, ctrl.Default)
	}

	go handleInterrupt()
	go stream(v.cam)

	return v
}

// GetDevicesPath 获取所有视频设备路径
func (v *Config) GetPath() []string {
	var paths []string
	devs := v4l.FindDevices()
	for _, dev := range devs {
		paths = append(paths, dev.Path)
	}
	return paths
}

// 开启设备
func (v *Config) Open() error {
	cam, err := v4l.Open(v.Path)
	if err != nil {
		return err
	}
	v.cam = cam
	return nil
}

type deviceSupportConfig struct {
	Width  int
	Height int
	FPS    int
}

// GetConfig 获取当前设备支持的配置
func (v *Config) ListConfigs() (map[string][]deviceSupportConfig, error) {
	dsc_map := make(map[string][]deviceSupportConfig)
	if v.cam == nil {
		return nil, nil
	}
	devCfg, err := v.cam.ListConfigs()
	if err != nil {
		return nil, err
	}

	var dsc []deviceSupportConfig
	for _, cfg := range devCfg {
		dsc = append(dsc, deviceSupportConfig{
			Width:  int(cfg.Width),
			Height: int(cfg.Height),
			FPS:    int(cfg.FPS.N) / int(cfg.FPS.D),
		})
	}
	dsc_map[v.Path] = dsc

	return dsc_map, nil
}

// SetConfig 设置设备配置
func (v *Config) SetConfig() error {
	cfg, err := v.cam.GetConfig()
	if err != nil {
		return err
	}
	cfg.FPS = v4l.Frac{N: uint32(v.FPS), D: 1}
	cfg.Format = mjpeg.FourCC
	cfg.Width = v.Width
	cfg.Height = v.Height
	// 将修改后的配置应用到摄像头设备
	return v.cam.SetConfig(cfg)
}

// StreamClient 视频流处理
func (v *Config) StreamClient(ctx context.Context) (*Client, error) {
	mu.Lock()
	defer mu.Unlock()
	if stopped {
		return nil, errors.New("video stream is stopped")
	}
	clt := &Client{
		i:  len(clients),
		CH: make(chan []byte, 1),
	}
	// 不再发送nil帧，等待stream函数发送实际的视频帧
	clients = append(clients, clt)
	return clt, nil
}

// 关闭设备
func (v *Config) Close() error {
	v.cam.Close()
	return nil
}
