package otg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// findUDC 查找可用的UDC设备
func findUDC(sysFSPrefix string, udcName string) (string, error) {
	udcListPath := filepath.Join(sysFSPrefix, "sys/class/udc")
	udcList, err := os.ReadDir(udcListPath)
	if err != nil {
		return "", fmt.Errorf("Failed to read UDC list: %w", err)
	}

	// 如果指定了UDC名称，直接返回
	if udcName != "" {
		for _, udc := range udcList {
			if udc.Name() == udcName {
				return udcName, nil
			}
		}
		return "", fmt.Errorf("UDC %s not found", udcName)
	}

	// 否则按字母顺序返回第一个UDC
	if len(udcList) == 0 {
		return "", fmt.Errorf("No UDC devices found")
	}

	udcNames := make([]string, len(udcList))
	for i, udc := range udcList {
		udcNames[i] = udc.Name()
	}
	sort.Strings(udcNames)

	return udcNames[0], nil
}

// getGadgetPath 构建gadget的路径
func getGadgetPath(sysFSPrefix string, gadgetName string, parts ...string) string {
	pathParts := []string{sysFSPrefix, "sys/kernel/config/usb_gadget", gadgetName}
	pathParts = append(pathParts, parts...)
	return filepath.Join(pathParts...)
}

// validateMAC 验证MAC地址格式
func validateMAC(mac string) bool {
	if mac == "" {
		return true
	}
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return false
	}
	for _, part := range parts {
		if len(part) != 2 {
			return false
		}
		if _, err := fmt.Sscanf(part, "%x", &[]byte{}); err != nil {
			return false
		}
	}
	return true
}

// createLogger 创建默认的日志记录器
func createLogger() func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}
}
