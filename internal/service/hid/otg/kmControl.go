package otg

import (
	"fmt"
	"log"
	"opskvm/internal/service/hid"
	"os"
	"syscall"
	"time"
)

// OTG 键盘按键编码（与CH9329保持一致）
const (
	KeyA = 0x04
	KeyB = 0x05
	KeyC = 0x06
	KeyD = 0x07
	KeyE = 0x08
	KeyF = 0x09
	KeyG = 0x0A
	KeyH = 0x0B
	KeyI = 0x0C
	KeyJ = 0x0D
	KeyK = 0x0E
	KeyL = 0x0F
	KeyM = 0x10
	KeyN = 0x11
	KeyO = 0x12
	KeyP = 0x13
	KeyQ = 0x14
	KeyR = 0x15
	KeyS = 0x16
	KeyT = 0x17
	KeyU = 0x18
	KeyV = 0x19
	KeyW = 0x1A
	KeyX = 0x1B
	KeyY = 0x1C
	KeyZ = 0x1D

	Key1 = 0x1E
	Key2 = 0x1F
	Key3 = 0x20
	Key4 = 0x21
	Key5 = 0x22
	Key6 = 0x23
	Key7 = 0x24
	Key8 = 0x25
	Key9 = 0x26
	Key0 = 0x27

	KeyEnter      = 0x28
	KeyEscape     = 0x29
	KeyBackspace  = 0x2A
	KeyTab        = 0x2B
	KeySpace      = 0x2C
	KeyMinus      = 0x2D
	KeyEqual      = 0x2E
	KeyLeftBrace  = 0x2F
	KeyRightBrace = 0x30
	KeyBackslash  = 0x31
	KeySemicolon  = 0x33
	KeyApostrophe = 0x34
	KeyGrave      = 0x35
	KeyComma      = 0x36
	KeyDot        = 0x37
	KeySlash      = 0x38

	KeyCapsLock = 0x39
	KeyF1       = 0x3A
	KeyF2       = 0x3B
	KeyF3       = 0x3C
	KeyF4       = 0x3D
	KeyF5       = 0x3E
	KeyF6       = 0x3F
	KeyF7       = 0x40
	KeyF8       = 0x41
	KeyF9       = 0x42
	KeyF10      = 0x43
	KeyF11      = 0x44
	KeyF12      = 0x45

	// 导航键
	KeyPrintScreen = 0x46
	KeyScrollLock  = 0x47
	KeyPause       = 0x48
	KeyInsert      = 0x49
	KeyHome        = 0x4A
	KeyPageUp      = 0x4B
	KeyDelete      = 0x4C
	KeyEnd         = 0x4D
	KeyPageDown    = 0x4E
	KeyRightArrow  = 0x4F
	KeyLeftArrow   = 0x50
	KeyDownArrow   = 0x51
	KeyUpArrow     = 0x52

	// 数字小键盘
	KeyNumLock        = 0x53
	KeyKeypadSlash    = 0x54
	KeyKeypadAsterisk = 0x55
	KeyKeypadMinus    = 0x56
	KeyKeypadPlus     = 0x57
	KeyKeypadEnter    = 0x58
	KeyKeypad1        = 0x59
	KeyKeypad2        = 0x5A
	KeyKeypad3        = 0x5B
	KeyKeypad4        = 0x5C
	KeyKeypad5        = 0x5D
	KeyKeypad6        = 0x5E
	KeyKeypad7        = 0x5F
	KeyKeypad8        = 0x60
	KeyKeypad9        = 0x61
	KeyKeypad0        = 0x62
	KeyKeypadDot      = 0x63
	KeyKeypadEqual    = 0x67

	// 更多功能键
	KeyF13 = 0x68
	KeyF14 = 0x69
	KeyF15 = 0x6A
	KeyF16 = 0x6B
	KeyF17 = 0x6C
	KeyF18 = 0x6D
	KeyF19 = 0x6E
	KeyF20 = 0x6F
	KeyF21 = 0x70
	KeyF22 = 0x71
	KeyF23 = 0x72
	KeyF24 = 0x73

	// 特殊功能键和国际键盘相关按键
	KeyIntlBackslash = 0x64 // 国际反斜杠键
	KeyContextMenu   = 0x65 // 上下文菜单键
	KeyPower         = 0x66 // 电源键

	// 多媒体键
	KeyAudioVolumeMute = 0x7F // 音量静音
	KeyAudioVolumeUp   = 0x80 // 音量增加
	KeyAudioVolumeDown = 0x81 // 音量减少
)

