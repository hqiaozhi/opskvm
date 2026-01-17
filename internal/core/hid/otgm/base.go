package otgm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GadgetInterface interface {
	// 基础操作
	Mkdir(path string) error
	Write(path string, value string) error
	WriteBytes(path string, data []byte) error
	Symlink(src, dest string) error
	Rmdir(path string) error
	Unlink(path string) error

	// 初始化配置(usb产品信息)
	InitConfig() (string, error)

	// 功能管理
	CreateFunction(funcName string) (string, error)
	StartFunction(funcName string) error

	// UDC设备管理
	StartUDC() error
	CloseUDC() error

	// Remove 删除USB Gadget配置
	Remove() error
}

// Gadget 管理USB Gadget的配置
type Gadget struct {
	Name           string
	UDCControlName string
	gadgetPath     string
}

// New 创建一个新的Gadget实例
func New(name, udcName string) GadgetInterface {
	return &Gadget{
		Name:           name,
		UDCControlName: udcName,
		gadgetPath:     filepath.Join("/sys/kernel/config/usb_gadget", name),
	}
}

// Mkdir 创建目录
func (g *Gadget) Mkdir(path string) error {
	// 直接使用 MkdirAll 创建目录，无需先检查是否存在
	return os.MkdirAll(path, 0755)
}

// Write 写入文件
func (g *Gadget) Write(path string, value string) error {
	// 直接写入文件，无需先检查是否存在
	// 使用 0644 权限，确保文件可读写
	return os.WriteFile(path, []byte(value), 0644)
}

// WriteBytes 写入字节数据
func (g *Gadget) WriteBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// Symlink 创建符号链接
func (g *Gadget) Symlink(src, dest string) error {
	return os.Symlink(src, dest)
}

// Rmdir 删除目录
func (g *Gadget) Rmdir(path string) error {
	return os.Remove(path)
}

