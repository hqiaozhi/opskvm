package otgdevices

import (
	"fmt"
)

// ExampleStartOTG 演示如何使用StartOTG函数启动OTG设备
func ExampleStartOTG() {
	// 创建配置
	config := GadgetConfig{
		VendorID:      0x1D6B, // Linux Foundation
		ProductID:     0x0104, // Multifunction Composite Gadget
		USBVersion:    0x0200, // USB 2.0
		DeviceVersion: 0x0100, // Version 1.0
		Manufacturer:  "OpsKVM Composite",
		Product:       "OTG Device",
		Serial:        "0000001",
		Config:        "OTG Configuration",
		MaxPower:      500,
		RemoteWakeup:  true,
		InitDelay:     0,
		Meta:          "/tmp/otg-meta",
		Endpoints:     32,
		UDC:           "", // 自动查找UDC
		Gadget:        "g1",
		SysFSPrefix:   "/",
		Devices: DevicesConfig{
			HID: HIDConfig{
				Keyboard: HIDDeviceConfig{Start: false},
				Mouse:    HIDDeviceConfig{Start: false},
			},
			MSD: MSDConfig{
				Start: true,
				Default: MSDDefaultConfig{
					CDROM: true,
					RW:    true,
				},
			},
			Ethernet: EthernetConfig{
				Enabled: false,
				Driver:  "rndis",
				HostMAC: "00:11:22:33:44:55",
				KVMMAC:  "00:11:22:33:44:56",
				Start:   false,
			},
			Serial: SerialConfig{
				Enabled: false,
				Start:   false,
			},
		},
	}

	// 打印配置信息（避免未使用变量错误）
	fmt.Printf("OTG设备配置: 制造商=%s, 产品=%s\n", config.Manufacturer, config.Product)

	// 启动OTG设备（需要root权限）
	err := StartOTG(config)
	if err != nil {
		fmt.Printf("Failed to start OTG device: %v\n", err)
		return
	}

	fmt.Println("OTG device started successfully")

	// 停止OTG设备（实际使用时根据需要调用）
	// err = StopOTG(config)
	// if err != nil {
	//     fmt.Printf("Failed to stop OTG device: %v\n", err)
	//     return
	// }

	// fmt.Println("OTG device stopped successfully")
}

// ExamplevalidateMAC 演示如何使用validateMAC函数验证MAC地址
func ExamplevalidateMAC() {
	macAddresses := []string{
		"00:11:22:33:44:55", // 有效
		"AA:BB:CC:DD:EE:FF", // 有效
		"00:11:22:33:44",    // 无效（长度错误）
		"00:11:22:33:44:GG", // 无效（字符错误）
		"",                  // 有效（空字符串）
	}

	for _, mac := range macAddresses {
		if validateMAC(mac) {
			fmt.Printf("%s is a valid MAC address\n", mac)
		} else {
			fmt.Printf("%s is NOT a valid MAC address\n", mac)
		}
	}
}

// ExamplegetGadgetPath 演示如何使用getGadgetPath函数构建路径
func ExamplegetGadgetPath() {
	// 构建默认路径
	path1 := getGadgetPath("/", "g1")
	fmt.Printf("Default path: %s\n", path1)

	// 构建带有配置的路径
	path2 := getGadgetPath("/", "g1", "configs", "c.1")
	fmt.Printf("Path with config: %s\n", path2)

	// 构建带有自定义sysfs前缀的路径
	path3 := getGadgetPath("/sysroot", "g1")
	fmt.Printf("Path with custom prefix: %s\n", path3)
}