// OTG 键盘修饰键
const (
	ModifierLeftCtrl   = 0x01
	ModifierLeftShift  = 0x02
	ModifierLeftAlt    = 0x04
	ModifierLeftGUI    = 0x08
	ModifierRightCtrl  = 0x10
	ModifierRightShift = 0x20
	ModifierRightAlt   = 0x40
	ModifierRightGUI   = 0x80
)

// OTG 鼠标按键
const (
	MouseLeft    = 0x01
	MouseRight   = 0x02
	MouseMiddle  = 0x04
	MouseBack    = 0x08 // Back/Up
	MouseForward = 0x10 // Forward/Down
)

// OTGKMHIDControl 实现KMHIDController接口（OTG模式）
type OTGKMHIDControl struct {
	keyboardDev      *os.File
	relativeMouseDev *os.File
	absoluteMouseDev *os.File
	absolute         bool // 鼠标模式：true为绝对模式，false为相对模式
	isOpen           bool // 设备是否处于打开状态

	// 键盘状态
	modifiers  byte   // 当前修饰键状态
	activeKeys []byte // 当前按下的按键（最多6个）

	// 鼠标状态
	mouseButtons byte // 当前鼠标按键状态
	mouseX       int  // 绝对鼠标X坐标
	mouseY       int  // 绝对鼠标Y坐标
	mouseDeltaX  int  // 相对鼠标X增量
	mouseDeltaY  int  // 相对鼠标Y增量
	mouseWheel   int8 // 鼠标滚轮状态
}

// NewOTGKMHIDControl 创建新的OTGKMHIDControl实例
func NewOTGKMHIDController() hid.KMHIDController {
	return &OTGKMHIDControl{
		absolute:     false,           // 默认使用相对鼠标模式
		isOpen:       false,           // 初始状态为关闭
		modifiers:    0x00,            // 初始修饰键状态为0
		activeKeys:   make([]byte, 0), // 初始按下的按键为空
		mouseButtons: 0x00,            // 初始鼠标按键状态为0
		mouseX:       0,               // 初始绝对鼠标X坐标为0
		mouseY:       0,               // 初始绝对鼠标Y坐标为0
		mouseDeltaX:  0,               // 初始相对鼠标X增量为0
		mouseDeltaY:  0,               // 初始相对鼠标Y增量为0
		mouseWheel:   0,               // 初始鼠标滚轮状态为0
	}
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *OTGKMHIDControl) SetAbsoluteMouse(absolute bool) error {
	log.Printf("OTG: Setting absolute mouse mode to: %v", absolute)
	// 切换模式时重置鼠标状态，避免上一次的状态影响新的模式
	d.mouseDeltaX = 0
	d.mouseDeltaY = 0
	d.mouseX = 0
	d.mouseY = 0
	d.mouseWheel = 0
	d.absolute = absolute
	return nil
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *OTGKMHIDControl) IsAbsoluteMouse() bool {
	return d.absolute
}

