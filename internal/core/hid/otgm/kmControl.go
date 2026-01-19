package otgm

import (
	"fmt"
	"log"
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

// HIDDevice 定义HID设备接口（与CH9329保持一致）
type HIDDevice interface {
	Open(devicePath string) error
	Close() error
	SendKeyboardReport(modifier byte, keys []byte) error
	SendMouseReport(absolute bool, buttons byte, dx, dy, wheelX, wheelY int8) error
	PressKey(key byte) error
	ReleaseKey(key byte) error
	PressKeyWithModifier(modifier byte, key byte) error
	PressKeyWithModifiers(modifier byte, keys ...byte) error
	ClearScreen() error
	Reboot() error
	TypeString(s string) error
	MoveMouse(dx, dy int8) error
	ClickMouse(button byte) error
	ScrollMouse(wheel int8) error
}

// OTGKMHIDControl 实现HIDDevice接口（OTG模式）
type OTGKMHIDControl struct {
	keyboardDev      *os.File
	relativeMouseDev *os.File
	absoluteMouseDev *os.File
	absolute         bool // 鼠标模式：true为绝对模式，false为相对模式
	isOpen           bool // 设备是否处于打开状态
}

// NewOTGKMHIDControl 创建新的OTGKMHIDControl实例
func NewOTGKMHIDControl() *OTGKMHIDControl {
	return &OTGKMHIDControl{
		absolute: false, // 默认使用相对鼠标模式
		isOpen:   false, // 初始状态为关闭
	}
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *OTGKMHIDControl) SetAbsoluteMouse(absolute bool) error {
	log.Printf("OTG: Setting absolute mouse mode to: %v", absolute)
	d.absolute = absolute
	return nil
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *OTGKMHIDControl) IsAbsoluteMouse() bool {
	return d.absolute
}

// open 内部方法：打开HID设备文件，添加重试机制
func (d *OTGKMHIDControl) open(devicePath string) error {
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
	return d.open("")
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
	// 检查设备是否已打开
	if !d.isOpen {
		return fmt.Errorf("device not open")
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
		// 尝试重新打开键盘设备
		newDev, reopenErr := os.OpenFile("/dev/hidg0", os.O_WRONLY, 0)
		if reopenErr != nil {
			log.Printf("Failed to reopen keyboard device: %v", reopenErr)
			return nil // 忽略，因为设备可能不存在
		}
		// 关闭旧设备，使用新设备
		d.keyboardDev.Close()
		d.keyboardDev = newDev
		log.Printf("Successfully reconnected keyboard device")
		// 再次尝试写入
		_, err = d.keyboardDev.Write(report)
		if err != nil {
			log.Printf("Failed to write keyboard report after reconnect: %v", err)
			return nil // 忽略，因为设备可能不存在
		}
	}
	return nil
}

// sendMouseReportInternal 内部方法：发送鼠标HID报告，根据内部绝对模式标志选择相对或绝对，支持水平滚动
func (d *OTGKMHIDControl) sendMouseReportInternal(buttons byte, dx, dy int, wheelX, wheelY int8) error {
	// 检查设备是否已打开
	if !d.isOpen {
		return fmt.Errorf("device not open")
	}

	log.Printf("Sending mouse report: internal_absolute=%v, buttons=0x%02x, dx=%d, dy=%d, wheelX=%d, wheelY=%d", d.absolute, buttons, dx, dy, wheelX, wheelY)
	if d.absolute {
		return d.SendAbsoluteMouseReport(buttons, dx, dy, wheelX, wheelY)
	}
	return d.SendRelativeMouseReport(buttons, dx, dy, wheelX, wheelY)
}

// SendMouseReport 实现KMHIDController接口的SendMouseReport方法
func (d *OTGKMHIDControl) SendMouseReport(buttons byte, dx, dy int, wheel int8) error {
	// 调用内部方法，传入当前的absolute状态，wheelY=wheel，wheelX=0
	return d.sendMouseReportInternal(buttons, dx, dy, 0, wheel)
}

// SendRelativeMouseReport 发送相对鼠标HID报告，支持水平滚动
func (d *OTGKMHIDControl) SendRelativeMouseReport(buttons byte, dx, dy int, wheelX, wheelY int8) error {
	// 相对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2字节：X轴增量
	// 第3字节：Y轴增量
	// 第4字节：垂直滚轮增量
	// 第5字节：水平滚轮增量（固定包含，与设备描述符一致）
	// 限制dx和dy在-127到127范围内
	reldx := int8(dx)
	if reldx < -127 {
		reldx = -127
	}
	if reldx > 127 {
		reldx = 127
	}

	reldy := int8(dy)
	if reldy < -127 {
		reldy = -127
	}
	if reldy > 127 {
		reldy = 127
	}

	report := make([]byte, 5)
	report[0] = buttons
	report[1] = byte(reldx)  // 在相对模式下，dx是相对增量，直接转换为byte
	report[2] = byte(reldy)  // 在相对模式下，dy是相对增量，直接转换为byte
	report[3] = byte(wheelY) // 垂直滚轮增量
	report[4] = byte(wheelX) // 水平滚轮增量（固定包含，与设备描述符一致）

	log.Printf("Sending relative mouse report: report=%v", report)

	// 写入相对鼠标设备文件 (/dev/hidg1)
	_, err := d.relativeMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write relative mouse report: %v, attempting to reopen device...", err)
		// 关闭当前设备
		d.relativeMouseDev.Close()
		d.relativeMouseDev = nil

		// 尝试重新打开相对鼠标设备，设置非阻塞模式
		newDev, reopenErr := os.OpenFile("/dev/hidg1", os.O_WRONLY|syscall.O_NONBLOCK, 0644)
		if reopenErr != nil {
			log.Printf("Failed to reopen relative mouse device: %v", reopenErr)
			// 不返回错误，让调用者继续执行
			return nil
		}

		// 使用新设备
		d.relativeMouseDev = newDev
		log.Printf("Successfully reopened relative mouse device")

		// 不再次尝试写入，让调用者重试
		return nil
	}
	return nil
}

