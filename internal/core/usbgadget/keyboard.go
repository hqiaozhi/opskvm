package usbgadget

import (
	"errors"
	"fmt"
	"log"
	"os"
)

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

	// 根据鼠标模式选择相应的HID配置
	var hidReportLen int
	var hidDataPath string
	switch g.mouseMode {
	case MouseModeRelative:
		hidReportLen = g.relHidReportLen
		hidDataPath = g.relHidDataPath
	case MouseModeAbsolute:
		hidReportLen = g.absHidReportLen
		hidDataPath = g.absHidDataPath
	default:
		return errors.New("未知的鼠标操作模式")
	}

	// 构造标准HID键盘报告（8字节）
	report := make([]byte, hidReportLen)
	report[0] = 0x00       // 报告ID
	report[1] = uint8(mod) // 修饰键
	report[2] = 0x00       // 保留
	report[3] = scanCode   // 主按键

	// 发送按下指令
	if err := os.WriteFile(hidDataPath, report, 0644); err != nil {
		return fmt.Errorf("发送键盘按下指令失败: %w", err)
	}
	// 发送释放指令（空报告）
	releaseReport := make([]byte, hidReportLen)
	if err := os.WriteFile(hidDataPath, releaseReport, 0644); err != nil {
		return fmt.Errorf("发送键盘释放指令失败: %w", err)
	}

	log.Printf("📡 发送键盘指令: 修饰键=%d, 按键=%s", mod, key)
	return nil
}
