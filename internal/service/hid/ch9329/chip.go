package ch9329

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.bug.st/serial"

	gLog "github.com/gogf/gf/v2/frame/g"
)

var (
	CH9329WatcherEnabled    bool
	CH9329WatcherStopCh     chan struct{}
	CH9329WatcherDevicePath string
)

// CH9329 命令类型
const (
	CmdTypeKeyboard      = 0x02
	CmdTypeMouseAbsolute = 0x04
	CmdTypeMouseRelative = 0x05
)

// CH9329Device 实现HIDDevice接口
type CH9329Device struct {
	devicePath   string
	port         serial.Port
	absolute     bool
	ledState     byte
	modifiers    byte   // 键盘修饰键状态
	activeKeys   []byte // 活动按键列表（最多6个）
	mouseButtons byte   // 鼠标按键状态
	mouseX       int    // 鼠标X坐标（绝对模式）
	mouseY       int    // 鼠标Y坐标（绝对模式）
	mouseDeltaX  int    // 鼠标X增量（相对模式）
	mouseDeltaY  int    // 鼠标Y增量（相对模式）
	mouseWheel   int8   // 鼠标滚轮状态
}

// NewCH9329Device 创建新的CH9329设备实例
func NewCH9329(devicePath string) *CH9329Device {
	return &CH9329Device{
		devicePath:   devicePath,
		absolute:     true,               // 默认使用绝对鼠标模式
		ledState:     0,                  // 初始LED状态
		modifiers:    0,                  // 初始修饰键状态
		activeKeys:   make([]byte, 0, 6), // 初始活动按键列表，最多6个
		mouseButtons: 0,                  // 初始鼠标按键状态
		mouseX:       0,                  // 初始鼠标X坐标
		mouseY:       0,                  // 初始鼠标Y坐标
		mouseDeltaX:  0,                  // 初始鼠标X增量
		mouseDeltaY:  0,                  // 初始鼠标Y增量
		mouseWheel:   0,                  // 初始鼠标滚轮状态
	}
}

// calculateChecksum 计算校验和
func (d *CH9329Device) calculateChecksum(data []byte) byte {
	var sum int
	for _, b := range data {
		sum += int(b)
	}
	return byte(sum % 256)
}

// ChipResponseError 定义芯片响应错误
type ChipResponseError struct {
	Message string
}

func (e *ChipResponseError) Error() string {
	return e.Message
}

// sendCommand 发送命令到CH9329设备，与Python版本完全一致
func (d *CH9329Device) sendCommand(cmd []byte) error {
	// 检查设备是否已打开
	if d.port == nil {
		return fmt.Errorf("device not opened")
	}

	// 完整命令格式：0x57 AB + cmd + checksum
	fullCmd := make([]byte, len(cmd)+3)
	fullCmd[0] = 0x57 // 起始字节1
	fullCmd[1] = 0xAB // 起始字节2
	copy(fullCmd[2:], cmd)
	checksum := d.calculateChecksum(fullCmd[:len(cmd)+2])
	fullCmd[len(cmd)+2] = checksum

	// 发送命令
	_, err := d.port.Write(fullCmd)
	if err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}

	// 添加微小延迟，确保设备有时间处理命令
	time.Sleep(1 * time.Millisecond)

	// 接收响应
	return d.receiveResponse()
}

// receiveResponse 接收并处理设备响应，与Python版本完全一致
func (d *CH9329Device) receiveResponse() error {
	// 读取响应，至少5字节
	data := make([]byte, 5)
	n, err := d.port.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if n < 5 {
		return &ChipResponseError{Message: "Too short response, HID might be disconnected"}
	}

	// 如果有更多数据需要读取
	if len(data) > 4 && data[4] > 0 {
		additionalData := make([]byte, data[4]+1)
		n, err = d.port.Read(additionalData)
		if err != nil {
			return fmt.Errorf("failed to read additional response data: %w", err)
		}

		if n < int(data[4]+1) {
			return &ChipResponseError{Message: "Incomplete additional response data"}
		}

		// 合并数据
		data = append(data, additionalData[:n]...)
	}

	// 验证校验和
	if len(data) > 1 {
		expectedChecksum := d.calculateChecksum(data[:len(data)-1])
		if data[len(data)-1] != expectedChecksum {
			return &ChipResponseError{Message: "Invalid response checksum"}
		}
	}

	// 检查响应中的错误码
	if len(data) > 5 && data[4] == 1 && data[5] != 0 {
		return &ChipResponseError{Message: fmt.Sprintf("Response error code = %v", data[5])}
	}

	// 更新LED状态（如果是信息响应）
	if len(data) > 7 && data[3] == 0x81 {
		d.ledState = data[7]
	}

	return nil
}