// open 内部方法：打开HID设备文件，添加重试机制
func (d *OTGKMHIDControl) open() error {
	var err error

	log.Printf("Opening OTG HID devices...")

	// 重试次数
	maxRetries := 5
	retryDelay := 200 * time.Millisecond

	// 打开键盘设备文件 (/dev/hidg0)，使用非阻塞模式
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening keyboard device /dev/hidg0... (attempt %d/%d)", i+1, maxRetries)
		d.keyboardDev, err = os.OpenFile("/dev/hidg0", os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			log.Printf("Successfully opened /dev/hidg0")
			break
		}
		log.Printf("Failed to open /dev/hidg0: %v, retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		if i == maxRetries-1 {
			return fmt.Errorf("failed to open /dev/hidg0 after %d retries: %w", maxRetries, err)
		}
	}

	// 打开绝对鼠标设备文件 (/dev/hidg1)，使用非阻塞模式
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening absolute mouse device /dev/hidg1... (attempt %d/%d)", i+1, maxRetries)
		d.absoluteMouseDev, err = os.OpenFile("/dev/hidg1", os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			log.Printf("Successfully opened /dev/hidg1")
			break
		}
		log.Printf("Failed to open /dev/hidg1: %v, retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		if i == maxRetries-1 {
			d.keyboardDev.Close()
			return fmt.Errorf("failed to open /dev/hidg1 after %d retries: %w", maxRetries, err)
		}
	}

	// 打开相对鼠标设备文件 (/dev/hidg2)，使用非阻塞模式
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening relative mouse device /dev/hidg2... (attempt %d/%d)", i+1, maxRetries)
		d.relativeMouseDev, err = os.OpenFile("/dev/hidg2", os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			log.Printf("Successfully opened /dev/hidg2")
			break
		}
		log.Printf("Failed to open /dev/hidg2: %v, retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		if i == maxRetries-1 {
			d.keyboardDev.Close()
			d.absoluteMouseDev.Close()
			return fmt.Errorf("failed to open /dev/hidg2 after %d retries: %w", maxRetries, err)
		}
	}

	// 设置设备为打开状态
	d.isOpen = true
	log.Printf("All OTG HID devices opened successfully")
	return nil
}

// Open 实现KMHIDController接口的Open方法，不接受参数
func (d *OTGKMHIDControl) Open() error {
	// 检查设备是否已打开
	if d.isOpen {
		return fmt.Errorf("device already open")
	}
	// 调用内部方法打开设备，忽略devicePath参数
	return d.open()
}

