package usbgadget

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// LinuxUSBGadget Linux USB Gadget实现（核心结构体）
type LinuxUSBGadget struct {
	isoPath         string    // ISO镜像路径（空则仅启用HID）
	gadgetName      string    // Gadget名称
	udcName         string    // USB控制器名称（自动识别/手动指定）
	mouseMode       MouseMode // 鼠标操作模式
	relHidReportLen int       // 相对鼠标HID报告长度
	absHidReportLen int       // 绝对鼠标HID报告长度
	gadgetDir       string    // Gadget配置目录
	relHidDataPath  string    // 相对鼠标HID数据写入路径
	absHidDataPath  string    // 绝对鼠标HID数据写入路径
	msFuncDir       string    // 大容量存储功能目录
	relHidFuncDir   string    // 相对鼠标HID功能目录
	absHidFuncDir   string    // 绝对鼠标HID功能目录
	isSetupSuccess  bool      // 设备是否初始化成功
}

// New 创建Linux USB Gadget实例（构造函数）
func New(isoPath, gadgetName string, mouseMode MouseMode) *LinuxUSBGadget {
	gadgetDir := filepath.Join("/sys/kernel/config/usb_gadget", gadgetName)
	relHidReportLen := 4 // 相对鼠标报告长度（按钮 + X + Y + 滚轮）
	absHidReportLen := 6 // 绝对鼠标报告长度（报告ID + 按钮 + X(2字节) + Y(2字节)）

	return &LinuxUSBGadget{
		isoPath:         isoPath,
		gadgetName:      gadgetName,
		mouseMode:       mouseMode,
		relHidReportLen: relHidReportLen,
		absHidReportLen: absHidReportLen,
		gadgetDir:       gadgetDir,
		relHidFuncDir:   filepath.Join(gadgetDir, "functions/hid.usb0"),
		absHidFuncDir:   filepath.Join(gadgetDir, "functions/hid.usb1"),
		msFuncDir:       filepath.Join(gadgetDir, "functions/mass_storage.usb0"),
		relHidDataPath:  filepath.Join(gadgetDir, "functions/hid.usb0/data"),
		absHidDataPath:  filepath.Join(gadgetDir, "functions/hid.usb1/data"),
	}
}

// NewRelative 创建相对鼠标模式的USB Gadget实例
func NewRelative(isoPath, gadgetName string) *LinuxUSBGadget {
	return New(isoPath, gadgetName, MouseModeRelative)
}

// NewAbsolute 创建绝对鼠标模式的USB Gadget实例
func NewAbsolute(isoPath, gadgetName string) *LinuxUSBGadget {
	return New(isoPath, gadgetName, MouseModeAbsolute)
}

// ValidateISO 验证ISO镜像有效性（接口实现）
func (g *LinuxUSBGadget) ValidateISO() error {
	if g.isoPath == "" {
		return errors.New("ISO镜像路径未配置")
	}

	// 1. 检查文件存在性
	if _, err := os.Stat(g.isoPath); os.IsNotExist(err) {
		return fmt.Errorf("ISO文件不存在: %s", g.isoPath)
	}

	// 2. 验证ISO格式
	cmd := exec.Command("file", g.isoPath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("执行file命令失败: %w", err)
	}
	if !g.contains(string(output), "ISO 9660") {
		return fmt.Errorf("文件非有效ISO 9660格式: %s", g.isoPath)
	}

	log.Printf("✅ ISO验证通过: %s", g.isoPath)
	return nil
}

