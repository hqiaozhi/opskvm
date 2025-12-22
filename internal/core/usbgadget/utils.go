package usbgadget

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeFile 写入文件到Gadget目录
func (g *LinuxUSBGadget) writeFile(path, content string) error {
	fullPath := filepath.Join(g.gadgetDir, path)
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// contains 字符串包含判断
func (g *LinuxUSBGadget) contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// bindFunction 绑定功能到配置项
func (g *LinuxUSBGadget) bindFunction(configDir, funcName string) error {
	funcPath := filepath.Join(g.gadgetDir, "functions", funcName)
	targetPath := filepath.Join(g.gadgetDir, configDir, funcName)
	if err := os.Symlink(funcPath, targetPath); err != nil && !os.IsExist(err) {
		return fmt.Errorf("创建符号链接失败: %w", err)
	}
	return nil
}
