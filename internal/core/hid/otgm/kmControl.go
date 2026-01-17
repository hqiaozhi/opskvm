package otgm

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 键盘按键编码
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
)

// CH9329 键盘修饰键
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

// CH9329 鼠标按键
const (
	MouseLeft   = 0x01
	MouseRight  = 0x02
	MouseMiddle = 0x04
)

// OTGKMHIDControl 实现HID鼠标和键盘控制
type OTGKMHIDControl struct {
	keyboardDev      *os.File
	relativeMouseDev *os.File
	absoluteMouseDev *os.File
	absolute         bool // 鼠标模式：true为绝对模式，false为相对模式
}

// NewOTGKMHIDControl 创建新的OTGKMHIDControl实例
func NewOTGKMHIDControl() *OTGKMHIDControl {
	return &OTGKMHIDControl{
		absolute: false, // 默认使用相对鼠标模式
	}
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *OTGKMHIDControl) SetAbsoluteMouse(absolute bool) error {
	d.absolute = absolute
	return nil
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *OTGKMHIDControl) IsAbsoluteMouse() bool {
	return d.absolute
}

// Open 打开HID设备文件，添加重试机制
func (d *OTGKMHIDControl) Open() error {
	var err error

	log.Printf("Opening OTG HID devices...")

	// 重试次数
	maxRetries := 5
	retryDelay := 200 * time.Millisecond

	// 打开键盘设备文件 (/dev/hidg0)
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening keyboard device /dev/hidg0... (attempt %d/%d)", i+1, maxRetries)
		d.keyboardDev, err = os.OpenFile("/dev/hidg0", os.O_WRONLY, 0)
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

	// 打开相对鼠标设备文件 (/dev/hidg1)
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening relative mouse device /dev/hidg1... (attempt %d/%d)", i+1, maxRetries)
		d.relativeMouseDev, err = os.OpenFile("/dev/hidg1", os.O_WRONLY, 0)
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

	// 打开绝对鼠标设备文件 (/dev/hidg2)
	for i := 0; i < maxRetries; i++ {
		log.Printf("Opening absolute mouse device /dev/hidg2... (attempt %d/%d)", i+1, maxRetries)
		d.absoluteMouseDev, err = os.OpenFile("/dev/hidg2", os.O_WRONLY, 0)
		if err == nil {
			log.Printf("Successfully opened /dev/hidg2")
			break
		}
		log.Printf("Failed to open /dev/hidg2: %v, retrying in %v...", err, retryDelay)
		time.Sleep(retryDelay)
		if i == maxRetries-1 {
			d.keyboardDev.Close()
			d.relativeMouseDev.Close()
			return fmt.Errorf("failed to open /dev/hidg2 after %d retries: %w", maxRetries, err)
		}
	}

	log.Printf("All OTG HID devices opened successfully")
	return nil
}

// Close 关闭HID设备文件
func (d *OTGKMHIDControl) Close() error {
	var err1, err2, err3 error

	if d.keyboardDev != nil {
		err1 = d.keyboardDev.Close()
	}

	if d.relativeMouseDev != nil {
		err2 = d.relativeMouseDev.Close()
	}

	if d.absoluteMouseDev != nil {
		err3 = d.absoluteMouseDev.Close()
	}

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
		// 尝试重新打开键盘设备
		newDev, reopenErr := os.OpenFile("/dev/hidg0", os.O_WRONLY, 0)
		if reopenErr != nil {
			log.Printf("Failed to reopen keyboard device: %v", reopenErr)
			return fmt.Errorf("failed to write and reopen keyboard device: %w", err)
		}
		// 关闭旧设备，使用新设备
		d.keyboardDev.Close()
		d.keyboardDev = newDev
		log.Printf("Successfully reconnected keyboard device")
		// 再次尝试写入
		_, err = d.keyboardDev.Write(report)
		if err != nil {
			log.Printf("Failed to write keyboard report after reconnect: %v", err)
			return err
		}
	}
	return nil
}

// SendMouseReport 发送鼠标HID报告，根据当前模式选择相对或绝对
func (d *OTGKMHIDControl) SendMouseReport(buttons byte, dx, dy, wheel int8) error {
	log.Printf("Sending mouse report: buttons=0x%02x, dx=%d, dy=%d, wheel=%d, absolute=%v", buttons, dx, dy, wheel, d.absolute)
	if d.absolute {
		return d.SendAbsoluteMouseReport(buttons, dx, dy, wheel)
	}
	return d.SendRelativeMouseReport(buttons, dx, dy, wheel)
}

