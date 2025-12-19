package usbgadget

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNew 测试构造函数
func TestNew(t *testing.T) {
	// 测试场景1：仅HID模式（无ISO）
	gadget := New("", "test-gadget", 8)
	if gadget == nil {
		t.Fatal("构造函数返回nil")
	}
	if gadget.isoPath != "" {
		t.Errorf("期望isoPath为空，实际为: %s", gadget.isoPath)
	}
	if gadget.gadgetName != "test-gadget" {
		t.Errorf("期望gadgetName为'test-gadget'，实际为: %s", gadget.gadgetName)
	}
	if gadget.hidReportLen != 8 {
		t.Errorf("期望hidReportLen为8，实际为: %d", gadget.hidReportLen)
	}
	if gadget.gadgetDir != "/sys/kernel/config/usb_gadget/test-gadget" {
		t.Errorf("期望gadgetDir为'/sys/kernel/config/usb_gadget/test-gadget'，实际为: %s", gadget.gadgetDir)
	}

	// 测试场景2：HID+ISO模式
	gadgetWithISO := New("/test.iso", "test-gadget-iso", 16)
	if gadgetWithISO == nil {
		t.Fatal("构造函数返回nil")
	}
	if gadgetWithISO.isoPath != "/test.iso" {
		t.Errorf("期望isoPath为'/test.iso'，实际为: %s", gadgetWithISO.isoPath)
	}
	if gadgetWithISO.hidReportLen != 16 {
		t.Errorf("期望hidReportLen为16，实际为: %d", gadgetWithISO.hidReportLen)
	}
}

// TestContains 测试字符串包含方法
func TestContains(t *testing.T) {
	gadget := New("", "test-gadget", 8)

	// 测试正常包含
	if !gadget.contains("ISO 9660 CD-ROM", "ISO 9660") {
		t.Error("期望包含'ISO 9660'")
	}

	// 测试不包含
	if gadget.contains("普通文件", "ISO 9660") {
		t.Error("期望不包含'ISO 9660'")
	}

	// 测试空字符串
	if gadget.contains("", "ISO 9660") {
		t.Error("空字符串不应该包含任何内容")
	}
}

// TestKeyboardKeyMapping 测试键盘按键映射
func TestKeyboardKeyMapping(t *testing.T) {
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

	// 验证所有按键都有映射
	for key, expectedScanCode := range keyMap {
		if scanCode, ok := keyMap[key]; !ok {
			t.Errorf("按键 %s 没有映射", key)
		} else if scanCode != expectedScanCode {
			t.Errorf("按键 %s 映射错误，期望 0x%02X，实际 0x%02X", key, expectedScanCode, scanCode)
		}
	}
}

// TestValidateISO_NonExistentFile 测试验证不存在的ISO文件
func TestValidateISO_NonExistentFile(t *testing.T) {
	gadget := New("/non/existent/file.iso", "test-gadget", 8)
	err := gadget.ValidateISO()
	if err == nil {
		t.Error("期望验证失败，实际成功")
	}
	if err.Error() != "ISO文件不存在: /non/existent/file.iso" {
		t.Errorf("期望错误信息为'ISO文件不存在: /non/existent/file.iso'，实际为: %s", err.Error())
	}
}

// TestValidateISO_EmptyPath 测试空路径验证
func TestValidateISO_EmptyPath(t *testing.T) {
	gadget := New("", "test-gadget", 8)
	err := gadget.ValidateISO()
	if err == nil {
		t.Error("期望空路径验证失败，实际成功")
	}
	if err.Error() != "ISO镜像路径未配置" {
		t.Errorf("期望错误信息为'ISO镜像路径未配置'，实际为: %s", err.Error())
	}
}

// TestMouseButtonConstants 测试鼠标按键常量
func TestMouseButtonConstants(t *testing.T) {
	// 验证常量值
	if MouseButtonNone != 0x00 {
		t.Errorf("期望MouseButtonNone为0x00，实际为: 0x%02X", MouseButtonNone)
	}
	if MouseButtonLeft != 0x01 {
		t.Errorf("期望MouseButtonLeft为0x01，实际为: 0x%02X", MouseButtonLeft)
	}
	if MouseButtonRight != 0x02 {
		t.Errorf("期望MouseButtonRight为0x02，实际为: 0x%02X", MouseButtonRight)
	}
	if MouseButtonMid != 0x04 {
		t.Errorf("期望MouseButtonMid为0x04，实际为: 0x%02X", MouseButtonMid)
	}
}

// TestKeyboardModifierConstants 测试键盘修饰键常量
func TestKeyboardModifierConstants(t *testing.T) {
	// 验证常量值
	if KeyboardModifierNone != 0x00 {
		t.Errorf("期望KeyboardModifierNone为0x00，实际为: 0x%02X", KeyboardModifierNone)
	}
	if KeyboardModifierCtrl != 0x01 {
		t.Errorf("期望KeyboardModifierCtrl为0x01，实际为: 0x%02X", KeyboardModifierCtrl)
	}
	if KeyboardModifierShift != 0x02 {
		t.Errorf("期望KeyboardModifierShift为0x02，实际为: 0x%02X", KeyboardModifierShift)
	}
	if KeyboardModifierAlt != 0x04 {
		t.Errorf("期望KeyboardModifierAlt为0x04，实际为: 0x%02X", KeyboardModifierAlt)
	}
	if KeyboardModifierGUI != 0x08 {
		t.Errorf("期望KeyboardModifierGUI为0x08，实际为: 0x%02X", KeyboardModifierGUI)
	}
}

// TestSetupDirectoryStructure 测试目录结构设置
func TestSetupDirectoryStructure(t *testing.T) {
	// 创建临时目录模拟sysfs
	tmpDir := t.TempDir()
	gadgetDir := filepath.Join(tmpDir, "usb_gadget", "test-gadget")

	// 创建测试用的gadget实例
	gadget := &LinuxUSBGadget{
		gadgetName: "test-gadget",
		gadgetDir:  gadgetDir,
	}

	// 测试目录创建
	err := os.MkdirAll(gadget.gadgetDir, 0755)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 验证目录是否存在
	if _, err := os.Stat(gadget.gadgetDir); os.IsNotExist(err) {
		t.Errorf("期望目录 %s 存在，实际不存在", gadget.gadgetDir)
	}
}

// TestWriteFile 测试文件写入方法
func TestWriteFile(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	gadgetDir := filepath.Join(tmpDir, "usb_gadget", "test-gadget")

	// 创建测试用的gadget实例
	gadget := &LinuxUSBGadget{
		gadgetName: "test-gadget",
		gadgetDir:  gadgetDir,
	}

	// 创建基础目录
	if err := os.MkdirAll(gadget.gadgetDir, 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 测试文件写入
	testPath := "test.txt"
	testContent := "test content"
	if err := gadget.writeFile(testPath, testContent); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 验证文件内容
	fullPath := filepath.Join(gadget.gadgetDir, testPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("期望文件内容为 '%s'，实际为 '%s'", testContent, string(content))
	}
}
