package otg

import (
	"fmt"
	"path/filepath"
)

// addEthernet 添加以太网功能
func (g *Gadget) addEthernet(start bool, driver, hostMAC, kvmMAC string) error {
	eps := 3

	// 验证MAC地址
	if hostMAC != "" && kvmMAC != "" && hostMAC == kvmMAC {
		return fmt.Errorf("Ethernet host_mac should not be equal to kvm_mac")
	}

	// 处理驱动程序
	realDriver := driver
	if driver == "rndis5" {
		realDriver = "rndis"
	}

	funcName := fmt.Sprintf("%s.usb0", realDriver)
	funcPath, err := g.createFunction(funcName)
	if err != nil {
		return err
	}

	// 配置MAC地址
	if hostMAC != "" {
		err = g.write(filepath.Join(funcPath, "host_addr"), hostMAC, false)
		if err != nil {
			return err
		}
	}

	if kvmMAC != "" {
		err = g.write(filepath.Join(funcPath, "dev_addr"), kvmMAC, false)
		if err != nil {
			return err
		}
	}

	// 配置OS描述符
	if driver == "ncm" || driver == "rndis" {
		err = g.write(filepath.Join(g.gadgetPath, "os_desc/use"), "1", false)
		if err != nil {
			return err
		}

		err = g.write(filepath.Join(g.gadgetPath, "os_desc/b_vendor_code"), "0xCD", false)
		if err != nil {
			return err
		}

		err = g.write(filepath.Join(g.gadgetPath, "os_desc/qw_sign"), "MSFT100", false)
		if err != nil {
			return err
		}

		if driver == "ncm" {
			err = g.write(filepath.Join(funcPath, "os_desc/interface.ncm/compatible_id"), "WINNCM", false)
		} else if driver == "rndis" {
			err = g.write(filepath.Join(funcPath, "os_desc/interface.rndis/compatible_id"), "RNDIS", false)
			err = g.write(filepath.Join(funcPath, "os_desc/interface.rndis/sub_compatible_id"), "5162001", false)
		}

		if err != nil {
			return err
		}

		// 创建OS描述符符号链接
		osDescPath := filepath.Join(g.gadgetPath, "os_desc", "c.1")
		err = g.symlink(g.profilePath, osDescPath)
		if err != nil {
			return err
		}
	}

	// 启动功能
	if start {
		err = g.startFunction(funcName, eps)
		if err != nil {
			return err
		}
	}

	// 创建元数据
	err = g.createMeta(funcName, "Ethernet", eps)
	if err != nil {
		return err
	}

	return nil
}
