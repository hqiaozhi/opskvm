package video

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"
)

// Camera 定义摄像头设备的核心操作接口
type Camera interface {
	FindDevicePath() string
	Open(path string) error
	ListConfigs() ([]v4l.DeviceConfig, error)
	GetConfig() (v4l.DeviceConfig, error)
	SetConfig(cfg v4l.DeviceConfig) error
	UpdateConfig(width, height int, fps uint32) (v4l.DeviceConfig, error)
	TurnOn() error
	TurnOff() error
	GetStatus() bool // 获取摄像头开关状态
	Capture() ([]byte, error)
	ResetControls() error
	Close() error
}

// V4LCamera 是Camera接口的V4L实现（path字段保存设备路径）
type V4LCamera struct {
	dev       *v4l.Device
	path      string     // 保存设备路径，用于重新打开
	mu        sync.Mutex // 配置更新锁
	cfg       v4l.DeviceConfig
	isOn      bool       // 摄像头开关状态
	captureMu sync.Mutex // 采集锁，避免并发采集问题
}

func NewV4LCamera() *V4LCamera {
	return &V4LCamera{}
}

func (c *V4LCamera) FindDevicePath() string {
	devs := v4l.FindDevices()
	if len(devs) == 0 {
		log.Fatalf("No video devices found")
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
	c.path = path // 保存设备路径
	cfg, err := dev.GetConfig()
	if err != nil {
		return err
	}
	c.cfg = cfg
	log.Printf("camera: Opened camera device: %s", path)
	return nil
}

// ApplyConfig 应用配置到摄像头
func (c *V4LCamera) ApplyConfig(width, height int, fps uint32) error {
	// c.mu.Lock()
	// defer c.mu.Unlock()

	if c.dev == nil {
		return fmt.Errorf("camera device not open")
	}

	// 只有在配置值大于0时才应用
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

// UpdateConfig 关闭设备→重新打开→应用新配置
func (c *V4LCamera) UpdateConfig(width, height int, fps uint32) (v4l.DeviceConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. 保存原有状态，用于失败时恢复
	oldCfg := c.cfg
	oldDev := c.dev

	// 2. 构建新配置
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

	// 3. 验证配置是否支持
	supportedCfgs, err := oldDev.ListConfigs()
	if err != nil {
		return oldCfg, fmt.Errorf("failed to list supported configs: %v", err)
	}

	// 检查设备是否支持该分辨率和格式
	resolutionSupported := false
	supportedFPS := make([]uint32, 0)
	for _, cfg := range supportedCfgs {
		if cfg.Format == mjpeg.FourCC && cfg.Width == newCfg.Width && cfg.Height == newCfg.Height {
			resolutionSupported = true
			// 收集该分辨率支持的所有帧率
			if cfg.FPS.N > 0 && cfg.FPS.D > 0 {
				supportedFPS = append(supportedFPS, cfg.FPS.N/cfg.FPS.D)
			}
		}
	}

	if !resolutionSupported {
		return oldCfg, fmt.Errorf("config %dx%d is not supported", newCfg.Width, newCfg.Height)
	}

	// 如果指定了帧率，检查是否支持
	if fps > 0 {
		fpsSupported := false
		// 检查帧率是否在支持列表中
		for _, supportedFrameRate := range supportedFPS {
			if supportedFrameRate == fps {
				fpsSupported = true
				break
			}
		}
		// 如果支持列表为空，说明设备支持任意帧率
		if len(supportedFPS) == 0 {
			fpsSupported = true
		}

		if !fpsSupported {
			return oldCfg, fmt.Errorf("fps %d is not supported for config %dx%d, supported fps: %v", fps, newCfg.Width, newCfg.Height, supportedFPS)
		}
	}

	// 4. 关闭当前设备（核心步骤）
	log.Println("Closing camera device for config update...")
	oldDev.Close()
	c.dev = nil // 标记设备已关闭

	// 5. 重新打开设备（核心步骤）
	log.Println("Reopening camera device with new config...")
	newDev, err := v4l.Open(c.path)
	if err != nil {
		// 尝试恢复原有设备（打开失败时）
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to reopen device: %v (and recover failed: %v)", err, recoverErr)
		}
		c.dev = recoverDev
		return oldCfg, fmt.Errorf("failed to reopen device: %v (recovered old device)", err)
	}
	c.dev = newDev

	// 6. 应用新配置到新打开的设备
	newCfg.Format = mjpeg.FourCC // 强制MJPEG格式
	if err := newDev.SetConfig(newCfg); err != nil {
		// 配置设置失败，恢复原有设备
		newDev.Close()
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to set new config: %v (recover failed: %v)", err, recoverErr)
		}
		c.dev = recoverDev
		c.dev.SetConfig(oldCfg)
		c.dev.TurnOn()
		return oldCfg, fmt.Errorf("failed to set new config: %v (recovered old config)", err)
	}

	// 7. 启动新配置的设备
	if err := newDev.TurnOn(); err != nil {
		// 启动失败，恢复原有设备
		newDev.Close()
		recoverDev, recoverErr := v4l.Open(c.path)
		if recoverErr != nil {
			return oldCfg, fmt.Errorf("failed to turn on new config: %v (recover failed: %v)", err, recoverErr)
		}
		c.dev = recoverDev
		c.dev.SetConfig(oldCfg)
		c.dev.TurnOn()
		c.isOn = true // 更新摄像头状态
		return oldCfg, fmt.Errorf("failed to turn on new config: %v (recovered old config)", err)
	}

	// 更新摄像头状态
	c.isOn = true

	// 8. 获取实际生效的配置
	actualCfg, err := newDev.GetConfig()
	if err != nil {
		return newCfg, fmt.Errorf("failed to get actual config: %v (but config applied)", err)
	}
	c.cfg = actualCfg

	log.Printf("Camera config updated successfully: %dx%d @ %.2f FPS",
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
	c.dev.TurnOff() // TurnOff方法没有返回值，直接调用
	c.isOn = false
	return nil
}

// GetStatus 获取摄像头开关状态
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
			log.Printf("Warning: failed to reset control %d: %v", ctrl.CID, err)
		}
	}
	return nil
}

func (c *V4LCamera) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dev != nil {
		c.dev.Close()
		c.dev = nil
	}
	return nil
}

// cfg2str 将设备配置转为字符串
func Cfg2str(cfg v4l.DeviceConfig) string {
	return fmt.Sprintf("%dx%d @ %.4g FPS", cfg.Width, cfg.Height,
		float64(cfg.FPS.N)/float64(cfg.FPS.D))
}