// Open 打开串口连接（实现hid.KMHIDController接口）
func (d *CH9329Device) Open() error {
	// 默认参数
	if d.devicePath == "" {
		matches, err := d.findTTYUSBPort()
		if err != nil {
			return err
		}
		if len(matches) > 0 {
			d.devicePath = matches[0]
		} else {
			return fmt.Errorf("no ttyUSB device found")
		}
	}
	return d.OpenWithParams(d.devicePath, 9600)
}

// FindTTYUSBPort 查找系统中所有 /dev/ttyUSB* 格式的串口设备
func (d *CH9329Device) findTTYUSBPort() ([]string, error) {
	// 直接匹配 /dev/ttyUSB 开头的所有设备文件
	matches, err := filepath.Glob("/dev/ttyUSB*")
	if err != nil {
		return nil, fmt.Errorf("查找串口设备失败: %v", err)
	}
	return matches, nil
}

// OpenWithParams 打开串口连接
func (d *CH9329Device) OpenWithParams(portName string, baudRate int) error {
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}
	d.port = port

	// 初始化CH9329设备，发送重置命令
	// RESET = [0x00,0x0F,0x00]
	resetCmd := []byte{0x00, 0x0F, 0x00}
	if err := d.sendCommand(resetCmd); err != nil {
		return fmt.Errorf("failed to reset device: %w", err)
	}

	// 等待设备初始化完成
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Close 关闭串口连接
func (d *CH9329Device) Close() error {
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}

// GetInfo 获取设备信息
func (d *CH9329Device) GetInfo() error {
	// GET_INFO = [0x00, 0x01, 0x00]
	infoCmd := []byte{0x00, 0x01, 0x00}
	return d.sendCommand(infoCmd)
}

// CheckDeviceExists 检查 CH9329 设备是否存在
func CheckDeviceExists(devicePath string) bool {
	_, err := os.Stat(devicePath)
	return err == nil
}

// WaitForCH9329 等待 CH9329 设备插入
// devicePath: 设备路径，为空则自动查找
// maxWaitTime: 最大等待时间，0 表示无限等待
// checkInterval: 检查间隔
func WaitForCH9329(devicePath string, maxWaitTime time.Duration, checkInterval time.Duration) (string, error) {
	startTime := time.Now()

	for {
		targetPath := devicePath
		if targetPath == "" {
			matches, err := filepath.Glob("/dev/ttyUSB*")
			if err == nil && len(matches) > 0 {
				gLog.Log().Infof(context.Background(), "Found CH9329 device: %s", matches[0])
				return matches[0], nil
			}
		} else {
			if CheckDeviceExists(targetPath) {
				gLog.Log().Infof(context.Background(), "Found CH9329 device: %s", targetPath)
				return targetPath, nil
			}
		}

		if maxWaitTime > 0 && time.Since(startTime) >= maxWaitTime {
			return "", fmt.Errorf("CH9329: Timeout waiting for device")
		}

		gLog.Log().Infof(context.Background(), "Waiting for CH9329 device to be inserted...")
		time.Sleep(checkInterval)
	}
}

// StartCH9329Watcher 启动 CH9329 设备状态监控器
// devicePath: 初始设备路径，如果是空则自动检测设备类别
// onLost: 设备丢失时的回调
// onReconnect: 设备重新插入时的回调，传入新的设备路径
func StartCH9329Watcher(devicePath string, onLost func(), onReconnect func(newDevicePath string)) {
	CH9329WatcherEnabled = true
	CH9329WatcherStopCh = make(chan struct{})
	CH9329WatcherDevicePath = devicePath

	go func() {
		for {
			select {
			case <-CH9329WatcherStopCh:
				gLog.Log().Info(context.Background(), "CH9329 watcher stopped")
				return
			default:
				if !CheckDeviceExists(devicePath) {
					gLog.Log().Warningf(context.Background(), "CH9329 device %s lost!", devicePath)
					if onLost != nil {
						onLost()
					}

					for {
						select {
						case <-CH9329WatcherStopCh:
							return
						default:
							newPath := findCH9329Device(devicePath)
							if newPath != "" {
								gLog.Log().Infof(context.Background(), "CH9329 device reconnected: %s", newPath)
								devicePath = newPath
								CH9329WatcherDevicePath = newPath
								if onReconnect != nil {
									onReconnect(newPath)
								}
								goto WaitNext
							}
							time.Sleep(1 * time.Second)
						}
					}
				WaitNext:
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	gLog.Log().Infof(context.Background(), "CH9329 watcher started for device: %s", devicePath)
}

func findCH9329Device(originalPath string) string {
	if originalPath != "" && CheckDeviceExists(originalPath) {
		return originalPath
	}
	matches, _ := filepath.Glob("/dev/ttyUSB*")
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// StopCH9329Watcher 停止 CH9329 设备监控器
func StopCH9329Watcher() {
	if CH9329WatcherEnabled {
		close(CH9329WatcherStopCh)
		CH9329WatcherEnabled = false
	}
}
