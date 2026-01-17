package otg

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// addKeyboard 添加HID键盘功能
func (g *Gadget) addKeyboard(start bool, remoteWakeup bool) error {
	return g.addHID("Keyboard", start, remoteWakeup, makeKeyboardHID(nil))
}

// addMouse 添加HID鼠标功能
func (g *Gadget) addMouse(start bool, remoteWakeup bool, absolute bool, horizontalWheel bool) error {
	desc := "Relative Mouse"
	if absolute {
		desc = "Absolute Mouse"
	}
	return g.addHID(desc, start, remoteWakeup, makeMouseHID(absolute, horizontalWheel, nil))
}

// addHID 添加HID功能
func (g *Gadget) addHID(desc string, start bool, remoteWakeup bool, hid Hid) error {
	eps := 1
	funcName := fmt.Sprintf("hid.usb%d", g.hidInstance)
	funcPath, err := g.createFunction(funcName)
	if err != nil {
		return err
	}

	// 写入HID配置
	err = g.write(filepath.Join(funcPath, "no_out_endpoint"), "1", true)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(funcPath, "protocol"), strconv.Itoa(hid.Protocol), false)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(funcPath, "subclass"), strconv.Itoa(hid.Subclass), false)
	if err != nil {
		return err
	}

	err = g.write(filepath.Join(funcPath, "report_length"), strconv.Itoa(hid.ReportLength), false)
	if err != nil {
		return err
	}

	err = g.writeBytes(filepath.Join(funcPath, "report_desc"), hid.ReportDescriptor)
	if err != nil {
		return err
	}

	// 启动功能
	if start {
		err = g.startFunction(funcName, eps)
		if err != nil {
			return err
		}
	}

	// 创建元数据
	err = g.createMeta(funcName, desc, eps)
	if err != nil {
		return err
	}

	g.hidInstance++
	return nil
}

// makeKeyboardHID 创建键盘HID描述符
func makeKeyboardHID(reportID *uint8) Hid {
	reportDescriptor := []byte{
		// Keyboard
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x06, // USAGE (Keyboard)
		0xA1, 0x01, // COLLECTION (Application)
	}

	// 添加Report ID
	if reportID != nil {
		reportDescriptor = append(reportDescriptor, 0x85, *reportID)
	}

	// 添加键盘描述符的其余部分
	reportDescriptor = append(reportDescriptor, []byte{
		// Modifiers
		0x05, 0x07, // USAGE_PAGE (Keyboard)
		0x19, 0xE0, // USAGE_MINIMUM (Keyboard LeftControl)
		0x29, 0xE7, // USAGE_MAXIMUM (Keyboard Right GUI)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x75, 0x01, // REPORT_SIZE (1)
		0x95, 0x08, // REPORT_COUNT (8)
		0x81, 0x02, // INPUT (Data,Var,Abs)

		// Reserved byte
		0x95, 0x01, // REPORT_COUNT (1)
		0x75, 0x08, // REPORT_SIZE (8)
		0x81, 0x01, // INPUT (Const,Array,Abs)

		// LEDs output
		0x95, 0x05, // REPORT_COUNT (5)
		0x75, 0x01, // REPORT_SIZE (1)
		0x05, 0x08, // USAGE_PAGE (LEDs)
		0x19, 0x01, // USAGE_MINIMUM (Num Lock)
		0x29, 0x05, // USAGE_MAXIMUM (Kana)
		0x91, 0x02, // OUTPUT (Data,Var,Abs)

		// Reserved 3 bits in output
		0x95, 0x01, // REPORT_COUNT (1)
		0x75, 0x03, // REPORT_SIZE (3)
		0x91, 0x01, // OUTPUT (Const,Array,Abs)

		// 6 keys
		0x95, 0x06, // REPORT_COUNT (6)
		0x75, 0x08, // REPORT_SIZE (8)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x26, 0xFF, 0x00, // LOGICAL_MAXIMUM (0xFF)
		0x05, 0x07, // USAGE_PAGE (Keyboard)
		0x19, 0x00, // USAGE_MINIMUM (Reserved)
		0x2A, 0xFF, 0x00, // USAGE_MAXIMUM (0xFF)
		0x81, 0x00, // INPUT (Data,Array,Abs)

		0xC0, // END_COLLECTION
	}...)

	return Hid{
		Protocol:         1, // Keyboard protocol
		Subclass:         1, // Boot interface subclass
		ReportLength:     8,
		ReportDescriptor: reportDescriptor,
	}
}

// makeMouseHID 创建鼠标HID描述符
func makeMouseHID(absolute bool, horizontalWheel bool, reportID *uint8) Hid {
	// 根据鼠标类型调用不同的创建函数
	if absolute {
		return makeAbsoluteHID(horizontalWheel, reportID)
	} else {
		return makeRelativeHID(horizontalWheel, reportID)
	}
}

