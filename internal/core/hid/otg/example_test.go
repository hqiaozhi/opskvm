package otg

import (
	"fmt"
	"testing"
)

// ExampleStartOTG 演示如何使用StartOTG函数启动OTG设备
func TestExampleStartOTG(t *testing.T) {
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
