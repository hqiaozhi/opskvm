package otg

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StartOTG 启动OTG设备
func StartOTG(config GadgetConfig) error {
	// 设置默认值
	if config.SysFSPrefix == "" {
		config.SysFSPrefix = "/"
	}

	// 查找UDC
	udc, err := findUDC(config.SysFSPrefix, config.UDC)
	if err != nil {
		return err
	}

	logger := func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}

	logger("Using UDC %s", udc)

	// 创建Gadget
	logger("Creating gadget %q ...", config.Gadget)
	gadgetPath := getGadgetPath(config.SysFSPrefix, config.Gadget)

	g := NewGadget(gadgetPath, filepath.Join(gadgetPath, "configs/c.1"), config.Meta, config.Endpoints, logger)

	// 创建元数据目录
	err = g.mkdir(config.Meta)
	if err != nil {
		return err
	}

	// 创建设备目录
	err = g.mkdir(gadgetPath)
	if err != nil {
		return err
	}

	// 设置设备信息
	err = g.write(filepath.Join(gadgetPath, "idVendor"), fmt.Sprintf("0x%04X", config.VendorID), false)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(gadgetPath, "idProduct"), fmt.Sprintf("0x%04X", config.ProductID), false)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(gadgetPath, "bcdUSB"), fmt.Sprintf("0x%04X", config.USBVersion), false)
	if err != nil {
		return err
	}

	// 设置设备版本
	deviceVersion := config.DeviceVersion
	if deviceVersion < 0 {
		deviceVersion = 0x0100
		if config.Devices.Ethernet.Enabled {
			if config.Devices.Ethernet.Driver == "ncm" {
				deviceVersion = 0x0102
			} else if config.Devices.Ethernet.Driver == "rndis" {
				deviceVersion = 0x0101
			}
		}
	}

	err = g.write(filepath.Join(gadgetPath, "bcdDevice"), fmt.Sprintf("0x%04X", deviceVersion), false)
	if err != nil {
		return err
	}

	// 配置字符串描述符
	langPath := filepath.Join(gadgetPath, "strings/0x409")
	err = g.mkdir(langPath)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(langPath, "manufacturer"), config.Manufacturer, false)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(langPath, "product"), config.Product, false)
	if err != nil {
		return err
	}

	if config.Serial != "" {
		err = g.write(filepath.Join(langPath, "serialnumber"), config.Serial, false)
		if err != nil {
			return err
		}
	}

	// 创建配置文件
	profilePath := filepath.Join(gadgetPath, "configs/c.1")
	err = g.mkdir(profilePath)
	if err != nil {
		return err
	}

	if config.Config != "" {
		configLangPath := filepath.Join(profilePath, "strings/0x409")
		err = g.mkdir(configLangPath)
		if err != nil {
			return err
		}

		err = g.write(filepath.Join(configLangPath, "configuration"), config.Config, false)
		if err != nil {
			return err
		}
	}

	err = g.write(filepath.Join(profilePath, "MaxPower"), strconv.Itoa(config.MaxPower), false)
	if err != nil {
		return err
	}

	if config.RemoteWakeup {
		err = g.write(filepath.Join(profilePath, "bmAttributes"), "0xA0", false)
		if err != nil {
			return err
		}
	}

	// 添加功能
	cod := config.Devices

	if config.KVMD.HID.Type == "otg" {
		logger("===== HID-Keyboard =====")
		err = g.addKeyboard(cod.HID.Keyboard.Start, config.RemoteWakeup)
		if err != nil {
			return err
		}

		logger("===== HID-Mouse =====")
		ckhm := config.KVMD.HID.Mouse
		err = g.addMouse(cod.HID.Mouse.Start, config.RemoteWakeup, ckhm.Absolute, ckhm.HorizontalWheel)
		if err != nil {
			return err
		}

		if config.KVMD.HID.MouseAlt.Device != "" {
			logger("===== HID-Mouse-Alt =====")
			err = g.addMouse(cod.HID.MouseAlt.Start, config.RemoteWakeup, !ckhm.Absolute, ckhm.HorizontalWheel)
			if err != nil {
				return err
			}
		}
	}

	// 添加以太网功能
	if cod.Ethernet.Enabled {
		logger("===== Ethernet =====")
		err = g.addEthernet(
			cod.Ethernet.Start,
			cod.Ethernet.Driver,
			cod.Ethernet.HostMAC,
			cod.Ethernet.KVMMAC,
		)
		if err != nil {
			return err
		}
	}

	// 添加串口功能
	if cod.Serial.Enabled {
		logger("===== Serial =====")
		err = g.addSerial(cod.Serial.Start)
		if err != nil {
			return err
		}
	}

	// 添加大容量MSD设备
	if cod.MSD.Start {
		logger("===== MSD =====")
		err = g.addMSD(cod.MSD.Start)
		if err != nil {
			return err
		}
		if cod.MSD.Default.CDROM {
			logger("Setting MSD to CD-ROM mode ...")
			err = g.write(filepath.Join(gadgetPath, "functions/mass_storage.usb0/lun.0/cdrom"), "1\n", false)
			if err != nil {
				return err
			}
		}

		if cod.MSD.Default.RW {
			logger("Setting MSD to read-write mode ...")
			err = g.write(filepath.Join(gadgetPath, "functions/mass_storage.usb0/lun.0/ro"), "0\n", false)
			if err != nil {
				return err
			}
		}
	}

	logger("===== Preparing complete =====")

	// 启用Gadget
	logger("Enabling the gadget ...")
	err = g.write(filepath.Join(gadgetPath, "UDC"), udc, false)
	if err != nil {
		return err
	}

	// 延迟初始化
	time.Sleep(time.Duration(config.InitDelay) * time.Millisecond)

	logger("Ready to work")
	return nil
}