// SendRelativeMouseReport 发送相对鼠标HID报告
func (d *OTGKMHIDControl) SendRelativeMouseReport(buttons byte, dx, dy, wheel int8) error {
	// 相对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2字节：X轴增量
	// 第3字节：Y轴增量
	// 第4字节：滚轮增量
	report := []byte{buttons, byte(dx), byte(dy), byte(wheel)}

	log.Printf("Sending relative mouse report: report=%v", report)

	// 写入相对鼠标设备文件 (/dev/hidg1)
	_, err := d.relativeMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write relative mouse report: %v, attempting to reconnect...", err)
		// 尝试重新打开相对鼠标设备
		newDev, reopenErr := os.OpenFile("/dev/hidg1", os.O_WRONLY, 0)
		if reopenErr != nil {
			log.Printf("Failed to reopen relative mouse device: %v", reopenErr)
			return fmt.Errorf("failed to write and reopen relative mouse device: %w", err)
		}
		// 关闭旧设备，使用新设备
		d.relativeMouseDev.Close()
		d.relativeMouseDev = newDev
		log.Printf("Successfully reconnected relative mouse device")
		// 再次尝试写入
		_, err = d.relativeMouseDev.Write(report)
		if err != nil {
			log.Printf("Failed to write relative mouse report after reconnect: %v", err)
			return err
		}
	}
	return nil
}

// SendAbsoluteMouseReport 发送绝对鼠标HID报告
func (d *OTGKMHIDControl) SendAbsoluteMouseReport(buttons byte, x, y, wheel int8) error {
	// 绝对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2-3字节：X轴绝对坐标（16位）
	// 第4-5字节：Y轴绝对坐标（16位）
	// 第6字节：滚轮增量
	absX := uint16(x)
	absY := uint16(y)
	report := make([]byte, 6)
	report[0] = buttons
	report[1] = byte(absX & 0xFF) // X坐标低字节
	report[2] = byte(absX >> 8)   // X坐标高字节
	report[3] = byte(absY & 0xFF) // Y坐标低字节
	report[4] = byte(absY >> 8)   // Y坐标高字节
	report[5] = byte(wheel)       // 滚轮增量

	log.Printf("Sending absolute mouse report: x=%d, y=%d, report=%v", x, y, report)

	// 写入绝对鼠标设备文件 (/dev/hidg2)
	_, err := d.absoluteMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write absolute mouse report: %v, attempting to reconnect...", err)
		// 尝试重新打开绝对鼠标设备
		newDev, reopenErr := os.OpenFile("/dev/hidg2", os.O_WRONLY, 0)
		if reopenErr != nil {
			log.Printf("Failed to reopen absolute mouse device: %v", reopenErr)
			return fmt.Errorf("failed to write and reopen absolute mouse device: %w", err)
		}
		// 关闭旧设备，使用新设备
		d.absoluteMouseDev.Close()
		d.absoluteMouseDev = newDev
		log.Printf("Successfully reconnected absolute mouse device")
		// 再次尝试写入
		_, err = d.absoluteMouseDev.Write(report)
		if err != nil {
			log.Printf("Failed to write absolute mouse report after reconnect: %v", err)
			return err
		}
	}
	return nil
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

	// 释放所有按键
	if err := d.SendKeyboardReport(0x00, []byte{}); err != nil {
		return err
	}

	return nil
}

// PressKeyWithModifiers 按下带有多个修饰键和多个普通键的组合
func (d *OTGKMHIDControl) PressKeyWithModifiers(modifier byte, keys ...byte) error {
	// 按下带有修饰键的多个按键
	if err := d.SendKeyboardReport(modifier, keys); err != nil {
		return err
	}

	// 释放所有按键
	if err := d.SendKeyboardReport(0x00, []byte{}); err != nil {
		return err
	}

	return nil
}

// MoveMouse 移动鼠标
func (d *OTGKMHIDControl) MoveMouse(dx, dy int8) error {
	return d.SendMouseReport(0x00, dx, dy, 0)
}

// ClickMouse 点击鼠标
func (d *OTGKMHIDControl) ClickMouse(button byte) error {
	// 按下
	if err := d.SendMouseReport(button, 0, 0, 0); err != nil {
		return err
	}

	// 释放
	if err := d.SendMouseReport(0x00, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标
func (d *OTGKMHIDControl) ScrollMouse(wheel int8) error {
	return d.SendMouseReport(0x00, 0, 0, wheel)
}

// MoveRelativeMouse 相对移动鼠标
func (d *OTGKMHIDControl) MoveRelativeMouse(dx, dy int8) error {
	return d.SendRelativeMouseReport(0x00, dx, dy, 0)
}

// MoveAbsoluteMouse 移动绝对鼠标
func (d *OTGKMHIDControl) MoveAbsoluteMouse(x, y int8) error {
	return d.SendAbsoluteMouseReport(0, x, y, 0)
}

// ClearScreen 清屏（模拟 Ctrl+L 快捷键）
func (d *OTGKMHIDControl) ClearScreen() error {
	// 按下 Ctrl+L
	if err := d.SendKeyboardReport(0x01, []byte{0x38}); err != nil {
		return err
	}
	// 释放所有按键
	return d.SendKeyboardReport(0x00, []byte{})
}

// Reboot 重启系统（模拟 Ctrl+Alt+Delete 快捷键）
func (d *OTGKMHIDControl) Reboot() error {
	// 按下 Ctrl+Alt+Delete
	if err := d.SendKeyboardReport(0x01|0x04, []byte{0x4c}); err != nil {
		return err
	}
	// 释放所有按键
	return d.SendKeyboardReport(0x00, []byte{})
}
