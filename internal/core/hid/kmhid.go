package hid

// KMHIDController 定义键盘鼠标控制器的统一接口
type KMHIDController interface {
	// 基础设备操作
	Open() error
	Close() error

	// 键盘操作
	SendKeyboardReport(modifier byte, keys []byte) error
	PressKey(key byte) error
	ReleaseKey(key byte) error
	PressKeyWithModifier(modifier byte, key byte) error
	PressKeyWithModifiers(modifier byte, keys ...byte) error

	// 鼠标操作
	SendMouseReport(buttons byte, dx, dy, wheel int8) error
	MoveMouse(dx, dy int8) error
	ClickMouse(button byte) error
	ScrollMouse(wheel int8) error

	// 鼠标模式切换
	SetAbsoluteMouse(absolute bool) error
	IsAbsoluteMouse() bool

	// 共有功能方法
	ClearScreen() error
	Reboot() error
}
