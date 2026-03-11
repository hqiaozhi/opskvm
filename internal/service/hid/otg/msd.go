package otg

import (
	"errors"
	"path/filepath"
	"time"
)

type MSDInterface interface {
	AddMSD() error
	Bind(absoltePath, cdrom string) error
	GetPath() (string, error)
	Remove() error
}

type MSD struct {
	GadgetInterface
	funcName string
	funcPath string
	cdrom    string
	ro       string
}

func NewMSD(gadget GadgetInterface) MSDInterface {
	return &MSD{
		GadgetInterface: gadget,
	}
}

func (m *MSD) GetPath() (string, error) {
	content, err := m.Read(filepath.Join(m.funcPath, "lun.0/file"))
	if err != nil {
		return "", err
	}
	return content, nil
}

// AddMSD 创建功能
func (m *MSD) AddMSD() error {
	m.funcName = "mass_storage.usb0"
	path, err := m.CreateFunction(m.funcName)
	if err != nil {
		return err
	}
	m.funcPath = path

	err = m.StartFunction(m.funcName)
	if err != nil {
		return err
	}
	return nil
}

// Bind 绑定镜像
func (m *MSD) Bind(absoltePath, cdrom string) error {
	// cdrom = 0 磁盘(flash) 可读可写
	// cdrom = 1 光驱(cd/dvd) 只读
	m.cdrom = cdrom
	switch cdrom {
	case "0":
		m.ro = "0" // 可写
	case "1":
		m.ro = "1" // 只读
	default:
		return errors.New("not surpported cdrom value")
	}

	// 写入配置
	path1 := filepath.Join(m.funcPath, "lun.0/cdrom")
	err := m.Write(path1, "\n")
	if err != nil {
		return err
	}
	err = m.Write(path1, m.cdrom)
	if err != nil {
		return err
	}
	path2 := filepath.Join(m.funcPath, "lun.0/ro")
	err = m.Write(path2, "\n")
	if err != nil {
		return err
	}
	err = m.Write(path2, m.ro)
	if err != nil {
		return err
	}
	path3 := filepath.Join(m.funcPath, "lun.0/file")
	err = m.Write(path3, "\n")
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	err = m.Write(path3, absoltePath)
	if err != nil {
		return err
	}
	return nil
}

// Remove 移除绑定
func (m *MSD) Remove() error {
	path := filepath.Join(m.funcPath, "lun.0/file")
	return m.Write(path, "\n")
}
