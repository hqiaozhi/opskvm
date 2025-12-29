package otg

import (
	"fmt"
)

// addMSD 添加MSD功能
// 初始化之后 对端被识别大小为0的光驱
func (g *Gadget) addMSD(start bool) error {
	eps := 3

	// 处理驱动程序
	realDriver := "mass_storage"

	funcName := fmt.Sprintf("%s.usb0", realDriver)
	_, err := g.createFunction(funcName)
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
	err = g.createMeta(funcName, "Mass Storage", eps)
	if err != nil {
		return err
	}

	return nil
}
