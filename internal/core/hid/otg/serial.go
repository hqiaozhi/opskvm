package otg

// addSerial 添加串口功能
func (g *Gadget) addSerial(start bool) error {
	eps := 3
	funcName := "acm.usb0"

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
	err = g.createMeta(funcName, "Serial Port", eps)
	if err != nil {
		return err
	}

	return nil
}
