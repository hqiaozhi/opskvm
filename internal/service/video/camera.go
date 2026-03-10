package video

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"
)

var (
	CameraWatcherEnabled    bool
	CameraWatcherStopCh     chan struct{}
	CameraWatcherDevicePath string
)

// Camera 定义摄像头设备的核心操作接口
type Camera interface {
	FindDevicePath() string
	Open(path string) error
	ApplyConfig(width, height int, fps uint32) error
	ListConfigs() ([]v4l.DeviceConfig, error)
	GetConfig() (v4l.DeviceConfig, error)
	SetConfig(cfg v4l.DeviceConfig) error
	UpdateConfig(width, height int, fps uint32) (v4l.DeviceConfig, error)
	TurnOn() error
	TurnOff() error
	GetStatus() bool
	Capture() ([]byte, error)
	ResetControls() error
	Close() error
}

// V4LCamera 是Camera接口的V4L实现
type V4LCamera struct {
	dev       *v4l.Device
	path      string
	mu        sync.Mutex
	cfg       v4l.DeviceConfig
	isOn      bool
	captureMu sync.Mutex
}

func NewV4LCamera() *V4LCamera {
	return &V4LCamera{}
}

func (c *V4LCamera) FindDevicePath() string {
	devs := v4l.FindDevices()
	if len(devs) == 0 {
		g.Log().Error(context.Background(), "No video devices found")
		os.Exit(1)
	}
	return devs[0].Path
}

func (c *V4LCamera) Open(path string) error {
	if path == "" {
		path = c.FindDevicePath()
	}
	dev, err := v4l.Open(path)
	if err != nil {
		return err
	}
	c.dev = dev
	c.path = path
	cfg, err := dev.GetConfig()
	if err != nil {
		return err
	}
	c.cfg = cfg
	g.Log().Infof(context.Background(), "camera: Opened camera device: %s", path)
	return nil
}

func (c *V4LCamera) ApplyConfig(width, height int, fps uint32) error {
	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}
	if width > 0 && height > 0 && fps > 0 {
		_, err := c.UpdateConfig(width, height, fps)
		return err
	}
	return nil
}

func (c *V4LCamera) ListConfigs() ([]v4l.DeviceConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev == nil {
		return nil, fmt.Errorf("camera device not open")
	}
	return c.dev.ListConfigs()
}

func (c *V4LCamera) GetConfig() (v4l.DeviceConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg, nil
}

func (c *V4LCamera) SetConfig(cfg v4l.DeviceConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}
	if err := c.dev.SetConfig(cfg); err != nil {
		return err
	}
	c.cfg = cfg
	return nil
}

func (c *V4LCamera) UpdateConfig(width, height int, fps uint32) (v4l.DeviceConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldCfg := c.cfg
	oldDev := c.dev

	newCfg := oldCfg
	if width > 0 {
		newCfg.Width = width
	}
	if height > 0 {
		newCfg.Height = height
	}
	if fps > 0 {
		newCfg.FPS = v4l.Frac{N: fps, D: 1}
	}

	supportedCfgs, err := oldDev.ListConfigs()
	if err != nil {
		return oldCfg, fmt.Errorf("failed to list supported configs: %v", err)
	}

	resolutionSupported := false
	supportedFPS := make([]uint32, 0)
	for _, cfg := range supportedCfgs {
		if cfg.Format == mjpeg.FourCC && cfg.Width == newCfg.Width && cfg.Height == newCfg.Height {
			resolutionSupported = true
			if cfg.FPS.N > 0 && cfg.FPS.D > 0 {
				supportedFPS = append(supportedFPS, cfg.FPS.N/cfg.FPS.D)
			}
		}
	}

	if !resolutionSupported {
		return oldCfg, fmt.Errorf("config %dx%d is not supported", newCfg.Width, newCfg.Height)
	}

	if fps > 0 {
		fpsSupported := false
		for _, supportedFrameRate := range supportedFPS {
			if supportedFrameRate == fps {
				fpsSupported = true
				break
			}
		}
		if len(supportedFPS) == 0 {
			fpsSupported = true
		}
		if !fpsSupported {
			return oldCfg, fmt.Errorf("fps %d is not supported for config %dx%d", fps, newCfg.Width, newCfg.Height)
		}
	}

	oldDev.Close()
	c.dev = nil

	newDev, err := v4l.Open(c.path)
	if err != nil {
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to reopen device: %v", err)
		}
		c.dev = recoverDev
		return oldCfg, fmt.Errorf("failed to reopen device")
	}
	c.dev = newDev

	newCfg.Format = mjpeg.FourCC
	if err := newDev.SetConfig(newCfg); err != nil {
		newDev.Close()
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to set config")
		}
		c.dev = recoverDev
		c.dev.SetConfig(oldCfg)
		c.dev.TurnOn()
		return oldCfg, fmt.Errorf("failed to set config")
	}

	if err := newDev.TurnOn(); err != nil {
		newDev.Close()
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to turn on")
		}
		c.dev = recoverDev
		c.dev.SetConfig(oldCfg)
		c.dev.TurnOn()
		c.isOn = true
		return oldCfg, fmt.Errorf("failed to turn on")
	}

	c.isOn = true

	actualCfg, err := newDev.GetConfig()
	if err != nil {
		return newCfg, fmt.Errorf("failed to get actual config")
	}
	c.cfg = actualCfg

	g.Log().Infof(context.Background(), "Camera config updated: %dx%d @ %.2f FPS",
		actualCfg.Width, actualCfg.Height,
		float64(actualCfg.FPS.N)/float64(actualCfg.FPS.D))

	return actualCfg, nil
}

