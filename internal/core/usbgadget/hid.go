package usbgadget

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// setupHID 配置HID键鼠功能
func (g *LinuxUSBGadget) setupHID() error {
	// 根据鼠标模式配置相应的HID设备
	switch g.mouseMode {
	case MouseModeRelative:
		return g.setupRelativeMouseHID()
	case MouseModeAbsolute:
		return g.setupAbsoluteMouseHID()
	default:
		return errors.New("未知的鼠标操作模式")
	}
}

// setupRelativeMouseHID 配置相对鼠标HID功能
func (g *LinuxUSBGadget) setupRelativeMouseHID() error {
	if err := os.MkdirAll(g.relHidFuncDir, 0755); err != nil {
		return fmt.Errorf("创建相对鼠标HID目录失败: %w", err)
	}

	// 相对鼠标HID配置
	hidConfigs := map[string]string{
		"protocol":        "2", // 鼠标设备
		"subclass":        "1", // 启动设备
		"report_length":   fmt.Sprintf("%d", g.relHidReportLen),
		"no_out_endpoint": "1",
	}
	for path, val := range hidConfigs {
		fullPath := filepath.Join(g.relHidFuncDir, path)
		if err := os.WriteFile(fullPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("写入相对鼠标HID配置[%s]失败: %w", path, err)
		}
	}

	// 相对鼠标HID报告描述符
	relativeDesc := []byte{
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x02, // USAGE (Mouse)
		0xa1, 0x01, // COLLECTION (Application)
		0x09, 0x01, // USAGE (Pointer)
		0xa1, 0x00, // COLLECTION (Physical)
		// 8 Buttons
		0x05, 0x09, // USAGE_PAGE (Button)
		0x19, 0x01, // USAGE_MINIMUM (Button 1)
		0x29, 0x08, // USAGE_MAXIMUM (Button 8)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x95, 0x08, // REPORT_COUNT (8)
		0x75, 0x01, // REPORT_SIZE (1)
		0x81, 0x02, // INPUT (Data,Var,Abs)
		// X, Y, Wheel
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x30, // USAGE (X)
		0x09, 0x31, // USAGE (Y)
		0x09, 0x38, // USAGE (Wheel)
		0x15, 0x81, // LOGICAL_MINIMUM (-127)
		0x25, 0x7f, // LOGICAL_MAXIMUM (127)
		0x75, 0x08, // REPORT_SIZE (8)
		0x95, 0x03, // REPORT_COUNT (3)
		0x81, 0x06, // INPUT (Data,Var,Rel)
		0xc0, // END COLLECTION (Physical)
		0xc0, // END COLLECTION
	}
	descPath := filepath.Join(g.relHidFuncDir, "report_desc")
	if err := os.WriteFile(descPath, relativeDesc, 0644); err != nil {
		return fmt.Errorf("写入相对鼠标HID描述符失败: %w", err)
	}
	return nil
}

// setupAbsoluteMouseHID 配置绝对鼠标HID功能
func (g *LinuxUSBGadget) setupAbsoluteMouseHID() error {
	if err := os.MkdirAll(g.absHidFuncDir, 0755); err != nil {
		return fmt.Errorf("创建绝对鼠标HID目录失败: %w", err)
	}

	// 绝对鼠标HID配置
	hidConfigs := map[string]string{
		"protocol":        "2", // 鼠标设备
		"subclass":        "0", // 非启动设备
		"report_length":   fmt.Sprintf("%d", g.absHidReportLen),
		"no_out_endpoint": "1",
	}
	for path, val := range hidConfigs {
		fullPath := filepath.Join(g.absHidFuncDir, path)
		if err := os.WriteFile(fullPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("写入绝对鼠标HID配置[%s]失败: %w", path, err)
		}
	}

	// 绝对鼠标HID报告描述符
	absoluteDesc := []byte{
		0x05, 0x01, // Usage Page (Generic Desktop Ctrls)
		0x09, 0x02, // Usage (Mouse)
		0xA1, 0x01, // Collection (Application)
		// Report ID 1: Absolute Mouse Movement
		0x85, 0x01, //     Report ID (1)
		0x09, 0x01, //     Usage (Pointer)
		0xA1, 0x00, //     Collection (Physical)
		0x05, 0x09, //         Usage Page (Button)
		0x19, 0x01, //         Usage Minimum (0x01)
		0x29, 0x03, //         Usage Maximum (0x03)
		0x15, 0x00, //         Logical Minimum (0)
		0x25, 0x01, //         Logical Maximum (1)
		0x75, 0x01, //         Report Size (1)
		0x95, 0x03, //         Report Count (3)
		0x81, 0x02, //         Input (Data, Var, Abs)
		0x95, 0x01, //         Report Count (1)
		0x75, 0x05, //         Report Size (5)
		0x81, 0x03, //         Input (Cnst, Var, Abs)
		0x05, 0x01, //         Usage Page (Generic Desktop Ctrls)
		0x09, 0x30, //         Usage (X)
		0x09, 0x31, //         Usage (Y)
		0x16, 0x00, 0x00, //         Logical Minimum (0)
		0x26, 0xFF, 0x7F, //         Logical Maximum (32767)
		0x36, 0x00, 0x00, //         Physical Minimum (0)
		0x46, 0xFF, 0x7F, //         Physical Maximum (32767)
		0x75, 0x10, //         Report Size (16)
		0x95, 0x02, //         Report Count (2)
		0x81, 0x02, //         Input (Data, Var, Abs)
		0xC0, //     End Collection
		// Report ID 2: Relative Wheel Movement
		0x85, 0x02, //     Report ID (2)
		0x09, 0x38, //     Usage (Wheel)
		0x15, 0x81, //     Logical Minimum (-127)
		0x25, 0x7F, //     Logical Maximum (127)
		0x35, 0x00, //     Physical Minimum (0) = Reset Physical Minimum
		0x45, 0x00, //     Physical Maximum (0) = Reset Physical Maximum
		0x75, 0x08, //     Report Size (8)
		0x95, 0x01, //     Report Count (1)
		0x81, 0x06, //     Input (Data, Var, Rel)
		0xC0, // End Collection
	}
	descPath := filepath.Join(g.absHidFuncDir, "report_desc")
	if err := os.WriteFile(descPath, absoluteDesc, 0644); err != nil {
		return fmt.Errorf("写入绝对鼠标HID描述符失败: %w", err)
	}
	return nil
}