// SendAbsoluteMouseReport 发送绝对鼠标HID报告，支持水平滚动
func (d *OTGKMHIDControl) SendAbsoluteMouseReport(buttons byte, x, y int, wheelX, wheelY int8) error {
	// 绝对鼠标模式报告格式：
	// 第1字节：按钮状态
	// 第2-3字节：X轴绝对坐标（16位）
	// 第4-5字节：Y轴绝对坐标（16位）
	// 第6字节：垂直滚轮增量
	// 第7字节：水平滚轮增量（固定包含，与设备描述符一致）
	// 归一化计算：处理客户端发送的坐标
	// 检查客户端发送的坐标范围：
	// 如果x在0-65535范围内，直接使用
	// 否则，将-32768-32767范围转换为0-65535范围
	var absX, absY int
	if x >= 0 && x <= 65535 {
		// 客户端发送的是0-65535范围的绝对坐标，直接使用
		absX = x
		absY = y
	} else {
		// 客户端发送的是-32768-32767范围的绝对坐标，转换为0-65535范围
		absX = x + 32768
		absY = y + 32768
	}
	if absX < 0 {
		absX = 0
	}
	if absX > 65535 {
		absX = 65535
	}
	if absY < 0 {
		absY = 0
	}
	if absY > 65535 {
		absY = 65535
	}

	report := make([]byte, 7)
	report[0] = buttons
	report[1] = byte(absX & 0xFF) // X坐标低字节
	report[2] = byte(absX >> 8)   // X坐标高字节
	report[3] = byte(absY & 0xFF) // Y坐标低字节
	report[4] = byte(absY >> 8)   // Y坐标高字节
	report[5] = byte(wheelY)      // 垂直滚轮增量
	report[6] = byte(wheelX)      // 总是包含水平滚轮字节，与设备描述符一致

	log.Printf("Sending absolute mouse report: x=%d, y=%d, absX=%d, absY=%d, report=%v", x, y, absX, absY, report)

	// 写入绝对鼠标设备文件 (/dev/hidg2)
	_, err := d.absoluteMouseDev.Write(report)
	if err != nil {
		log.Printf("Failed to write absolute mouse report: %v, attempting to reopen device...", err)
		// 关闭当前设备
		d.absoluteMouseDev.Close()
		d.absoluteMouseDev = nil

		// 尝试重新打开绝对鼠标设备，设置非阻塞模式
		newDev, reopenErr := os.OpenFile("/dev/hidg2", os.O_WRONLY|syscall.O_NONBLOCK, 0644)
		if reopenErr != nil {
			log.Printf("Failed to reopen absolute mouse device: %v", reopenErr)
			// 不返回错误，让调用者继续执行
			return nil
		}

		// 使用新设备
		d.absoluteMouseDev = newDev
		log.Printf("Successfully reopened absolute mouse device")

		// 不再次尝试写入，让调用者重试
		return nil
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
	return d.sendMouseReportInternal(0x00, dx, dy, 0, 0)
}

// ClickMouse 点击鼠标
func (d *OTGKMHIDControl) ClickMouse(button byte) error {
	// 按下
	if err := d.sendMouseReportInternal(button, 0, 0, 0, 0); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// 释放
	if err := d.sendMouseReportInternal(0x00, 0, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标（垂直）
func (d *OTGKMHIDControl) ScrollMouse(wheel int8) error {
	return d.sendMouseReportInternal(0x00, 0, 0, 0, wheel)
}

// MoveRelativeMouse 相对移动鼠标
func (d *OTGKMHIDControl) MoveRelativeMouse(dx, dy int) error {
	return d.SendRelativeMouseReport(0x00, dx, dy, 0, 0)
}

// MoveAbsoluteMouse 移动绝对鼠标
func (d *OTGKMHIDControl) MoveAbsoluteMouse(x, y int) error {
	return d.SendAbsoluteMouseReport(0, x, y, 0, 0)
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
		case '"':
			key = KeyApostrophe
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
		case '\n':
			key = KeyEnter
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
		case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '+', '{', '}', '|', ':', '"', '~', '<', '>', '?':
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
