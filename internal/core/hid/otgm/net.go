package otgm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NetInterface interface {
	AddNET() (nic string, err error)
	getNIC() error
}

type NET struct {
	GadgetInterface
	funcPath string
	nic      string
}

func NewNET(gadget GadgetInterface) NetInterface {
	return &NET{
		GadgetInterface: gadget,
	}
}

func (n *NET) AddNET() (nic string, err error) {

	// 创建功能
	// DriverName rndis 兼容老版本的windows系统，性能较低
	// DriverName ncm 跨平台支持，这里使用ncm模式 （配置不一样，这里仅仅支持ncm模式）
	DriverName := "ncm"
	funcName := fmt.Sprintf("%s.usb0", DriverName)
	funcPath, err := n.CreateFunction(funcName)
	if err != nil {
		return "", err
	}
	n.funcPath = funcPath

	err = n.Write(filepath.Join(funcPath, "os_desc/interface.ncm/compatible_id"), "WINNCM")
	if err != nil {
		return "", err
	}
	// 启动功能
	err = n.StartFunction(funcName)
	if err != nil {
		return "", err
	}

	// 获取网卡名
	err = n.getNIC()
	if err != nil {
		return "", err
	}

	return n.nic, nil
}

func (n *NET) getNIC() error {
	path := filepath.Join(n.funcPath, "ifname")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 去除换行符并赋值
	n.nic = strings.TrimSpace(string(content))
	if n.nic == "" {
		return fmt.Errorf("网卡名称为空，文件路径: %s", path)
	}
	fmt.Printf("网卡: %s", n.nic)
	return nil
}