// Unlink 删除文件或符号链接
func (g *Gadget) Unlink(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

// InitConfig 初始化配置
func (g *Gadget) InitConfig() (string, error) {
	type config struct {
		VendorID      string
		ProductID     string
		USBVersion    string
		Manufacturer  string
		Product       string
		Serial        string
		Configuration string
		MaxPower      string
	}
	conf := config{
		VendorID:      "0x1D6B",            // Linux Foundation
		ProductID:     "0x0104",            // Multifunction Composite Gadget
		USBVersion:    "0x0200",            // USB 2.0
		Manufacturer:  "OpsKVM Composite",  // 产品制造商
		Product:       "OTG Device",        // 产品名称
		Serial:        "0000001",           // 产品序列号
		Configuration: "OTG Configuration", // 配置名称
		MaxPower:      "500",               // 最大功率
	}

	log.Printf("Initializing USB Gadget at path: %s", g.gadgetPath)

	// 检查configfs是否挂载
	configfsPath := "/sys/kernel/config"
	if _, err := os.Stat(configfsPath); os.IsNotExist(err) {
		return "", logError("configfs not mounted at %s", configfsPath)
	}
	log.Printf("configfs is mounted at %s", configfsPath)

	// 首先创建Gadget根目录
	log.Printf("Creating Gadget root directory: %s", g.gadgetPath)
	err := g.Mkdir(g.gadgetPath)
	if err != nil {
		return "", logError("Failed to create Gadget root directory: %w", err)
	}
	log.Printf("Successfully created Gadget root directory: %s", g.gadgetPath)

	// 设置USB描述符
	log.Printf("Setting USB descriptors...")
	err = g.Write(filepath.Join(g.gadgetPath, "idVendor"), conf.VendorID)
	if err != nil {
		return "", logError("Failed to write idVendor: %w", err)
	}
	err = g.Write(filepath.Join(g.gadgetPath, "idProduct"), conf.ProductID)
	if err != nil {
		return "", logError("Failed to write idProduct: %w", err)
	}
	err = g.Write(filepath.Join(g.gadgetPath, "bcdUSB"), conf.USBVersion)
	if err != nil {
		return "", logError("Failed to write bcdUSB: %w", err)
	}
	log.Printf("Successfully set USB descriptors")

	// 创建字符串描述符目录
	log.Printf("Creating strings directory...")
	stringsPath := filepath.Join(g.gadgetPath, "strings")
	err = g.Mkdir(stringsPath)
	if err != nil {
		return "", logError("Failed to create strings directory: %w", err)
	}
	strings0x409Path := filepath.Join(stringsPath, "0x409")
	err = g.Mkdir(strings0x409Path)
	if err != nil {
		return "", logError("Failed to create strings/0x409 directory: %w", err)
	}
	log.Printf("Successfully created strings directory structure")

	// 写入字符串描述符
	log.Printf("Writing string descriptors...")
	err = g.Write(filepath.Join(strings0x409Path, "manufacturer"), conf.Manufacturer)
	if err != nil {
		return "", logError("Failed to write manufacturer: %w", err)
	}
	err = g.Write(filepath.Join(strings0x409Path, "product"), conf.Product)
	if err != nil {
		return "", logError("Failed to write product: %w", err)
	}
	err = g.Write(filepath.Join(strings0x409Path, "serialnumber"), conf.Serial)
	if err != nil {
		return "", logError("Failed to write serialnumber: %w", err)
	}
	log.Printf("Successfully wrote string descriptors")

	// 创建functions目录
	functionsPath := filepath.Join(g.gadgetPath, "functions")
	log.Printf("Creating functions directory: %s", functionsPath)
	err = g.Mkdir(functionsPath)
	if err != nil {
		return "", logError("Failed to create functions directory: %w", err)
	}
	log.Printf("Successfully created functions directory: %s", functionsPath)

	// 创建配置
	log.Printf("Creating configuration directory...")
	configPath := filepath.Join(g.gadgetPath, "configs", "c.1")
	err = g.Mkdir(configPath)
	if err != nil {
		return "", logError("Failed to create configuration directory: %w", err)
	}
	log.Printf("Successfully created configuration directory: %s", configPath)

	// 写入配置
	log.Printf("Writing configuration...")
	err = g.Write(filepath.Join(configPath, "MaxPower"), conf.MaxPower)
	if err != nil {
		return "", logError("Failed to write MaxPower: %w", err)
	}
	err = g.Write(filepath.Join(configPath, "bmAttributes"), "0xA0")
	if err != nil {
		return "", logError("Failed to write bmAttributes: %w", err)
	}
	configStringsPath := filepath.Join(configPath, "strings/0x409")
	err = g.Mkdir(configStringsPath)
	if err != nil {
		return "", logError("Failed to create config strings directory: %w", err)
	}
	err = g.Write(filepath.Join(configStringsPath, "configuration"), conf.Configuration)
	if err != nil {
		return "", logError("Failed to write configuration string: %w", err)
	}
	log.Printf("Successfully wrote configuration")

	log.Printf("USB Gadget initialization completed successfully")
	return g.Name, nil
}

// logError 记录错误并返回
func logError(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	log.Printf("Error: %v", err)
	return err
}

// CreateFunction 创建功能目录
func (g *Gadget) CreateFunction(funcName string) (string, error) {
	funcPath := filepath.Join(g.gadgetPath, "functions", funcName)
	log.Printf("Creating function directory: %s", funcPath)

	// 检查functions目录是否存在
	functionsPath := filepath.Join(g.gadgetPath, "functions")
	if _, err := os.Stat(functionsPath); os.IsNotExist(err) {
		return "", logError("functions directory does not exist: %s", functionsPath)
	}

	// 创建功能目录
	err := g.Mkdir(funcPath)
	if err != nil {
		return "", logError("Failed to create function directory: %w", err)
	}
	log.Printf("Successfully created function directory: %s", funcPath)
	return funcPath, nil
}

// StartFunction 启动功能
func (g *Gadget) StartFunction(funcName string) error {
	funcPath := filepath.Join(g.gadgetPath, "functions", funcName)

	destPath := filepath.Join(g.gadgetPath, "configs/c.1", funcName)
	if _, err := os.Stat(destPath); err == nil {
		err = os.Remove(destPath)
	}

	err := g.Symlink(funcPath, destPath)
	if err != nil {
		return err
	}
	return nil
}

func (g *Gadget) StartUDC() error {
	err := g.Write(filepath.Join(g.gadgetPath, "UDC"), g.UDCControlName)
	if err != nil {
		return err
	}
	return nil
}

func (g *Gadget) CloseUDC() error {
	// 关闭UDC设备
	err := g.Write(filepath.Join(g.gadgetPath, "UDC"), "\n")
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (g *Gadget) Remove() error {
	// 关闭UDC设备
	err := g.CloseUDC()
	if err != nil {
		return err
	}

	// 删除功能符号链接
	profilePath := filepath.Join(g.gadgetPath, "configs/c.1")
	entries, err := os.ReadDir(profilePath)
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "hid") {
				g.Unlink(filepath.Join(profilePath, entry.Name()))
			}
		}
	}

	// 删除配置字符串目录
	g.Rmdir(filepath.Join(profilePath, "strings/0x409"))

	// 删除配置目录
	g.Rmdir(profilePath)

	// 删除功能目录
	funcsPath := filepath.Join(g.gadgetPath, "functions")
	entries, err = os.ReadDir(funcsPath)
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "hid.usb") {
				g.Rmdir(filepath.Join(funcsPath, entry.Name()))
			}
		}
	}

	// 删除设备字符串目录
	g.Rmdir(filepath.Join(g.gadgetPath, "strings/0x409"))

	// 删除Gadget目录
	g.Rmdir(g.gadgetPath)

	return nil
}
