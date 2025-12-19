package mk

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// -------------------------- 核心接口定义 --------------------------
// USBCompositeDevice 复合USB设备接口（ISO挂载 + HID键鼠控制）
type USBCompositeDevice interface {
	// 验证ISO镜像有效性（仅ISO场景需调用）
	ValidateISO() error
	// 初始化设备（配置Gadget + 绑定ISO + 绑定HID）
	Setup() error
	// 发送鼠标控制指令
	SendMouse(btn MouseButton, x, y int8) error
	// 发送键盘控制指令
	SendKeyboard(mod KeyboardModifier, key KeyboardKey) error
	// 自动识别USB控制器名称
	GetUDCName() (string, error)
	// 清理设备资源（退出时调用）
	Cleanup() error
}

// -------------------------- 枚举与类型定义 --------------------------
// MouseButton 鼠标按键枚举
type MouseButton uint8

const (
	MouseButtonNone  MouseButton = 0x00 // 无按键
	MouseButtonLeft  MouseButton = 0x01 // 左键
	MouseButtonRight MouseButton = 0x02 // 右键
	MouseButtonMid   MouseButton = 0x04 // 中键
)

// KeyboardModifier 键盘修饰键枚举
type KeyboardModifier uint8

const (
	KeyboardModifierNone  KeyboardModifier = 0x00 // 无修饰键
	KeyboardModifierCtrl  KeyboardModifier = 0x01 // Ctrl
	KeyboardModifierShift KeyboardModifier = 0x02 // Shift
	KeyboardModifierAlt   KeyboardModifier = 0x04 // Alt
	KeyboardModifierGUI   KeyboardModifier = 0x08 // Win/Cmd
)

// KeyboardKey 键盘按键枚举（标准HID扫描码映射）
type KeyboardKey string

const (
	KeyA     KeyboardKey = "a"
	KeyB     KeyboardKey = "b"
	KeyEnter KeyboardKey = "enter"
	KeySpace KeyboardKey = "space"
	KeyUp    KeyboardKey = "up"
	KeyDown  KeyboardKey = "down"
	KeyLeft  KeyboardKey = "left"
	KeyRight KeyboardKey = "right"
	KeyNone  KeyboardKey = ""
	// 可扩展更多按键...
)

// -------------------------- 结构体实现 --------------------------
// LinuxUSBGadget Linux USB Gadget实现（核心结构体）
type LinuxUSBGadget struct {
	isoPath        string // ISO镜像路径（空则仅启用HID）
	gadgetName     string // Gadget名称
	udcName        string // USB控制器名称（自动识别/手动指定）
	hidReportLen   int    // HID报告长度
	gadgetDir      string // Gadget配置目录
	hidDataPath    string // HID数据写入路径（发送键鼠指令）
	msFuncDir      string // 大容量存储功能目录
	hidFuncDir     string // HID功能目录
	isSetupSuccess bool   // 设备是否初始化成功
}

