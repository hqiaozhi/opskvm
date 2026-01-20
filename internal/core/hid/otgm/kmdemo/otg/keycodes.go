package otg

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