// Close 关闭HID设备文件，在关闭前释放所有按键
func (d *OTGKMHIDControl) Close() error {
	var err1, err2, err3 error

	// 释放所有按键和修饰键
	if d.keyboardDev != nil {
		d.SendKeyboardReport(0x00, []byte{})
		err1 = d.keyboardDev.Close()
		d.keyboardDev = nil
	}

	if d.relativeMouseDev != nil {
		err2 = d.relativeMouseDev.Close()
		d.relativeMouseDev = nil
	}

	if d.absoluteMouseDev != nil {
		err3 = d.absoluteMouseDev.Close()
		d.absoluteMouseDev = nil
	}

	// 重置设备状态
	d.isOpen = false

	// 返回第一个遇到的错误
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

// SendKeyboardReport 发送键盘HID报告，添加错误处理和重连机制
func (d *OTGKMHIDControl) SendKeyboardReport(modifier byte, keys []byte) error {
	// 检查设备是否已打开，如果未打开则尝试打开
	if !d.isOpen {
		log.Printf("Device not open, attempting to open...")
		if err := d.open(); err != nil {
			log.Printf("Failed to open devices: %v", err)
			return nil
		}
	}

	// 检查设备文件描述符是否有效
	if d.keyboardDev == nil {
		log.Printf("Keyboard device not initialized, attempting to reopen...")
		// 重新打开所有设备
		if err := d.open(); err != nil {
			log.Printf("Failed to reopen devices: %v", err)
			d.isOpen = false
			return nil
		}
	}

	// 键盘报告格式：
	// 第1字节：修饰键
	// 第2字节：保留
	// 第3-8字节：最多6个按键码
	report := []byte{modifier, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// 填充按键码（最多6个按键）
	for i := 0; i < 6; i++ {
		if i < len(keys) {
			report[2+i] = keys[i]
		} else {
			report[2+i] = 0x00
		}
	}

	log.Printf("Sending keyboard report: modifier=0x%02x, keys=%v, report=%v", modifier, keys, report)

	// 写入键盘设备文件
	_, err := d.keyboardDev.Write(report)
	if err != nil {
		log.Printf("Failed to write keyboard report: %v, attempting to reconnect...", err)
		// 关闭所有设备
		d.closeDevices()

		// 尝试重新打开所有设备
		if reopenErr := d.open(); reopenErr != nil {
			log.Printf("Failed to reopen all devices: %v", reopenErr)
			d.isOpen = false
			return nil
		}

		// 重连后增加短暂延迟，给设备时间初始化
		time.Sleep(100 * time.Millisecond)

		// 再次尝试写入
		_, err = d.keyboardDev.Write(report)
		if err != nil {
			log.Printf("Failed to write keyboard report after reconnect: %v", err)
			// 关闭所有设备并重置状态
			d.closeDevices()
			d.isOpen = false
			return nil
		}
		log.Printf("Successfully wrote keyboard report after reconnect")
	}
	return nil
}

// sendMouseReportInternal 内部方法：发送鼠标HID报告，根据内部绝对模式标志选择相对或绝对
func (d *OTGKMHIDControl) sendMouseReportInternal(buttons byte, dx, dy int, wheelY int8) error {
	// 检查设备是否已打开，如果未打开则尝试打开
	if !d.isOpen {
		log.Printf("Device not open, attempting to open...")
		if err := d.open(); err != nil {
			log.Printf("Failed to open devices: %v", err)
			return nil
		}
	}

	log.Printf("Sending mouse report: internal_absolute=%v, buttons=0x%02x, dx=%d, dy=%d, wheelY=%d", d.absolute, buttons, dx, dy, wheelY)
	if d.absolute {
		return d.SendAbsoluteMouseReport(buttons, dx, dy, wheelY)
	}
	return d.SendRelativeMouseReport(buttons, dx, dy, wheelY)
}

// SendMouseReport 实现KMHIDController接口的SendMouseReport方法
func (d *OTGKMHIDControl) SendMouseReport(buttons byte, dx, dy int, wheel int8) error {
	// 调用内部方法，传入当前的absolute状态，wheelY=wheel
	return d.sendMouseReportInternal(buttons, dx, dy, wheel)
}

// SendRelativeMouseReport 发送相对鼠标HID报告
func (d *OTGKMHIDControl) SendRelativeMouseReport(buttons byte, dx, dy int, wheelY int8) error {
	// 检查设备是否已打开，如果未打开则尝试打开
	if !d.isOpen {
		log.Printf("Device not open, attempting to open...")
		if err := d.open(); err != nil {
			log.Printf("Failed to open devices: %v", err)
			return nil
		}
	}

	// 检查设备文件描述符是否有效
	if d.relativeMouseDev == nil {
		log.Printf("Relative mouse device not initialized, attempting to reopen...")
		// 重新打开所有设备
		if err := d.open(); err != nil {
			log.Printf("Failed to reopen devices: %v", err)
			d.isOpen = false
			return nil
		}
	}

	// 相对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2字节：X轴增量
	// 第3字节：Y轴增量
	// 第4字节：垂直滚轮增量
	// 确保dx和dy在-127到127范围内
	relDx := dx
	relDy := dy

	// 确保增量在有效范围内
	if relDx < -127 {
		relDx = -127
	}
	if relDx > 127 {
		relDx = 127
	}
	if relDy < -127 {
		relDy = -127
	}
	if relDy > 127 {
		relDy = 127
	}

	// 灵敏度调整：除以3并向上取整
	if relDx > 0 {
		relDx = (relDx + 2) / 3
	} else if relDx < 0 {
		relDx = -((-relDx + 2) / 3)
	}

	if relDy > 0 {
		relDy = (relDy + 2) / 3
	} else if relDy < 0 {
		relDy = -((-relDy + 2) / 3)
	}

	// 正确处理负值，转换为无符号字节表示
	adjustedDx := byte(0)
	if relDx > 0 {
		adjustedDx = byte(relDx)
	} else if relDx < 0 {
		adjustedDx = byte(255 + relDx) // 负值转换为无符号字节
	}

	adjustedDy := byte(0)
	if relDy > 0 {
		adjustedDy = byte(relDy)
	} else if relDy < 0 {
		adjustedDy = byte(255 + relDy) // 负值转换为无符号字节
	}

	// 调整滚轮：1表示向上，255表示向下，0表示不动
	adjustedWheel := byte(0)
	if wheelY > 0 {
		adjustedWheel = 1
	} else if wheelY < 0 {
		adjustedWheel = 255
	}

	report := make([]byte, 4)
	report[0] = buttons
	report[1] = adjustedDx    // 调整后的X增量
	report[2] = adjustedDy    // 调整后的Y增量
	report[3] = adjustedWheel // 调整后的滚轮增量

	log.Printf("Sending relative mouse report: report=%v", report)

	// 写入相对鼠标设备文件 (/dev/hidg2)
	_, err := d.relativeMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write relative mouse report: %v, attempting to reconnect...", err)
		// 关闭所有设备
		d.closeDevices()

		// 尝试重新打开所有设备
		if reopenErr := d.open(); reopenErr != nil {
			log.Printf("Failed to reopen all devices: %v", reopenErr)
			d.isOpen = false
			return nil
		}

		// 重连后增加短暂延迟，给设备时间初始化
		time.Sleep(100 * time.Millisecond)

		// 再次尝试写入
		_, err = d.relativeMouseDev.Write(report)
		if err != nil {
			log.Printf("Failed to write relative mouse report after reconnect: %v", err)
			// 关闭所有设备并重置状态
			d.closeDevices()
			d.isOpen = false
			return nil
		}
		log.Printf("Successfully wrote relative mouse report after reconnect")
	}
	return nil
}

// SendAbsoluteMouseReport 发送绝对鼠标HID报告
func (d *OTGKMHIDControl) SendAbsoluteMouseReport(buttons byte, x, y int, wheelY int8) error {
	// 检查设备是否已打开，如果未打开则尝试打开
	if !d.isOpen {
		log.Printf("Device not open, attempting to open...")
		if err := d.open(); err != nil {
			log.Printf("Failed to open devices: %v", err)
			return nil
		}
	}

	// 检查设备文件描述符是否有效
	if d.absoluteMouseDev == nil {
		log.Printf("Absolute mouse device not initialized, attempting to reopen...")
		// 重新打开所有设备
		if err := d.open(); err != nil {
			log.Printf("Failed to reopen devices: %v", err)
			d.isOpen = false
			return nil
		}
	}

	// 绝对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2-3字节：X轴绝对坐标（16位）
	// 第4-5字节：Y轴绝对坐标（16位）
	// 第6字节：垂直滚轮增量
	// 确保坐标在0-65535范围内
	if x < 0 {
		x = 0
	}
	if x > 65535 {
		x = 65535
	}
	if y < 0 {
		y = 0
	}
	if y > 65535 {
		y = 65535
	}

	// 与CH9329保持一致：调整滚轮，1表示向上，255表示向下，0表示不动
	adjustedWheel := byte(0)
	if wheelY > 0 {
		adjustedWheel = 1
	} else if wheelY < 0 {
		adjustedWheel = 255
	}

	// 移除除以8的坐标调整，直接使用原始坐标，因为前端已经处理了归一化坐标
	report := make([]byte, 6)
	report[0] = buttons
	report[1] = byte(x & 0xFF) // X坐标低字节
	report[2] = byte(x >> 8)   // X坐标高字节
	report[3] = byte(y & 0xFF) // Y坐标低字节
	report[4] = byte(y >> 8)   // Y坐标高字节
	report[5] = adjustedWheel  // 调整后的滚轮增量

	log.Printf("Sending absolute mouse report: x=%d, y=%d, report=%v", x, y, report)

	// 写入绝对鼠标设备文件 (/dev/hidg1)
	_, err := d.absoluteMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write absolute mouse report: %v, attempting to reconnect...", err)
		// 关闭所有设备
		d.closeDevices()

		// 尝试重新打开所有设备
		if reopenErr := d.open(); reopenErr != nil {
			log.Printf("Failed to reopen all devices: %v", reopenErr)
			d.isOpen = false
			return nil
		}

		// 重连后增加短暂延迟，给设备时间初始化
		time.Sleep(100 * time.Millisecond)

		// 再次尝试写入
		_, err = d.absoluteMouseDev.Write(report)
		if err != nil {
			log.Printf("Failed to write absolute mouse report after reconnect: %v", err)
			// 关闭所有设备并重置状态
			d.closeDevices()
			d.isOpen = false
			return nil
		}
		log.Printf("Successfully wrote absolute mouse report after reconnect")
	}
	return nil
}