// NewLinuxUSBGadget 创建Linux USB Gadget实例（构造函数）
func NewLinuxUSBGadget(isoPath, gadgetName string, hidReportLen int) *LinuxUSBGadget {
	gadgetDir := filepath.Join("/sys/kernel/config/usb_gadget", gadgetName)
	return &LinuxUSBGadget{
		isoPath:      isoPath,
		gadgetName:   gadgetName,
		hidReportLen: hidReportLen,
		gadgetDir:    gadgetDir,
		hidFuncDir:   filepath.Join(gadgetDir, "functions/hid.usb0"),
		msFuncDir:    filepath.Join(gadgetDir, "functions/mass_storage.usb0"),
		hidDataPath:  filepath.Join(gadgetDir, "functions/hid.usb0/data"),
	}
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
	// 绑定HID
	if err := g.bindFunction(configDir, "hid.usb0"); err != nil {
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

// SendMouse 发送鼠标指令（接口实现）
func (g *LinuxUSBGadget) SendMouse(btn MouseButton, x, y int8) error {
	if !g.isSetupSuccess {
		return errors.New("设备未初始化成功，请先调用Setup()")
	}

	// 构造标准HID鼠标报告（8字节）
	report := make([]byte, g.hidReportLen)
	report[0] = 0x00       // 报告ID
	report[1] = uint8(btn) // 鼠标按键
	report[2] = 0x00       // 保留
	report[3] = uint8(x)   // X偏移
	report[4] = uint8(y)   // Y偏移

	if err := os.WriteFile(g.hidDataPath, report, 0644); err != nil {
		return fmt.Errorf("发送鼠标指令失败: %w", err)
	}
	log.Printf("📡 发送鼠标指令: 按键=%d, X=%d, Y=%d", btn, x, y)
	return nil
}

// SendKeyboard 发送键盘指令（接口实现）
func (g *LinuxUSBGadget) SendKeyboard(mod KeyboardModifier, key KeyboardKey) error {
	if !g.isSetupSuccess {
		return errors.New("设备未初始化成功，请先调用Setup()")
	}

	// 按键扫描码映射（标准HID）
	keyMap := map[KeyboardKey]uint8{
		KeyA:     0x04,
		KeyB:     0x05,
		KeyEnter: 0x28,
		KeySpace: 0x2C,
		KeyUp:    0x52,
		KeyDown:  0x51,
		KeyLeft:  0x50,
		KeyRight: 0x4F,
		KeyNone:  0x00,
	}
	scanCode, ok := keyMap[key]
	if !ok {
		return fmt.Errorf("不支持的按键: %s", key)
	}

	// 构造标准HID键盘报告（8字节）
	report := make([]byte, g.hidReportLen)
	report[0] = 0x00       // 报告ID
	report[1] = uint8(mod) // 修饰键
	report[2] = 0x00       // 保留
	report[3] = scanCode   // 主按键

	// 发送按下指令
	if err := os.WriteFile(g.hidDataPath, report, 0644); err != nil {
		return fmt.Errorf("发送键盘按下指令失败: %w", err)
	}
	// 发送释放指令（空报告）
	releaseReport := make([]byte, g.hidReportLen)
	if err := os.WriteFile(g.hidDataPath, releaseReport, 0644); err != nil {
		return fmt.Errorf("发送键盘释放指令失败: %w", err)
	}

	log.Printf("📡 发送键盘指令: 修饰键=%d, 按键=%s", mod, key)
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

// -------------------------- 私有辅助方法 --------------------------
// setupHID 配置HID键鼠功能
func (g *LinuxUSBGadget) setupHID() error {
	if err := os.MkdirAll(g.hidFuncDir, 0755); err != nil {
		return fmt.Errorf("创建HID目录失败: %w", err)
	}

	// HID基础配置
	hidConfigs := map[string]string{
		"protocol":      "1", // 键鼠组合设备
		"subclass":      "1", // 启动设备
		"report_length": fmt.Sprintf("%d", g.hidReportLen),
	}
	for path, val := range hidConfigs {
		fullPath := filepath.Join(g.hidFuncDir, path)
		if err := os.WriteFile(fullPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("写入HID配置[%s]失败: %w", path, err)
		}
	}

	// 标准HID键鼠报告描述符
	desc := []byte{
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x06, // USAGE (Keyboard)
		0xA1, 0x01, // COLLECTION (Application)
		0x05, 0x07, // USAGE_PAGE (Keyboard/Keypad)
		0x19, 0xE0, // USAGE_MINIMUM (Keyboard LeftControl)
		0x29, 0xE7, // USAGE_MAXIMUM (Keyboard Right GUI)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x75, 0x01, // REPORT_SIZE (1)
		0x95, 0x08, // REPORT_COUNT (8)
		0x81, 0x02, // INPUT (Data,Var,Abs)
		0x95, 0x01, // REPORT_COUNT (1)
		0x75, 0x08, // REPORT_SIZE (8)
		0x81, 0x03, // INPUT (Cnst,Var,Abs)
		0x95, 0x06, // REPORT_COUNT (6)
		0x75, 0x08, // REPORT_SIZE (8)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x65, // LOGICAL_MAXIMUM (101)
		0x05, 0x07, // USAGE_PAGE (Keyboard/Keypad)
		0x19, 0x00, // USAGE_MINIMUM (Reserved)
		0x29, 0x65, // USAGE_MAXIMUM (Keyboard Application)
		0x81, 0x00, // INPUT (Data,Array)
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x02, // USAGE (Mouse)
		0xA1, 0x01, // COLLECTION (Application)
		0x09, 0x01, // USAGE (Pointer)
		0xA1, 0x00, // COLLECTION (Physical)
		0x05, 0x09, // USAGE_PAGE (Button)
		0x19, 0x01, // USAGE_MINIMUM (Button 1)
		0x29, 0x03, // USAGE_MAXIMUM (Button 3)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x75, 0x01, // REPORT_SIZE (1)
		0x95, 0x03, // REPORT_COUNT (3)
		0x81, 0x02, // INPUT (Data,Var,Abs)
		0x75, 0x05, // REPORT_SIZE (5)
		0x95, 0x01, // REPORT_COUNT (1)
		0x81, 0x03, // INPUT (Cnst,Var,Abs)
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x30, // USAGE (X)
		0x09, 0x31, // USAGE (Y)
		0x15, 0x81, // LOGICAL_MINIMUM (-127)
		0x25, 0x7F, // LOGICAL_MAXIMUM (127)
		0x75, 0x08, // REPORT_SIZE (8)
		0x95, 0x02, // REPORT_COUNT (2)
		0x81, 0x06, // INPUT (Data,Var,Rel)
		0xC0, // END_COLLECTION
		0xC0, // END_COLLECTION
		0xC0, // END_COLLECTION
	}
	descPath := filepath.Join(g.hidFuncDir, "report_desc")
	if err := os.WriteFile(descPath, desc, 0644); err != nil {
		return fmt.Errorf("写入HID描述符失败: %w", err)
	}
	return nil
}

// setupMassStorage 配置ISO大容量存储功能
func (g *LinuxUSBGadget) setupMassStorage() error {
	if err := os.MkdirAll(g.msFuncDir, 0755); err != nil {
		return fmt.Errorf("创建大容量存储目录失败: %w", err)
	}

	// ISO挂载核心配置
	msConfigs := map[string]string{
		"lun.0/file":           g.isoPath, // 绑定ISO
		"lun.0/cdrom":          "1",       // 识别为光盘（0=U盘）
		"lun.0/ro":             "1",       // 只读（ISO必须）
		"lun.0/removable":      "1",       // 可移除设备
		"lun.0/inquiry_string": "Go ISO",  // 设备标识
	}
	for path, val := range msConfigs {
		fullPath := filepath.Join(g.msFuncDir, path)
		if err := os.WriteFile(fullPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("写入大容量存储配置[%s]失败: %w", path, err)
		}
	}
	return nil
}

// writeFile 写入文件到Gadget目录
func (g *LinuxUSBGadget) writeFile(path, content string) error {
	fullPath := filepath.Join(g.gadgetDir, path)
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// contains 字符串包含判断
func (g *LinuxUSBGadget) contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// bindFunction 绑定功能到配置项
func (g *LinuxUSBGadget) bindFunction(configDir, funcName string) error {
	funcPath := filepath.Join(g.gadgetDir, "functions", funcName)
	targetPath := filepath.Join(g.gadgetDir, configDir, funcName)
	if err := os.Symlink(funcPath, targetPath); err != nil && !os.IsExist(err) {
		return fmt.Errorf("创建符号链接失败: %w", err)
	}
	return nil
}

// -------------------------- 示例使用 --------------------------
func main() {
	// 1. 创建设备实例（灵活配置：ISO路径/设备名称/报告长度）
	// 场景1：仅键鼠控制（ISO路径为空）
	// gadget := NewLinuxUSBGadget("", "hid-only-gadget", 8)
	// 场景2：键鼠+ISO挂载（指定ISO路径）
	gadget := NewLinuxUSBGadget("/tmp/test.iso", "hid-iso-gadget", 8)

	// 2. 延迟清理资源
	defer func() {
		if err := gadget.Cleanup(); err != nil {
			log.Fatalf("清理资源失败: %v", err)
		}
	}()

	// 3. 初始化设备
	if err := gadget.Setup(); err != nil {
		log.Fatalf("设备初始化失败: %v", err)
	}

	// 4. 示例：键鼠控制
	log.Println("🚀 设备已启动，开始测试键鼠控制...")
	// 鼠标右移10，下移5
	if err := gadget.SendMouse(MouseButtonNone, 10, 5); err != nil {
		log.Printf("鼠标控制失败: %v", err)
	}
	// 鼠标左键点击
	if err := gadget.SendMouse(MouseButtonLeft, 0, 0); err != nil {
		log.Printf("鼠标点击失败: %v", err)
	}
	// 按下Ctrl+A
	if err := gadget.SendKeyboard(KeyboardModifierCtrl, KeyA); err != nil {
		log.Printf("键盘控制失败: %v", err)
	}

	// 5. 持续运行（阻塞）
	log.Println("✅ 键鼠+ISO设备运行中，按Ctrl+C退出...")
	select {}
}
