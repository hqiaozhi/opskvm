package usbgadget

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// SendMouseRelative 发送相对鼠标指令（接口实现）
func (g *LinuxUSBGadget) SendMouseRelative(btn MouseButton, x, y int8) error {
	if !g.isSetupSuccess {
		return errors.New("设备未初始化成功，请先调用Setup()")
	}

	// 构造相对鼠标HID报告（4字节）
	report := make([]byte, g.relHidReportLen)
	report[0] = uint8(btn) // 鼠标按键
	report[1] = uint8(x)   // X偏移
	report[2] = uint8(y)   // Y偏移
	report[3] = 0x00       // 滚轮

	if err := os.WriteFile(g.relHidDataPath, report, 0644); err != nil {
		return fmt.Errorf("发送相对鼠标指令失败: %w", err)
	}
	log.Printf("📡 发送相对鼠标指令: 按键=%d, X=%d, Y=%d", btn, x, y)
	return nil
}

// SendMouseAbsolute 发送绝对鼠标指令（接口实现）
func (g *LinuxUSBGadget) SendMouseAbsolute(btn MouseButton, x, y int) error {
	if !g.isSetupSuccess {
		return errors.New("设备未初始化成功，请先调用Setup()")
	}

	// 限制坐标范围在0-32767之间
	if x < 0 {
		x = 0
	} else if x > 32767 {
		x = 32767
	}
	if y < 0 {
		y = 0
	} else if y > 32767 {
		y = 32767
	}

	// 构造绝对鼠标HID报告（6字节）
	report := make([]byte, g.absHidReportLen)
	report[0] = 0x01          // 报告ID 1（绝对鼠标移动）
	report[1] = uint8(btn)    // 鼠标按键
	report[2] = uint8(x)      // X坐标低字节
	report[3] = uint8(x >> 8) // X坐标高字节
	report[4] = uint8(y)      // Y坐标低字节
	report[5] = uint8(y >> 8) // Y坐标高字节

	if err := os.WriteFile(g.absHidDataPath, report, 0644); err != nil {
		return fmt.Errorf("发送绝对鼠标指令失败: %w", err)
	}
	log.Printf("📡 发送绝对鼠标指令: 按键=%d, X=%d, Y=%d", btn, x, y)
	return nil
}