// closeDevices 关闭所有设备文件描述符
func (d *OTGKMHIDControl) closeDevices() {
	if d.keyboardDev != nil {
		d.keyboardDev.Close()
		d.keyboardDev = nil
	}
	if d.relativeMouseDev != nil {
		d.relativeMouseDev.Close()
		d.relativeMouseDev = nil
	}
	if d.absoluteMouseDev != nil {
		d.absoluteMouseDev.Close()
		d.absoluteMouseDev = nil
	}
}

// PressKey 按下单个按键
func (d *OTGKMHIDControl) PressKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{key})
}

// ReleaseKey 释放所有按键
func (d *OTGKMHIDControl) ReleaseKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{})
}

// PressKeyWithModifier 按下带有修饰键的按键
func (d *OTGKMHIDControl) PressKeyWithModifier(modifier byte, key byte) error {
	// 按下带有修饰键的按键
	if err := d.SendKeyboardReport(modifier, []byte{key}); err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	// 释放所有按键
	if err := d.SendKeyboardReport(0x00, []byte{}); err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	return nil
}

// PressKeyWithModifiers 按下带有多个修饰键和多个普通键的组合
func (d *OTGKMHIDControl) PressKeyWithModifiers(modifier byte, keys ...byte) error {
	// 按下带有修饰键的多个按键
	if err := d.SendKeyboardReport(modifier, keys); err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	// 释放所有按键
	if err := d.SendKeyboardReport(0x00, []byte{}); err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	return nil
}