// makeAbsoluteHID 创建绝对鼠标HID描述符
func makeAbsoluteHID(horizontalWheel bool, reportID *uint8) Hid {
	reportDescriptor := []byte{
		// Mouse
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x02, // USAGE (Mouse)
		0xA1, 0x01, // COLLECTION (Application)
	}

	// 添加Report ID
	if reportID != nil {
		reportDescriptor = append(reportDescriptor, 0x85, *reportID)
	}

	// 添加鼠标描述符的其余部分
	reportDescriptor = append(reportDescriptor, []byte{
		// Pointer and Physical are required by Apple Recovery
		0x09, 0x01, // USAGE (Pointer)
		0xA1, 0x00, // COLLECTION (Physical)

		// 8 Buttons
		0x05, 0x09, // USAGE_PAGE (Button)
		0x19, 0x01, // USAGE_MINIMUM (Button 1)
		0x29, 0x08, // USAGE_MAXIMUM (Button 8)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x95, 0x08, // REPORT_COUNT (8)
		0x75, 0x01, // REPORT_SIZE (1)
		0x81, 0x02, // INPUT (Data,Var,Abs)

		// X, Y
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x30, // USAGE (X)
		0x09, 0x31, // USAGE (Y)
		0x16, 0x00, 0x00, // LOGICAL_MINIMUM (0)
		0x26, 0xFF, 0x7F, // LOGICAL_MAXIMUM (32767)
		0x75, 0x10, // REPORT_SIZE (16)
		0x95, 0x02, // REPORT_COUNT (2)
		0x81, 0x02, // INPUT (Data,Var,Abs)

		// Wheel
		0x09, 0x38, // USAGE (Wheel)
		0x15, 0x81, // LOGICAL_MINIMUM (-127)
		0x25, 0x7F, // LOGICAL_MAXIMUM (127)
		0x75, 0x08, // REPORT_SIZE (8)
		0x95, 0x01, // REPORT_COUNT (1)
		0x81, 0x06, // INPUT (Data,Var,Rel)
	}...)

	// 水平滚轮
	if horizontalWheel {
		reportDescriptor = append(reportDescriptor, []byte{
			0x05, 0x0C, // USAGE_PAGE (Consumer Devices)
			0x0A, 0x38, 0x02, // USAGE (AC Pan)
			0x15, 0x81, // LOGICAL_MINIMUM (-127)
			0x25, 0x7F, // LOGICAL_MAXIMUM (127)
			0x75, 0x08, // REPORT_SIZE (8)
			0x95, 0x01, // REPORT_COUNT (1)
			0x81, 0x06, // INPUT (Data,Var,Rel)
		}...)
	}

	// 结束集合
	reportDescriptor = append(reportDescriptor, []byte{
		0xC0, // END_COLLECTION (Physical)
		0xC0, // END_COLLECTION
	}...)

	// 设置报告长度
	reportLength := 6
	if horizontalWheel {
		reportLength = 7
	}

	return Hid{
		Protocol:         0, // None protocol
		Subclass:         0, // No subclass
		ReportLength:     reportLength,
		ReportDescriptor: reportDescriptor,
	}
}

// makeRelativeHID 创建相对鼠标HID描述符
func makeRelativeHID(horizontalWheel bool, reportID *uint8) Hid {
	reportDescriptor := []byte{
		// Mouse
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x02, // USAGE (Mouse)
		0xA1, 0x01, // COLLECTION (Application)
	}

	// 添加Report ID
	if reportID != nil {
		reportDescriptor = append(reportDescriptor, 0x85, *reportID)
	}

	// 添加鼠标描述符的其余部分
	reportDescriptor = append(reportDescriptor, []byte{
		// Pointer and Physical are required by Apple Recovery
		0x09, 0x01, // USAGE (Pointer)
		0xA1, 0x00, // COLLECTION (Physical)

		// 8 Buttons
		0x05, 0x09, // USAGE_PAGE (Button)
		0x19, 0x01, // USAGE_MINIMUM (Button 1)
		0x29, 0x08, // USAGE_MAXIMUM (Button 8)
		0x15, 0x00, // LOGICAL_MINIMUM (0)
		0x25, 0x01, // LOGICAL_MAXIMUM (1)
		0x95, 0x08, // REPORT_COUNT (8)
		0x75, 0x01, // REPORT_SIZE (1)
		0x81, 0x02, // INPUT (Data,Var,Abs)

		// X, Y
		0x05, 0x01, // USAGE_PAGE (Generic Desktop)
		0x09, 0x30, // USAGE (X)
		0x09, 0x31, // USAGE (Y)

		// Wheel
		0x09, 0x38, // USAGE (Wheel)
		0x15, 0x81, // LOGICAL_MINIMUM (-127)
		0x25, 0x7F, // LOGICAL_MAXIMUM (127)
		0x75, 0x08, // REPORT_SIZE (8)
		0x95, 0x03, // REPORT_COUNT (3)
		0x81, 0x06, // INPUT (Data,Var,Rel)
	}...)

	// 水平滚轮
	if horizontalWheel {
		reportDescriptor = append(reportDescriptor, []byte{
			0x05, 0x0C, // USAGE_PAGE (Consumer Devices)
			0x0A, 0x38, 0x02, // USAGE (AC Pan)
			0x15, 0x81, // LOGICAL_MINIMUM (-127)
			0x25, 0x7F, // LOGICAL_MAXIMUM (127)
			0x75, 0x08, // REPORT_SIZE (8)
			0x95, 0x01, // REPORT_COUNT (1)
			0x81, 0x06, // INPUT (Data,Var,Rel)
		}...)
	}

	// 结束集合
	reportDescriptor = append(reportDescriptor, []byte{
		0xC0, // END_COLLECTION (Physical)
		0xC0, // END_COLLECTION
	}...)

	// 设置报告长度
	reportLength := 4
	if horizontalWheel {
		reportLength = 5
	}

	return Hid{
		Protocol:         2, // Mouse protocol
		Subclass:         1, // Boot interface subclass
		ReportLength:     reportLength,
		ReportDescriptor: reportDescriptor,
	}
}
