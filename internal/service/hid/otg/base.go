package otg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gLog "github.com/gogf/gf/v2/frame/g"
)

type GadgetInterface interface {
	// 基础操作
	Mkdir(path string) error
	Write(path string, value string) error
	WriteBytes(path string, data []byte) error
	Read(path string) (string, error)
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
func New(udcName, udcControllerName string) GadgetInterface {
	return &Gadget{
		Name:           udcName,
		UDCControlName: udcControllerName,
		gadgetPath:     filepath.Join("/sys/kernel/config/usb_gadget", udcName),
	}
}

// Mkdir 创建目录
func (g *Gadget) Mkdir(path string) error {
	// 直接使用 MkdirAll 创建目录，无需先检查是否存在
	return os.MkdirAll(path, 0755)
}

// Write 写入文件
func (g *Gadget) Write(path string, value string) error {
	// 写入文件，不存在则创建
	return os.WriteFile(path, []byte(value), 0644)
}

// WriteBytes 写入字节数据
func (g *Gadget) WriteBytes(path string, data []byte) error {
	// 写入文件，不存在则创建
	return os.WriteFile(path, data, 0644)
}

// Read 读取文件内容
func (g *Gadget) Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Symlink 创建符号链接
func (g *Gadget) Symlink(src, dest string) error {
	return os.Symlink(src, dest)
}

// Rmdir 删除目录
func (g *Gadget) Rmdir(path string) error {
	// 使用 RemoveAll 删除目录及其所有内容
	return os.RemoveAll(path)
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
	gLog.Log().Info(context.Background(), "Gadget: Starting USB Gadget initialization")

	if err := g.Remove(); err != nil {
		gLog.Log().Errorf(context.Background(), "Gadget: Failed to remove old Gadget directory: %v", err)
	}
	time.Sleep(5 * time.Second)

	// ===================================================
	type config struct {
		BcdDevice string
		IdProduct string
		IdVendor  string
		BcdUSB    string

		Manufacturer string
		Product      string
		Serialnumber string

		Configuration string

		MaxPower     string
		BmAttributes string
	}
	conf := config{
		// 设备信息
		BcdDevice: "0x0100", // Version 1.0
		BcdUSB:    "0x0200", // USB 2.0
		IdVendor:  "0x1d6b", // Linux Foundation
		IdProduct: "0x0104", // Multifunction Composite Gadget

		// 产品信息
		Manufacturer: "OpsKVM",                  // 产品制造商
		Product:      "OpsKVM Composite Device", // 产品名称
		Serialnumber: "156491324564",            // 产品序列号

		// c.1配置信息
		BmAttributes: "0xa0", // 功能描述
		MaxPower:     "500",  // 最大功率
	}

	// 创建Gadget根目录
	gLog.Log().Infof(context.Background(), "Creating Gadget root directory: %s", g.gadgetPath)
	err := g.Mkdir(g.gadgetPath)
	if err != nil {
		return "", logError("Failed to create Gadget root directory: %w", err)
	}
	gLog.Log().Infof(context.Background(), "Successfully created Gadget root directory: %s", g.gadgetPath)

	// 给系统时间更新目录结构
	time.Sleep(100 * time.Millisecond)

	// ===================================================
	// 设置USB描述符
	gLog.Log().Info(context.Background(), "Setting USB descriptors...")
	gadgetDescriptors := map[string]string{
		"idVendor":  conf.IdVendor,
		"idProduct": conf.IdProduct,
		"bcdUSB":    conf.BcdUSB,
		"bcdDevice": conf.BcdDevice,
	}

	for desc, value := range gadgetDescriptors {
		path := filepath.Join(g.gadgetPath, desc)
		err = g.Write(path, value)
		if err != nil {
			return "", logError("Failed to write %s: %w", desc, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	gLog.Log().Info(context.Background(), "Successfully set USB descriptors")

	// ===================================================
	// 创建字符串描述符目录
	stringsPath := filepath.Join(g.gadgetPath, "strings", "0x409")
	err = g.Mkdir(stringsPath)
	if err != nil {
		return "", logError("Failed to create strings/0x409 directory: %w", err)
	}
	gLog.Log().Info(context.Background(), "Successfully created strings directory structure")
	time.Sleep(100 * time.Millisecond)

	// 写入字符串描述符
	gLog.Log().Info(context.Background(), "Writing string descriptors...")
	stringDescriptors := map[string]string{
		"manufacturer": conf.Manufacturer,
		"product":      conf.Product,
		"serialnumber": conf.Serialnumber,
	}

	for file, value := range stringDescriptors {
		path := filepath.Join(stringsPath, file)
		err = g.Write(path, value)
		if err != nil {
			return "", logError("Failed to write %s: %w", file, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	gLog.Log().Info(context.Background(), "Successfully wrote string descriptors")

	// 创建配置目录c.1
	configPath := filepath.Join(g.gadgetPath, "configs", "c.1")
	err = g.Mkdir(configPath)
	if err != nil {
		return "", logError("Failed to create configuration directory: %w", err)
	}
	gLog.Log().Infof(context.Background(), "Successfully created configuration directory: %s", configPath)
	time.Sleep(100 * time.Millisecond)

	// 写入c.1配置
	// ==================================================================================
	// 写入MaxPower
	err = g.Write(filepath.Join(configPath, "MaxPower"), conf.MaxPower)
	if err != nil {
		return "", logError("Failed to write MaxPower: %w", err)
	}
	time.Sleep(50 * time.Millisecond)

	// bmAttributes：0xa0表示总线供电，支持远程唤醒
	err = g.Write(filepath.Join(configPath, "bmAttributes"), conf.BmAttributes)
	if err != nil {
		return "", logError("Failed to write bmAttributes: %w", err)
	}

	time.Sleep(2 * time.Second)

	gLog.Log().Info(context.Background(), "USB Gadget initialization completed successfully")
	return g.Name, nil
}

// logError 记录错误并返回
func logError(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	gLog.Log().Errorf(context.Background(), "%v", err)
	return err
}

// CreateFunction 创建功能目录
func (g *Gadget) CreateFunction(funcName string) (string, error) {
	funcPath := filepath.Join(g.gadgetPath, "functions", funcName)

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
	gLog.Log().Infof(context.Background(), "Successfully created function directory: %s", funcPath)
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
	udcPath := filepath.Join(g.gadgetPath, "UDC")
	maxRetries := 3
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		// 启动UDC设备
		gLog.Log().Infof(context.Background(), "Starting UDC device: %s (attempt %d/%d)", g.UDCControlName, i+1, maxRetries)
		err := g.Write(udcPath, g.UDCControlName)
		if err != nil {
			gLog.Log().Errorf(context.Background(), "Failed to start UDC: %v", err)
			if i < maxRetries-1 {
				gLog.Log().Infof(context.Background(), "Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return err
		}

		// 给控制器一些时间初始化
		time.Sleep(500 * time.Millisecond)
		gLog.Log().Infof(context.Background(), "UDC device started successfully: %s", g.UDCControlName)
		return nil
	}

	return fmt.Errorf("Gadget: Failed to start UDC after %d attempts", maxRetries)
}

func (g *Gadget) CloseUDC() error {
	udcPath := filepath.Join(g.gadgetPath, "UDC")

	// 先检查文件是否存在
	if _, err := os.Stat(udcPath); os.IsNotExist(err) {
		// 文件不存在，直接返回nil，不报错
		return nil
	}

	// 先打开文件看看是否有内容，如果为空则跳过
	content, err := os.ReadFile(udcPath)
	if err != nil {
		// 读取文件失败，返回错误
		return fmt.Errorf("读取UDC文件失败：%w", err)
	}
	if len(content) < 2 {
		// 内容为空，跳过关闭
		return nil
	}

	// 关闭UDC设备，写入换行符
	err = g.Write(udcPath, "\n")
	if err != nil {
		return err
	}
	// 给控制器足够的时间重置
	time.Sleep(3 * time.Second)
	return nil
}

func (g *Gadget) Remove() error {
	gLog.Log().Info(context.Background(), "Removing old Gadget directory")

	// 判断目录是否存在，不存在则跳过
	if _, err := os.Stat(g.gadgetPath); os.IsNotExist(err) {
		gLog.Log().Info(context.Background(), "Removing Gadget root directory successfully")
		return nil
	}

	// 关闭UDC设备
	err := g.CloseUDC()
	if err != nil {
		return err
	}
	gLog.Log().Info(context.Background(), "UDC device closed successfully")

	// 删除功能符号链接
	profilePath := filepath.Join(g.gadgetPath, "configs/c.1")
	entries, err := os.ReadDir(profilePath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(entry.Name(), "usb") {
				g.Unlink(filepath.Join(profilePath, entry.Name()))
			}
		}
	}
	gLog.Log().Info(context.Background(), "Removing function symlinks successfully")

	// 删除配置字符串目录
	g.Rmdir(filepath.Join(profilePath, "strings/0x409"))
	gLog.Log().Info(context.Background(), "Remove 'strings/0x409' directory successfully")

	// 删除配置目录
	g.Rmdir(profilePath)
	gLog.Log().Infof(context.Background(), "Removing '%s' directory successfully", "configs/c.1")

	// 删除功能目录
	funcsPath := filepath.Join(g.gadgetPath, "functions")
	entries, err = os.ReadDir(funcsPath)
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "hid.usb") {
				g.Rmdir(filepath.Join(funcsPath, entry.Name()))
				gLog.Log().Infof(context.Background(), "Remove '%s' directory successfully", entry.Name())
			}
		}
	}

	// 删除设备字符串目录
	g.Rmdir(filepath.Join(g.gadgetPath, "strings/0x409"))
	gLog.Log().Info(context.Background(), "Remove 'strings/0x409' directory successfully")

	// 删除Gadget目录
	g.Rmdir(g.gadgetPath)
	gLog.Log().Infof(context.Background(), "Removing Gadget root %s directory successfully", g.gadgetPath)

	return nil
}