// MoveMouse 移动鼠标（相对模式）
func (d *OTGKMHIDControl) MoveMouse(dx, dy int) error {
	return d.sendMouseReportInternal(0x00, dx, dy, 0)
}

// ClickMouse 点击鼠标
func (d *OTGKMHIDControl) ClickMouse(button byte) error {
	// 按下
	if err := d.sendMouseReportInternal(button, 0, 0, 0); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// 释放
	if err := d.sendMouseReportInternal(0x00, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标（垂直）
func (d *OTGKMHIDControl) ScrollMouse(wheel int8) error {
	return d.sendMouseReportInternal(0x00, 0, 0, wheel)
}

// MoveRelativeMouse 相对移动鼠标
func (d *OTGKMHIDControl) MoveRelativeMouse(dx, dy int) error {
	return d.SendRelativeMouseReport(0x00, dx, dy, 0)
}

// MoveAbsoluteMouse 移动绝对鼠标
func (d *OTGKMHIDControl) MoveAbsoluteMouse(x, y int) error {
	return d.SendAbsoluteMouseReport(0, x, y, 0)
}

// ProcessButton 处理鼠标按键事件，与CH9329保持一致
func (d *OTGKMHIDControl) ProcessButton(button byte, state bool) error {
	if state {
		d.mouseButtons |= button
	} else {
		d.mouseButtons &= ^button
	}

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	if d.absolute {
		return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
	} else {
		return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
	}
}

// ProcessMove 处理鼠标绝对移动事件，与CH9329保持一致
func (d *OTGKMHIDControl) ProcessMove(toX, toY int) error {
	d.mouseX = toX
	d.mouseY = toY

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
}

// ProcessRelativeMove 处理鼠标相对移动事件，与CH9329保持一致
func (d *OTGKMHIDControl) ProcessRelativeMove(deltaX, deltaY int) error {
	d.mouseDeltaX = deltaX
	d.mouseDeltaY = deltaY

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
}

// ProcessWheel 处理鼠标滚轮事件，与CH9329保持一致
func (d *OTGKMHIDControl) ProcessWheel(deltaY int8) error {
	d.mouseWheel = deltaY

	// 发送鼠标报告
	if d.absolute {
		return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
	} else {
		return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
	}
}

// ProcessKey 处理键盘按键事件，与CH9329保持一致
func (d *OTGKMHIDControl) ProcessKey(key byte, isModifier bool, state bool) error {
	if state {
		if isModifier {
			d.modifiers |= key
		} else if len(d.activeKeys) < 6 {
			// 检查按键是否已在活动列表中
			found := false
			for _, k := range d.activeKeys {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				d.activeKeys = append(d.activeKeys, key)
			}
		}
	} else {
		if isModifier {
			d.modifiers &= ^key
		} else {
			// 从活动列表中移除按键
			for i, k := range d.activeKeys {
				if k == key {
					d.activeKeys = append(d.activeKeys[:i], d.activeKeys[i+1:]...)
					break
				}
			}
		}
	}
	// 发送键盘报告
	return d.SendKeyboardReport(d.modifiers, d.activeKeys)
}

// ClearScreen 清屏（模拟 Ctrl+L 快捷键）
func (d *OTGKMHIDControl) ClearScreen() error {
	return d.PressKeyWithModifier(ModifierLeftCtrl, KeyL)
}

// Reboot 重启系统（模拟 Ctrl+Alt+Delete 快捷键）
func (d *OTGKMHIDControl) Reboot() error {
	// 组合 Ctrl+Alt 修饰键并显式转换为byte类型
	modifier := byte(ModifierLeftCtrl | ModifierLeftAlt)
	return d.PressKeyWithModifier(modifier, KeyDelete)
}

// TypeString 输入字符串（支持字母、数字和符号，与CH9329保持一致）
func (d *OTGKMHIDControl) TypeString(s string) error {
	for _, c := range s {
		var key byte
		var modifier byte

		// 转换字符为按键码
		switch c {
		case 'a', 'A':
			key = KeyA
		case 'b', 'B':
			key = KeyB
		case 'c', 'C':
			key = KeyC
		case 'd', 'D':
			key = KeyD
		case 'e', 'E':
			key = KeyE
		case 'f', 'F':
			key = KeyF
		case 'g', 'G':
			key = KeyG
		case 'h', 'H':
			key = KeyH
		case 'i', 'I':
			key = KeyI
		case 'j', 'J':
			key = KeyJ
		case 'k', 'K':
			key = KeyK
		case 'l', 'L':
			key = KeyL
		case 'm', 'M':
			key = KeyM
		case 'n', 'N':
			key = KeyN
		case 'o', 'O':
			key = KeyO
		case 'p', 'P':
			key = KeyP
		case 'q', 'Q':
			key = KeyQ
		case 'r', 'R':
			key = KeyR
		case 's', 'S':
			key = KeyS
		case 't', 'T':
			key = KeyT
		case 'u', 'U':
			key = KeyU
		case 'v', 'V':
			key = KeyV
		case 'w', 'W':
			key = KeyW
		case 'x', 'X':
			key = KeyX
		case 'y', 'Y':
			key = KeyY
		case 'z', 'Z':
			key = KeyZ
		case '1':
			key = Key1
		case '2':
			key = Key2
		case '3':
			key = Key3
		case '4':
			key = Key4
		case '5':
			key = Key5
		case '6':
			key = Key6
		case '7':
			key = Key7
		case '8':
			key = Key8
		case '9':
			key = Key9
		case '0':
			key = Key0
		case '-':
			key = KeyMinus
		case '=':
			key = KeyEqual
		case '[':
			key = KeyLeftBrace
		case ']':
			key = KeyRightBrace
		case '\\':
			key = KeyBackslash
		case ';':
			key = KeySemicolon
		case '\'':
			key = KeyApostrophe
		case '`':
			key = KeyGrave
		case ',':
			key = KeyComma
		case '.':
			key = KeyDot
		case '/':
			key = KeySlash
		case '!':
			key = Key1
		case '@':
			key = Key2
		case '#':
			key = Key3
		case '$':
			key = Key4
		case '%':
			key = Key5
		case '^':
			key = Key6
		case '&':
			key = Key7
		case '*':
			key = Key8
		case '(':
			key = Key9
		case ')':
			key = Key0
		case '_':
			key = KeyMinus
		case '+':
			key = KeyEqual
		case '{':
			key = KeyLeftBrace
		case '}':
			key = KeyRightBrace
		case '|':
			key = KeyBackslash
		case ':':
			key = KeySemicolon
		case '~':
			key = KeyGrave
		case '<':
			key = KeyComma
		case '>':
			key = KeyDot
		case '?':
			key = KeySlash
		case ' ':
			key = KeySpace
		case '·':
			key = KeyGrave
		case '\t':
			key = KeyTab
		case '\n':
			key = KeyEnter
		// 全角英文字母
		case 'ａ', 'Ａ':
			key = KeyA
		case 'ｂ', 'Ｂ':
			key = KeyB
		case 'ｃ', 'Ｃ':
			key = KeyC
		case 'ｄ', 'Ｄ':
			key = KeyD
		case 'ｅ', 'Ｅ':
			key = KeyE
		case 'ｆ', 'Ｆ':
			key = KeyF
		case 'ｇ', 'Ｇ':
			key = KeyG
		case 'ｈ', 'Ｈ':
			key = KeyH
		case 'ｉ', 'Ｉ':
			key = KeyI
		case 'ｊ', 'Ｊ':
			key = KeyJ
		case 'ｋ', 'Ｋ':
			key = KeyK
		case 'ｌ', 'Ｌ':
			key = KeyL
		case 'ｍ', 'Ｍ':
			key = KeyM
		case 'ｎ', 'Ｎ':
			key = KeyN
		case 'ｏ', 'Ｏ':
			key = KeyO
		case 'ｐ', 'Ｐ':
			key = KeyP
		case 'ｑ', 'Ｑ':
			key = KeyQ
		case 'ｒ', 'Ｒ':
			key = KeyR
		case 'ｓ', 'Ｓ':
			key = KeyS
		case 'ｔ', 'Ｔ':
			key = KeyT
		case 'ｕ', 'Ｕ':
			key = KeyU
		case 'ｖ', 'Ｖ':
			key = KeyV
		case 'ｗ', 'Ｗ':
			key = KeyW
		case 'ｘ', 'Ｘ':
			key = KeyX
		case 'ｙ', 'Ｙ':
			key = KeyY
		case 'ｚ', 'Ｚ':
			key = KeyZ
		// 全角数字
		case '１':
			key = Key1
		case '２':
			key = Key2
		case '３':
			key = Key3
		case '４':
			key = Key4
		case '５':
			key = Key5
		case '６':
			key = Key6
		case '７':
			key = Key7
		case '８':
			key = Key8
		case '９':
			key = Key9
		case '０':
			key = Key0
		// 全角符号
		case '－':
			key = KeyMinus
		case '＝':
			key = KeyEqual
		case '［':
			key = KeyLeftBrace
		case '］':
			key = KeyRightBrace
		case '＼':
			key = KeyBackslash
		case '；':
			key = KeySemicolon
		case '＇':
			key = KeyApostrophe
		case '｀':
			key = KeyGrave
		case '，':
			key = KeyComma
		case '．':
			key = KeyDot
		case '／':
			key = KeySlash
		case '！':
			key = Key1
		case '＠':
			key = Key2
		case '＃':
			key = Key3
		case '＄':
			key = Key4
		case '％':
			key = Key5
		case '＾':
			key = Key6
		case '＆':
			key = Key7
		case '＊':
			key = Key8
		case '（':
			key = Key9
		case '）':
			key = Key0
		case '＿':
			key = KeyMinus
		case '＋':
			key = KeyEqual
		case '｛':
			key = KeyLeftBrace
		case '｝':
			key = KeyRightBrace
		case '｜':
			key = KeyBackslash
		case '：':
			key = KeySemicolon
		case '＂':
			key = KeyApostrophe
		case '～':
			key = KeyGrave
		case '＜':
			key = KeyComma
		case '＞':
			key = KeyDot
		case '？':
			key = KeySlash
		case '　':
			key = KeySpace
		default:
			// 忽略不支持的字符
			continue
		}

		// 设置修饰键
		if c >= 'A' && c <= 'Z' {
			modifier = ModifierLeftShift
		}
		// 需要Shift键的符号
		switch c {
		case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '+', '{', '}', '|', ':', '"', '~', '<', '>', '?',
			'！', '＠', '＃', '＄', '％', '＾', '＆', '＊', '（', '）', '＿', '＋', '｛', '｝', '｜', '：', '＂', '～', '＜', '＞', '？':
			modifier = ModifierLeftShift
		}

		if err := d.SendKeyboardReport(modifier, []byte{key}); err != nil {
			return err
		}

		time.Sleep(10 * time.Millisecond)

		if err := d.SendKeyboardReport(0x00, []byte{}); err != nil {
			return err
		}

		time.Sleep(10 * time.Millisecond)
	}

	return nil
}