// Setup 初始化复合设备（接口实现）
func (g *LinuxUSBGadget) Setup() error {
	// 1. 前置检查
	if os.Geteuid() != 0 {
		return errors.New("必须以root权限运行")
	}
	// 自动识别UDC（未手动指定时）
	if g.udcName == "" {
		udcName, err := g.GetUDCName()
		if err != nil {
			return fmt.Errorf("自动识别UDC失败: %w", err)
		}
		g.udcName = udcName
		log.Printf("🔌 自动识别UDC控制器: %s", g.udcName)
	}

	// 2. 清理旧Gadget
	g.Cleanup()

	// 3. 创建基础目录
	if err := os.MkdirAll(g.gadgetDir, 0755); err != nil {
		return fmt.Errorf("创建Gadget目录失败: %w", err)
	}

	// 4. 基础设备配置（免驱核心：大厂VID/PID）
	configs := map[string]string{
		"idVendor":  "0x046D", // 罗技VID
		"idProduct": "0xC216", // 罗技复合设备PID
		"bcdUSB":    "0x0200", // USB 2.0
		"bcdDevice": "0x0100", // 设备版本
	}
	for path, val := range configs {
		if err := g.writeFile(path, val); err != nil {
			return fmt.Errorf("写入基础配置[%s]失败: %w", path, err)
		}
	}

	// 5. 设备描述配置
	langDir := filepath.Join("strings/0x409")
	if err := os.MkdirAll(filepath.Join(g.gadgetDir, langDir), 0755); err != nil {
		return fmt.Errorf("创建语言目录失败: %w", err)
	}
	langConfigs := map[string]string{
		filepath.Join(langDir, "manufacturer"): "Logitech",
		filepath.Join(langDir, "product"):      "USB Keyboard+Mouse+ISO",
	}
	for path, val := range langConfigs {
		if err := g.writeFile(path, val); err != nil {
			return fmt.Errorf("写入语言配置[%s]失败: %w", path, err)
		}
	}

	// 6. 配置HID键鼠（必选）
	if err := g.setupHID(); err != nil {
		return fmt.Errorf("配置HID失败: %w", err)
	}

	// 7. 配置ISO大容量存储（可选：ISO路径非空时）
	if g.isoPath != "" {
		if err := g.ValidateISO(); err != nil {
			return fmt.Errorf("ISO验证失败: %w", err)
		}
		if err := g.setupMassStorage(); err != nil {
			return fmt.Errorf("配置大容量存储失败: %w", err)
		}
	}

	// 8. 配置复合设备项
	configDir := filepath.Join("configs/c.1")
	if err := os.MkdirAll(filepath.Join(g.gadgetDir, configDir), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	configConfigs := map[string]string{
		filepath.Join(configDir, "bmAttributes"): "0x80", // 总线供电
		filepath.Join(configDir, "MaxPower"):     "500",  // 500mA
	}
	for path, val := range configConfigs {
		if err := g.writeFile(path, val); err != nil {
			return fmt.Errorf("写入配置项[%s]失败: %w", path, err)
		}
	}

	// 9. 绑定功能到配置
	// 根据鼠标模式绑定相应的HID设备
	hidFuncName := ""
	switch g.mouseMode {
	case MouseModeRelative:
		hidFuncName = "hid.usb0"
	case MouseModeAbsolute:
		hidFuncName = "hid.usb1"
	default:
		return errors.New("未知的鼠标操作模式")
	}

	if err := g.bindFunction(configDir, hidFuncName); err != nil {
		return fmt.Errorf("绑定HID失败: %w", err)
	}

	// 绑定大容量存储（ISO场景）
	if g.isoPath != "" {
		if err := g.bindFunction(configDir, "mass_storage.usb0"); err != nil {
			return fmt.Errorf("绑定大容量存储失败: %w", err)
		}
	}

	// 10. 启用设备
	if err := g.writeFile("UDC", g.udcName); err != nil {
		return fmt.Errorf("启用UDC失败: %w（请检查OTG是否启用）", err)
	}

	g.isSetupSuccess = true
	log.Println("✅ 复合设备初始化完成")
	return nil
}

// GetUDCName 自动识别USB控制器（接口实现）
func (g *LinuxUSBGadget) GetUDCName() (string, error) {
	udcDir := "/sys/class/udc"
	dirs, err := os.ReadDir(udcDir)
	if err != nil {
		return "", fmt.Errorf("读取UDC目录失败: %w", err)
	}
	if len(dirs) == 0 {
		return "", errors.New("未找到USB控制器（请启用OTG）")
	}
	return dirs[0].Name(), nil
}

// Cleanup 清理设备资源（接口实现）
func (g *LinuxUSBGadget) Cleanup() error {
	// 停用UDC
	_ = g.writeFile("UDC", "")
	// 删除Gadget目录
	if err := os.RemoveAll(g.gadgetDir); err != nil {
		log.Printf("⚠️ 清理Gadget目录警告: %v", err)
		return err
	}
	g.isSetupSuccess = false
	log.Println("🗑️ 设备资源已清理")
	return nil
}

// GetMouseMode 获取当前鼠标操作模式（接口实现）
func (g *LinuxUSBGadget) GetMouseMode() MouseMode {
	return g.mouseMode
}
