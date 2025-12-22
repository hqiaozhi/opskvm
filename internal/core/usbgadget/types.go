package usbgadget

// -------------------------- 枚举与类型定义 --------------------------
// MouseButton 鼠标按键枚举
type MouseButton uint8

const (
	MouseButtonNone  MouseButton = 0x00 // 无按键
	MouseButtonLeft  MouseButton = 0x01 // 左键
	MouseButtonRight MouseButton = 0x02 // 右键
	MouseButtonMid   MouseButton = 0x04 // 中键
)

// KeyboardModifier 键盘修饰键枚举
type KeyboardModifier uint8

const (
	KeyboardModifierNone  KeyboardModifier = 0x00 // 无修饰键
	KeyboardModifierCtrl  KeyboardModifier = 0x01 // Ctrl
	KeyboardModifierShift KeyboardModifier = 0x02 // Shift
	KeyboardModifierAlt   KeyboardModifier = 0x04 // Alt
	KeyboardModifierGUI   KeyboardModifier = 0x08 // Win/Cmd
)

// KeyboardKey 键盘按键枚举（标准HID扫描码映射）
type KeyboardKey string

const (
	KeyA     KeyboardKey = "a"
	KeyB     KeyboardKey = "b"
	KeyEnter KeyboardKey = "enter"
	KeySpace KeyboardKey = "space"
	KeyUp    KeyboardKey = "up"
	KeyDown  KeyboardKey = "down"
	KeyLeft  KeyboardKey = "left"
	KeyRight KeyboardKey = "right"
	KeyNone  KeyboardKey = ""
	// 可扩展更多按键...
)

// MouseMode 鼠标操作模式枚举
type MouseMode int

const (
	MouseModeRelative MouseMode = iota // 相对坐标模式
	MouseModeAbsolute                  // 绝对坐标模式
)
