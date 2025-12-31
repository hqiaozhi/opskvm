package otg

type GadgetInterface interface {
	// 基础操作
	mkdir(path string) error
	write(path string, value string, optional bool) error
	writeBytes(path string, data []byte) error
	symlink(src, dest string) error
	rmdir(path string) error
	unlink(path string, optional bool) error
	createFunction(funcName string) (string, error)
	startFunction(funcName string, eps int) error
	createMeta(funcName, funcDesc string, eps int) error
	// HID设备管理
	// 通用HID设备添加
	addHID(desc string, start bool, remoteWakeup bool, hid Hid) error

	// 键盘
	addKeyboard(start bool, remoteWakeup bool) error
	// 鼠标
	addMouse(start bool, remoteWakeup bool, absolute bool, horizontalWheel bool) error

	// 网络设备管理
	addEthernet(start bool, driver, hostMAC, kvmMAC string) error
	// 大容量存储设备管理
	addMSD(start bool) error
	// 串口设备管理
	addSerial(start bool) error
}
