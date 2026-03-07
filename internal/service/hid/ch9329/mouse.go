package ch9329

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CH9329 鼠标按键
const (
	MouseLeft    = 0x01
	MouseRight   = 0x02
	MouseMiddle  = 0x04
	MouseBack    = 0x08 // Back/Up
	MouseForward = 0x10 // Forward/Down
)

// MouseRange 定义鼠标坐标范围，与Python版本完全一致
const (
	MouseRangeMin = 0
	MouseRangeMax = 0xFFFF
)

// MouseDelta 定义鼠标增量范围，与Python版本完全一致
const (
	MouseDeltaMin = -127
	MouseDeltaMax = 127
)

// remap 将值从一个范围映射到另一个范围，与Python版本的MouseRange.remap一致
func remap(value, fromMin, fromMax, toMin, toMax int) int {
	// 计算输入值在原范围内的比例
	ratio := float64(value-fromMin) / float64(fromMax-fromMin)
	// 映射到目标范围
	return int(float64(toMin) + ratio*float64(toMax-toMin))
}

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *CH9329Device) SetAbsoluteMouse(absolute bool) error {
	g.Log().Infof(context.Background(), "Setting absolute mouse mode to: %v", absolute)
	d.absolute = absolute
	return nil
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *CH9329Device) IsAbsoluteMouse() bool {
	return d.absolute
}

// SendMouseReport 发送鼠标HID报告，与Python版本完全一致
func (d *CH9329Device) SendMouseReport(buttons byte, dx, dy int, wheel int8) error {
	if d.absolute || buttons > 0 || wheel != 0 || dx != 0 || dy != 0 {
		g.Log().Debugf(context.Background(), "SendMouseReport called - absolute: %v, buttons: 0x%02x, dx: %d, dy: %d, wheel: %d", d.absolute, buttons, dx, dy, wheel)
	}
	var cmd []byte

	if d.absolute {
		// 绝对鼠标模式，与Python版本完全一致的坐标处理逻辑
		// 前端发送的是0-65535范围的绝对坐标
		var absDx, absDy int

		// 直接使用前端发送的坐标，不再进行额外转换
		// 确保坐标在0-65535范围内
		if dx < 0 {
			dx = 0
		}
		if dx > 65535 {
			dx = 65535
		}
		if dy < 0 {
			dy = 0
		}
		if dy > 65535 {
			dy = 65535
		}

		absDx = dx
		absDy = dy

		// 直接使用0-65535范围的坐标，不需要remap
		// 直接除以8并向上取整，与Python的math.ceil一致
		// 65535 / 8 = 8191.875，向上取整为8192，与CH9329的要求一致
		fixedX := (absDx + 7) / 8
		fixedY := (absDy + 7) / 8
		g.Log().Debugf(context.Background(), "Absolute mouse coords - raw: (%d,%d), fixed: (%d,%d)", absDx, absDy, fixedX, fixedY)

		// 绝对鼠标命令格式，与Python完全一致：
		// [0, 0x04, 0x07, 0x02, buttons, x_low, x_high, y_low, y_high, wheel]
		wheelByte := byte(0)
		if wheel > 0 {
			wheelByte = 1 // 向上滚动
		} else if wheel < 0 {
			wheelByte = 255 // 向下滚动（无符号字节，-1表示为255）
		}
		cmd = []byte{
			0x00,                 // 保留字节
			CmdTypeMouseAbsolute, // 命令类型：绝对鼠标
			0x07,                 // 数据长度
			0x02,                 // 绝对模式标识
			buttons,              // 按键状态
			byte(fixedX & 0xFF),  // X坐标低字节
			byte(fixedX >> 8),    // X坐标高字节
			byte(fixedY & 0xFF),  // Y坐标低字节
			byte(fixedY >> 8),    // Y坐标高字节
			wheelByte,            // 滚轮：1表示向上，0表示向下或不动
		}
	} else {
		// 相对鼠标模式，与Python版本完全一致的坐标处理逻辑
		// Python代码：
		// def __fix_relative(self, value: int) -> int:
		//     assert MouseDelta.MIN <= value <= MouseDelta.MAX
		//     value = math.ceil(value / 3)
		//     return (value if value >= 0 else (255 + value))

		relDx := dx
		relDy := dy

		// 确保增量在有效范围内
		if relDx < MouseDeltaMin {
			relDx = MouseDeltaMin
		}
		if relDx > MouseDeltaMax {
			relDx = MouseDeltaMax
		}
		if relDy < MouseDeltaMin {
			relDy = MouseDeltaMin
		}
		if relDy > MouseDeltaMax {
			relDy = MouseDeltaMax
		}

		// 除以3并向上取整，与Python的math.ceil一致
		if relDx > 0 {
			relDx = (relDx + 2) / 3
		} else if relDx < 0 {
			relDx = -((-relDx + 2) / 3)
		}

		if relDy > 0 {
			relDy = (relDy + 2) / 3
		} else if relDy < 0 {
			relDy = -((-relDy + 2) / 3)
		}

		// 调整为无符号字节表示，与Python版本一致
		adjustedDx := byte(0)
		if relDx > 0 {
			adjustedDx = byte(relDx)
		} else if relDx < 0 {
			adjustedDx = byte(255 + relDx) // 与Python版本一致：(255 + value)
		}

		adjustedDy := byte(0)
		if relDy > 0 {
			adjustedDy = byte(relDy)
		} else if relDy < 0 {
			adjustedDy = byte(255 + relDy) // 与Python版本一致：(255 + value)
		}

		// 调整滚轮：1表示向上，255表示向下，0表示不动
		adjustedWheel := byte(0)
		if wheel > 0 {
			adjustedWheel = 1
		} else if wheel < 0 {
			adjustedWheel = 255
		}

		// 相对鼠标命令格式，与Python完全一致：
		// [0, 0x05, 0x05, 0x01, buttons, dx, dy, wheel]
		cmd = []byte{
			0x00,                 // 保留字节
			CmdTypeMouseRelative, // 命令类型：相对鼠标
			0x05,                 // 数据长度
			0x01,                 // 相对模式标识
			buttons,              // 按键状态
			adjustedDx,           // 调整后的X增量
			adjustedDy,           // 调整后的Y增量
			adjustedWheel,        // 调整后的滚轮
		}
	}

	// 发送命令
	return d.sendCommand(cmd)
}

