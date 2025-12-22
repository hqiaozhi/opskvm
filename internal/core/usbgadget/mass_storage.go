package usbgadget

import (
	"fmt"
	"os"
	"path/filepath"
)

// setupMassStorage 配置ISO大容量存储功能
func (g *LinuxUSBGadget) setupMassStorage() error {
	if err := os.MkdirAll(g.msFuncDir, 0755); err != nil {
		return fmt.Errorf("创建大容量存储目录失败: %w", err)
	}

	// ISO挂载核心配置
	msConfigs := map[string]string{
		"lun.0/file":           g.isoPath, // 绑定ISO
		"lun.0/cdrom":          "1",       // 识别为光盘（0=U盘）
		"lun.0/ro":             "1",       // 只读（ISO必须）
		"lun.0/removable":      "1",       // 可移除设备
		"lun.0/inquiry_string": "Go ISO",  // 设备标识
	}
	for path, val := range msConfigs {
		fullPath := filepath.Join(g.msFuncDir, path)
		if err := os.WriteFile(fullPath, []byte(val), 0644); err != nil {
			return fmt.Errorf("写入大容量存储配置[%s]失败: %w", path, err)
		}
	}
	return nil
}
