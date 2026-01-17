package otgm

import (
	"fmt"
	"os"
	"sort"
)

// FindUDC 查找可用的UDC控制器
func FindUDC() (string, error) {
	// 读取UDC设备列表
	udcList, err := os.ReadDir("/sys/class/udc")
	if err != nil {
		return "", fmt.Errorf("Failed to read UDC list: %w", err)
	}

	// 如果没有UDC设备，返回错误
	if len(udcList) == 0 {
		return "", fmt.Errorf("No UDC devices found")
	}

	// 按字母顺序返回第一个UDC
	udcNames := make([]string, len(udcList))
	for i, udc := range udcList {
		udcNames[i] = udc.Name()
	}
	sort.Strings(udcNames)

	return udcNames[0], nil
}
