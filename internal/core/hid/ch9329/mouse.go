package ch9329

import (
	"log"
	"time"
)

// CH9329 鼠标按键
const (
	MouseLeft    = 0x01
	MouseRight   = 0x02
	MouseMiddle  = 0x04
	MouseBack    = 0x08 // Back/Up
	MouseForward = 0x10 // Forward/Down
)

// MouseRange 定义鼠标坐标范围（与Python版本一致）
const (
	MouseRangeMin = 0
	MouseRangeMax = 0xFFFF
)

// MouseDelta 定义鼠标增量范围（与Python版本一致）
const (
	MouseDeltaMin = -127
	MouseDeltaMax = 127
)

// SetAbsoluteMouse 设置鼠标是否使用绝对模式
func (d *CH9329Device) SetAbsoluteMouse(absolute bool) error {
	log.Printf("CH9329: Setting absolute mouse mode to: %v", absolute)
	d.absolute = absolute
	return nil
}

// IsAbsoluteMouse 检查鼠标是否使用绝对模式
func (d *CH9329Device) IsAbsoluteMouse() bool {
	return d.absolute
}

// SendMouseReport 发送鼠标HID报告，与Python版本完全一致
func (d *CH9329Device) SendMouseReport(buttons byte, dx, dy int, wheel int8) error {
	// 绝对模式下，所有移动事件都记录日志；相对模式下，只有按键或滚轮操作才记录日志
	if d.absolute || buttons > 0 || wheel != 0 || dx != 0 || dy != 0 {
		log.Printf("CH9329: SendMouseReport called - absolute: %v, buttons: 0x%02x, dx: %d, dy: %d, wheel: %d", d.absolute, buttons, dx, dy, wheel)
	}
	var cmd []byte

	if d.absolute {
		// 绝对鼠标模式，处理客户端发送的归一化坐标
		// 前端发送的是归一化坐标（0-32767），需要转换为原始屏幕坐标
		// 原始屏幕分辨率假设为1920x1080，可根据实际情况调整
		const (
			ScreenWidth  = 1920
			ScreenHeight = 1080
		)

		var rawDx, rawDy int
		if dx >= 0 && dx <= 32767 && dy >= 0 && dy <= 32767 {
			// 前端发送的是归一化坐标（0-32767），转换为原始屏幕坐标
			rawDx = int(float64(dx) / 32767 * ScreenWidth)  // 转换为原始X坐标
			rawDy = int(float64(dy) / 32767 * ScreenHeight) // 转换为原始Y坐标
		} else {
			// 如果不是归一化坐标，直接使用
			rawDx = dx
			rawDy = dy
		}

		// 限制原始坐标在屏幕范围内
		if rawDx < 0 {
			rawDx = 0
		}
		if rawDx > ScreenWidth {
			rawDx = ScreenWidth
		}
		if rawDy < 0 {
			rawDy = 0
		}
		if rawDy > ScreenHeight {
			rawDy = ScreenHeight
		}

		// 转换为CH9329所需的0-65535范围坐标
		absDx := (rawDx * 65535) / ScreenWidth
		absDy := (rawDy * 65535) / ScreenHeight

		// 与Python版本一致：将坐标除以8并向上取整
		// Python代码：to_fixed = math.ceil(MouseRange.remap(value, 0, MouseRange.MAX) / 8)
		fixedX := (absDx + 7) / 8 // 向上取整的简化计算
		fixedY := (absDy + 7) / 8 // 向上取整的简化计算

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
		// 相对鼠标模式，dx和dy是增量
		// 与Python相同的坐标处理逻辑：除以3并向上取整
		relDx := dx
		relDy := dy
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

		// 除以3并向上取整
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

		// 调整为无符号字节表示
		adjustedDx := byte(0)
		if relDx > 0 {
			adjustedDx = byte(relDx)
		} else if relDx < 0 {
			adjustedDx = byte(256 + relDx)
		}

		adjustedDy := byte(0)
		if relDy > 0 {
			adjustedDy = byte(relDy)
		} else if relDy < 0 {
			adjustedDy = byte(256 + relDy)
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
