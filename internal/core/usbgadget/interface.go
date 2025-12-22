package usbgadget

// -------------------------- 核心接口定义 --------------------------
// USBCompositeDevice 复合USB设备接口（ISO挂载 + HID键鼠控制）
type USBCompositeDevice interface {
	// 验证ISO镜像有效性（仅ISO场景需调用）
	ValidateISO() error
	// 初始化设备（配置Gadget + 绑定ISO + 绑定HID）
	Setup() error
	// 发送鼠标控制指令（相对坐标模式）
	SendMouseRelative(btn MouseButton, x, y int8) error
	// 发送鼠标控制指令（绝对坐标模式）
	SendMouseAbsolute(btn MouseButton, x, y int) error
	// 发送键盘控制指令
	SendKeyboard(mod KeyboardModifier, key KeyboardKey) error
	// 自动识别USB控制器名称
	GetUDCName() (string, error)
	// 清理设备资源（退出时调用）
	Cleanup() error
	// 获取当前鼠标操作模式
	GetMouseMode() MouseMode
}
