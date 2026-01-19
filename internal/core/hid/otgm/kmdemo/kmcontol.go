package kmdemo

import (
	"fmt"
	"os"
	"syscall"
	"time"
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

// OTGDevice 实现HIDDevice接口（OTG模式）
type OTGDevice struct {
	fd         *os.File
	devicePath string
	absolute   bool // 鼠标是否使用绝对模式
}

// NewOTGDevice 创建新的OTG设备实例
func NewOTGDevice() *OTGDevice {
	return &OTGDevice{
		fd:       nil,
		absolute: false, // 默认使用相对鼠标模式
	}
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *OTGDevice) SetAbsoluteMouse(absolute bool) {
	d.absolute = absolute
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *OTGDevice) IsAbsoluteMouse() bool {
	return d.absolute
}

// Open 打开OTG设备
func (d *OTGDevice) Open(devicePath string) error {
	if d.fd != nil {
		return fmt.Errorf("device already open")
	}

	// 打开设备文件（非阻塞读写模式）
	fd, err := os.OpenFile(devicePath, os.O_RDWR|syscall.O_NONBLOCK, 0644)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}

	d.fd = fd
	d.devicePath = devicePath
	return nil
}

// Close 关闭OTG设备
func (d *OTGDevice) Close() error {
	if d.fd == nil {
		return nil
	}

	// 释放所有按键和修饰键
	d.SendKeyboardReport(0x00, []byte{})

	err := d.fd.Close()
	if err != nil {
		return fmt.Errorf("failed to close device: %w", err)
	}

	d.fd = nil
	d.devicePath = ""
	return nil
}

// SendKeyboardReport 发送键盘HID报告（与One-KVM的make_keyboard_report保持一致）
func (d *OTGDevice) SendKeyboardReport(modifier byte, keys []byte) error {
	if d.fd == nil {
		return fmt.Errorf("device not open")
	}

	// OTG 键盘报告格式（8字节）：
	// [modifier, 0, key1, key2, key3, key4, key5, key6]
	report := make([]byte, 8)
	report[0] = modifier

	// 填充按键码（最多6个按键）
	for i := 0; i < 6; i++ {
		if i < len(keys) {
			report[2+i] = keys[i]
		} else {
			report[2+i] = 0x00
		}
	}

	// 写入设备
	n, err := d.fd.Write(report)
	if err != nil {
		return fmt.Errorf("failed to send keyboard report: %w", err)
	}

	if n != len(report) {
		return fmt.Errorf("incomplete keyboard report sent: %d/%d bytes", n, len(report))
	}

	return nil
}

// SendMouseReport 发送鼠标HID报告（与One-KVM的make_mouse_report保持一致）
func (d *OTGDevice) SendMouseReport(absolute bool, buttons byte, dx, dy, wheelX, wheelY int8) error {
	if d.fd == nil {
		return fmt.Errorf("device not open")
	}

	var report []byte

	if absolute {
		// 绝对鼠标模式报告格式（6字节）：
		// [buttons, x_low, x_high, y_low, y_high, wheel]
		x := uint16(dx)
		y := uint16(dy)
		report = []byte{
			buttons,
			byte(x & 0xFF),
			byte(x >> 8),
			byte(y & 0xFF),
			byte(y >> 8),
			byte(wheelY),
		}
		if wheelX != 0 {
			// 如果有水平滚动，则添加一个字节
			report = append(report, byte(wheelX))
		}
	} else {
		// 相对鼠标模式报告格式（4或5字节）：
		// [buttons, dx, dy, wheel]
		report = []byte{
			buttons,
			byte(dx),
			byte(dy),
			byte(wheelY),
		}
		if wheelX != 0 {
			// 如果有水平滚动，则添加一个字节
			report = append(report, byte(wheelX))
		}
	}

	// 写入设备
	n, err := d.fd.Write(report)
	if err != nil {
		return fmt.Errorf("failed to send mouse report: %w", err)
	}

	if n != len(report) {
		return fmt.Errorf("incomplete mouse report sent: %d/%d bytes", n, len(report))
	}

	return nil
}

// PressKey 按下单个按键
func (d *OTGDevice) PressKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{key})
}

// ReleaseKey 释放所有按键
func (d *OTGDevice) ReleaseKey(key byte) error {
	return d.SendKeyboardReport(0x00, []byte{})
}

// PressKeyWithModifier 按下带有修饰键的按键
func (d *OTGDevice) PressKeyWithModifier(modifier byte, key byte) error {
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
func (d *OTGDevice) ClearScreen() error {
	return d.PressKeyWithModifier(ModifierLeftCtrl, KeyL)
}

// PressKeyWithModifiers 按下带有多个修饰键和多个普通键的组合
func (d *OTGDevice) PressKeyWithModifiers(modifier byte, keys ...byte) error {
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
func (d *OTGDevice) Reboot() error {
	// 组合 Ctrl+Alt 修饰键并显式转换为byte类型
	modifier := byte(ModifierLeftCtrl | ModifierLeftAlt)
	return d.PressKeyWithModifier(modifier, KeyDelete)
}

// TypeString 输入字符串（支持字母、数字和符号，与CH9329保持一致）
func (d *OTGDevice) TypeString(s string) error {
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

// MoveMouse 移动鼠标（相对模式）
func (d *OTGDevice) MoveMouse(dx, dy int8) error {
	return d.SendMouseReport(d.absolute, 0x00, dx, dy, 0, 0)
}

// ClickMouse 点击鼠标
func (d *OTGDevice) ClickMouse(button byte) error {
	// 按下
	if err := d.SendMouseReport(d.absolute, button, 0, 0, 0, 0); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// 释放
	if err := d.SendMouseReport(d.absolute, 0x00, 0, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标
func (d *OTGDevice) ScrollMouse(wheel int8) error {
	return d.SendMouseReport(d.absolute, 0x00, 0, 0, 0, wheel)
}
