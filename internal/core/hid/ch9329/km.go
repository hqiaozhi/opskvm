package ch9329

import (
	"time"

	"go.bug.st/serial"
)

// CH9329 命令类型
const (
	CmdTypeKeyboard      = 0x02
	CmdTypeMouseAbsolute = 0x04
	CmdTypeMouseRelative = 0x05
)

// CH9329 键盘按键编码（部分常用键）
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

// HIDDevice 定义HID设备接口
type HIDDevice interface {
	Open(portName string, baudRate int) error
	Close() error
	SendKeyboardReport(modifier byte, keys []byte) error
	SendMouseReport(buttons byte, dx, dy, wheel int8) error
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

// CH9329Device 实现HIDDevice接口
type CH9329Device struct {
	port     serial.Port
	absolute bool
}

// NewCH9329Device 创建新的CH9329设备实例
func NewCH9329() *CH9329Device {
	return &CH9329Device{
		absolute: false, // 默认使用相对鼠标模式
	}
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *CH9329Device) SetAbsoluteMouse(absolute bool) {
	d.absolute = absolute
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *CH9329Device) IsAbsoluteMouse() bool {
	return d.absolute
}

// calculateChecksum 计算校验和
func (d *CH9329Device) calculateChecksum(data []byte) byte {
	var sum int
	for _, b := range data {
		sum += int(b)
	}
	return byte(sum % 256)
}

// Open 打开串口连接
func (d *CH9329Device) Open(portName string, baudRate int) error {
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return err
	}
	d.port = port
	return nil
}

// Close 关闭串口连接
func (d *CH9329Device) Close() error {
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}

// SendKeyboardReport 发送键盘HID报告
func (d *CH9329Device) SendKeyboardReport(modifier byte, keys []byte) error {
	// CH9329 键盘命令格式（参考One-KVM）：
	// 头部(2字节) + [0, 0x02, 0x08, modifier, 0, key1, key2, key3, key4, key5, key6] + 校验和(1字节)
	cmd := make([]byte, 14)
	cmd[0] = 0x57            // 起始字节1
	cmd[1] = 0xAB            // 起始字节2
	cmd[2] = 0x00            // 保留
	cmd[3] = CmdTypeKeyboard // 命令类型
	cmd[4] = 0x08            // 数据长度
	cmd[5] = modifier        // 修饰键
	cmd[6] = 0x00            // 保留

	// 填充按键码（最多6个按键）
	for i := 0; i < 6; i++ {
		if i < len(keys) {
			cmd[7+i] = keys[i]
		} else {
			cmd[7+i] = 0x00
		}
	}

	// 计算校验和
	checksum := d.calculateChecksum(cmd[:13])
	cmd[13] = checksum

	_, err := d.port.Write(cmd)
	return err
}

// SendMouseReport 发送鼠标HID报告
func (d *CH9329Device) SendMouseReport(buttons byte, dx, dy, wheel int8) error {
	var cmd []byte

	if d.absolute {
		// 绝对鼠标模式（参考One-KVM）：
		// 头部(2字节) + [0, 0x04, 0x07, 0x02, buttons, x_low, x_high, y_low, y_high, wheel] + 校验和(1字节)
		// 注意：这里的dx和dy被当作绝对坐标值处理
		x := uint16(dx)
		y := uint16(dy)

		cmd = make([]byte, 14)
		cmd[0] = 0x57                 // 起始字节1
		cmd[1] = 0xAB                 // 起始字节2
		cmd[2] = 0x00                 // 保留
		cmd[3] = CmdTypeMouseAbsolute // 命令类型
		cmd[4] = 0x07                 // 数据长度
		cmd[5] = 0x02                 // 绝对模式标识
		cmd[6] = buttons              // 按键状态
		cmd[7] = byte(x & 0xFF)       // X坐标低字节
		cmd[8] = byte(x >> 8)         // X坐标高字节
		cmd[9] = byte(y & 0xFF)       // Y坐标低字节
		cmd[10] = byte(y >> 8)        // Y坐标高字节
		cmd[11] = byte(wheel)         // 滚轮
		cmd[12] = 0x00                // 保留

	} else {
		// 相对鼠标模式（参考One-KVM）：
		// 头部(2字节) + [0, 0x05, 0x05, 0x01, buttons, dx, dy, wheel] + 校验和(1字节)
		cmd = make([]byte, 12)
		cmd[0] = 0x57                 // 起始字节1
		cmd[1] = 0xAB                 // 起始字节2
		cmd[2] = 0x00                 // 保留
		cmd[3] = CmdTypeMouseRelative // 命令类型
		cmd[4] = 0x05                 // 数据长度
		cmd[5] = 0x01                 // 相对模式标识
		cmd[6] = buttons              // 按键状态
		cmd[7] = byte(dx)             // X增量
		cmd[8] = byte(dy)             // Y增量
		cmd[9] = byte(wheel)          // 滚轮
		cmd[10] = 0x00                // 保留
	}

	// 计算校验和
	checksum := d.calculateChecksum(cmd[:len(cmd)-1])
	cmd[len(cmd)-1] = checksum

	_, err := d.port.Write(cmd)
	return err
}

// PressKey 按下单个按键
func (d *CH9329Device) PressKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{key})
}

// ReleaseKey 释放所有按键
func (d *CH9329Device) ReleaseKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{})
}

// PressKeyWithModifier 按下带有修饰键的按键
func (d *CH9329Device) PressKeyWithModifier(modifier byte, key byte) error {
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

// ClearScreen 清屏（模拟 Ctrl+L 快捷键）
func (d *CH9329Device) ClearScreen() error {
	return d.PressKeyWithModifier(ModifierLeftCtrl, KeyL)
}

// PressKeyWithModifiers 按下带有多个修饰键和多个普通键的组合
func (d *CH9329Device) PressKeyWithModifiers(modifier byte, keys ...byte) error {
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

// Reboot 重启系统（模拟 Ctrl+Alt+Delete 快捷键）
func (d *CH9329Device) Reboot() error {
	// 组合 Ctrl+Alt 修饰键并显式转换为byte类型
	modifier := byte(ModifierLeftCtrl | ModifierLeftAlt)
	return d.PressKeyWithModifier(modifier, KeyDelete)
}

// TypeString 输入字符串（仅支持字母、数字和部分符号）
func (d *CH9329Device) TypeString(s string) error {
	for _, c := range s {
		var key byte

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

		// 按下并释放按键
		modifier := byte(0x00)
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

// MoveMouse 移动鼠标
func (d *CH9329Device) MoveMouse(dx, dy int8) error {
	return d.SendMouseReport(0x00, dx, dy, 0)
}

// ClickMouse 点击鼠标
func (d *CH9329Device) ClickMouse(button byte) error {
	// 按下
	if err := d.SendMouseReport(button, 0, 0, 0); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// 释放
	if err := d.SendMouseReport(0x00, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标
func (d *CH9329Device) ScrollMouse(wheel int8) error {
	return d.SendMouseReport(0x00, 0, 0, wheel)
}