// ProcessButton 处理鼠标按键事件，与Python版本一致
func (d *CH9329Device) ProcessButton(button byte, state bool) error {
	if state {
		d.mouseButtons |= button
	} else {
		d.mouseButtons &= ^button
	}

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	if d.absolute {
		return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
	} else {
		return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
	}
}

// ProcessMove 处理鼠标绝对移动事件，与Python版本一致
func (d *CH9329Device) ProcessMove(toX, toY int) error {
	d.mouseX = toX
	d.mouseY = toY

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
}

// ProcessRelativeMove 处理鼠标相对移动事件，与Python版本一致
func (d *CH9329Device) ProcessRelativeMove(deltaX, deltaY int) error {
	d.mouseDeltaX = deltaX
	d.mouseDeltaY = deltaY

	// 重置滚轮状态
	d.mouseWheel = 0

	// 发送鼠标报告
	return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
}

// ProcessWheel 处理鼠标滚轮事件，与Python版本一致
func (d *CH9329Device) ProcessWheel(deltaY int8) error {
	d.mouseWheel = deltaY

	// 发送鼠标报告
	if d.absolute {
		return d.SendMouseReport(d.mouseButtons, d.mouseX, d.mouseY, d.mouseWheel)
	} else {
		return d.SendMouseReport(d.mouseButtons, d.mouseDeltaX, d.mouseDeltaY, d.mouseWheel)
	}
}

// MoveMouse 移动鼠标
func (d *CH9329Device) MoveMouse(dx, dy int) error {
	return d.SendMouseReport(0x00, dx, dy, 0)
}

// ClickMouse 点击鼠标
func (d *CH9329Device) ClickMouse(button byte) error {
	// 按下
	if err := d.SendMouseReport(button, 0, 0, 0); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// 释放
	if err := d.SendMouseReport(0x00, 0, 0, 0); err != nil {
		return err
	}

	return nil
}

// ScrollMouse 滚动鼠标
func (d *CH9329Device) ScrollMouse(wheel int8) error {
	return d.SendMouseReport(0x00, 0, 0, wheel)
}
