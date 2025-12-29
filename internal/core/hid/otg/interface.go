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
	addKeyboard(start bool, remoteWakeup bool) error
	addMouse(start bool, remoteWakeup bool, absolute bool, horizontalWheel bool) error
	addHID(desc string, start bool, remoteWakeup bool, hid Hid) error
	addEthernet(start bool, driver, hostMAC, kvmMAC string) error
	addMSD(start bool) error
	addSerial(start bool) error
}