func (c *V4LCamera) TurnOn() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}
	err := c.dev.TurnOn()
	if err == nil {
		c.isOn = true
	}
	return err
}

func (c *V4LCamera) TurnOff() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}
	c.dev.TurnOff()
	c.isOn = false
	return nil
}

func (c *V4LCamera) GetStatus() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isOn
}

func (c *V4LCamera) Capture() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dev == nil {
		return nil, fmt.Errorf("camera device not open")
	}
	if !c.isOn {
		return nil, fmt.Errorf("camera is turned off")
	}
	buf, err := c.dev.Capture()
	if err != nil {
		return nil, err
	}
	data := make([]byte, buf.Size())
	_, err = buf.ReadAt(data, 0)
	return data, err
}

func (c *V4LCamera) ResetControls() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}
	ctrls, err := c.dev.ListControls()
	if err != nil {
		return err
	}
	for _, ctrl := range ctrls {
		if err := c.dev.SetControl(ctrl.CID, ctrl.Default); err != nil {
			g.Log().Warningf(context.Background(), "failed to reset control %d: %v", ctrl.CID, err)
		}
	}
	return nil
}

func (c *V4LCamera) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev != nil {
		dev := c.dev
		c.dev = nil
		dev.Close()
	}
	return nil
}

// CheckCameraExists 检查摄像头设备是否存在
func CheckCameraExists(devicePath string) bool {
	_, err := os.Stat(devicePath)
	return err == nil
}

// WaitForCamera 等待摄像头设备插入
func WaitForCamera(devicePath string, maxWaitTime time.Duration, checkInterval time.Duration) (string, error) {
	startTime := time.Now()

	for {
		targetPath := devicePath
		if targetPath == "" {
			devs := v4l.FindDevices()
			if len(devs) > 0 {
				g.Log().Infof(context.Background(), "Found camera device: %s", devs[0].Path)
				return devs[0].Path, nil
			}
		} else {
			if CheckCameraExists(targetPath) {
				g.Log().Infof(context.Background(), "Found camera device: %s", targetPath)
				return targetPath, nil
			}
		}

		if maxWaitTime > 0 && time.Since(startTime) >= maxWaitTime {
			return "", fmt.Errorf("camera: Timeout waiting for device")
		}

		g.Log().Infof(context.Background(), "Waiting for camera device to be inserted...")
		time.Sleep(checkInterval)
	}
}

// findCameraDevice 查找摄像头设备
func findCameraDevice() string {
	devs := v4l.FindDevices()
	if len(devs) > 0 {
		return devs[0].Path
	}
	return ""
}

// StartCameraWatcher 启动摄像头设备状态监控器
func StartCameraWatcher(onLost func(), onReconnect func(newDevicePath string)) {
	CameraWatcherEnabled = true
	CameraWatcherStopCh = make(chan struct{})

	g.Log().Infof(context.Background(), "Camera watcher starting...")

	go func() {
		for {
			select {
			case <-CameraWatcherStopCh:
				g.Log().Info(context.Background(), "Camera watcher stopped")
				return
			default:
			}

			currentPath := findCameraDevice()

			if currentPath == "" {
				g.Log().Infof(context.Background(), "No camera found, waiting for device...")
				for {
					select {
					case <-CameraWatcherStopCh:
						g.Log().Info(context.Background(), "Camera watcher stopped while waiting")
						return
					default:
					}

					newPath := findCameraDevice()
					if newPath != "" {
						g.Log().Infof(context.Background(), "Camera device found: %s", newPath)
						CameraWatcherDevicePath = newPath
						if onReconnect != nil {
							onReconnect(newPath)
						}
						break
					}
					time.Sleep(1 * time.Second)
				}
			} else if !CheckCameraExists(currentPath) {
				g.Log().Warningf(context.Background(), "Camera device lost!")
				CameraWatcherDevicePath = ""
				if onLost != nil {
					onLost()
				}

				for {
					select {
					case <-CameraWatcherStopCh:
						g.Log().Info(context.Background(), "Camera watcher stopped while waiting for reconnection")
						return
					default:
					}

					newPath := findCameraDevice()
					if newPath != "" {
						g.Log().Infof(context.Background(), "Camera device reconnected: %s", newPath)
						CameraWatcherDevicePath = newPath
						if onReconnect != nil {
							onReconnect(newPath)
						}
						break
					}
					time.Sleep(1 * time.Second)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
	g.Log().Infof(context.Background(), "Camera watcher started")
}

// StopCameraWatcher 停止摄像头设备监控器
func StopCameraWatcher() {
	if CameraWatcherEnabled {
		g.Log().Infof(context.Background(), "Stopping camera watcher...")
		CameraWatcherEnabled = false
		if CameraWatcherStopCh != nil {
			close(CameraWatcherStopCh)
		}
	}
}
