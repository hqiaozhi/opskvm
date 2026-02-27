package ch9329

import (
	"time"
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

// SendKeyboardReport 发送键盘HID报告，与Python版本完全一致
func (d *CH9329Device) SendKeyboardReport(modifier byte, keys []byte) error {
	// CH9329 键盘命令格式，与Python完全一致：
	// [0, 0x02, 0x08, modifier, 0, key1, key2, key3, key4, key5, key6]
	cmd := []byte{
		0x00,            // 保留字节
		CmdTypeKeyboard, // 命令类型：键盘
		0x08,            // 数据长度
		modifier,        // 修饰键
		0x00,            // 保留字节
		0x00,            // 按键1
		0x00,            // 按键2
		0x00,            // 按键3
		0x00,            // 按键4
		0x00,            // 按键5
		0x00,            // 按键6
	}

	// 填充按键码（最多6个按键），与Python完全一致
	for i := 0; i < 6 && i < len(keys); i++ {
		cmd[5+i] = keys[i]
	}

	// 发送命令
	return d.sendCommand(cmd)
}

// SetLEDs 设置LED状态，与Python版本一致
func (d *CH9329Device) SetLEDs(ledByte byte) {
	d.ledState = ledByte
}

// LEDStatus LED状态
func (d *CH9329Device) LEDStatus() map[string]bool {
	return map[string]bool{
		"num":    (d.ledState & 1) != 0,
		"caps":   ((d.ledState >> 1) & 1) != 0,
		"scroll": ((d.ledState >> 2) & 1) != 0,
	}
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

// ProcessKey 处理键盘按键事件，与Python版本一致
func (d *CH9329Device) ProcessKey(key byte, isModifier bool, state bool) error {
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

// TypeString 输入字符串（仅支持字母、数字和部分符号）
func (d *CH9329Device) TypeString(s string) error {
	for _, c := range s {
		var key byte

		// 转换字符为按键码
		switch c {
		// 半角英文字母
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
		// 半角数字
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
		// 半角符号
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
		// 全角字符（转换为对应的半角字符处理）
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
		// 控制字符
		case ' ':
			key = KeySpace
		case '\t':
			key = KeyTab
		case '\n':
			key = KeyEnter
		default:
			// 忽略不支持的字符
			continue
		}

		// 按下并释放按键
		modifier := byte(0x00)
		// 半角或全角大写字母需要Shift键
		if (c >= 'A' && c <= 'Z') || (c >= 'Ａ' && c <= 'Ｚ') {
			modifier = ModifierLeftShift
		}
		// 需要Shift键的半角符号
		switch c {
		case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '+', '{', '}', '|', ':', '"', '~', '<', '>', '?':
			modifier = ModifierLeftShift
		}
		// 需要Shift键的全角符号
		switch c {
		case '！', '＠', '＃', '＄', '％', '＾', '＆', '＊', '（', '）', '＿', '＋', '｛', '｝', '｜', '：', '＂', '～', '＜', '＞', '？':
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