// StopOTG 停止OTG设备
func StopOTG(config GadgetConfig) error {
	// 设置默认值
	if config.SysFSPrefix == "" {
		config.SysFSPrefix = "/"
	}

	logger := func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}

	gadgetPath := getGadgetPath(config.SysFSPrefix, config.Gadget)

	g := NewGadget(gadgetPath, filepath.Join(gadgetPath, "configs/c.1"), config.Meta, config.Endpoints, logger)

	// 禁用Gadget
	logger("Disabling gadget %q ...", config.Gadget)
	err := g.write(filepath.Join(gadgetPath, "UDC"), "\n", false)
	if err != nil {
		return err
	}

	// 删除OS描述符符号链接
	g.unlink(filepath.Join(gadgetPath, "os_desc", "c.1"), true)

	// 删除功能符号链接
	profilePath := filepath.Join(gadgetPath, "configs/c.1")
	entries, err := os.ReadDir(profilePath)
	if err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".usb0") || strings.HasSuffix(entry.Name(), ".usb1") || strings.HasSuffix(entry.Name(), ".usb2") {
				g.unlink(filepath.Join(profilePath, entry.Name()), false)
			}
		}
	}

	// 删除配置字符串目录
	g.rmdir(filepath.Join(profilePath, "strings/0x409"))

	// 删除配置目录
	g.rmdir(profilePath)

	// 删除功能目录
	funcsPath := filepath.Join(gadgetPath, "functions")
	entries, err = os.ReadDir(funcsPath)
	if err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".usb0") || strings.HasSuffix(entry.Name(), ".usb1") || strings.HasSuffix(entry.Name(), ".usb2") {
				g.rmdir(filepath.Join(funcsPath, entry.Name()))
			}
		}
	}

	// 删除设备字符串目录
	g.rmdir(filepath.Join(gadgetPath, "strings/0x409"))

	// 删除Gadget目录
	g.rmdir(gadgetPath)

	// 删除元数据文件
	if config.Meta != "" {
		entries, err = os.ReadDir(config.Meta)
		if err == nil {
			for _, entry := range entries {
				g.unlink(filepath.Join(config.Meta, entry.Name()), false)
			}

			// 删除元数据目录
			g.rmdir(config.Meta)
		}
	}

	logger("Bye-bye")
	return nil
}
