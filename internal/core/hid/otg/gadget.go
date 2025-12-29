package otg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Gadget 管理USB Gadget的配置
type Gadget struct {
	gadgetPath  string
	profilePath string
	metaPath    string
	epsMax      int
	epsUsed     int
	hidInstance int
	msdInstance int
	logger      func(string, ...interface{})
}

// NewGadget 创建一个新的Gadget实例
func NewGadget(gadgetPath, profilePath, metaPath string, eps int, logger func(string, ...interface{})) GadgetInterface {
	if logger == nil {
		logger = func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		}
	}

	return &Gadget{
		gadgetPath:  gadgetPath,
		profilePath: profilePath,
		metaPath:    metaPath,
		epsMax:      eps,
		epsUsed:     0,
		hidInstance: 0,
		msdInstance: 0,
		logger:      logger,
	}
}

// mkdir 创建目录
func (g *Gadget) mkdir(path string) error {
	g.logger("MKDIR -- %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// write 写入文件
func (g *Gadget) write(path string, value string, optional bool) error {
	if optional {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			g.logger("WRITE -- [SKIPPED] %s", path)
			return nil
		}
	}

	g.logger("WRITE -- %s", path)
	return os.WriteFile(path, []byte(value), 0644)
}

// writeBytes 写入字节数据
func (g *Gadget) writeBytes(path string, data []byte) error {
	g.logger("WRITE -- %s", path)
	return os.WriteFile(path, data, 0644)
}

// symlink 创建符号链接
func (g *Gadget) symlink(src, dest string) error {
	g.logger("SYMLINK - %s --> %s", dest, src)
	return os.Symlink(src, dest)
}

// rmdir 删除目录
func (g *Gadget) rmdir(path string) error {
	g.logger("RMDIR -- %s", path)
	return os.Remove(path)
}

// unlink 删除文件或符号链接
func (g *Gadget) unlink(path string, optional bool) error {
	if optional {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			g.logger("RM ------ [SKIPPED] %s", path)
			return nil
		}
	}

	g.logger("RM ------ %s", path)
	return os.Remove(path)
}

// createFunction 创建功能目录
func (g *Gadget) createFunction(funcName string) (string, error) {
	funcPath := filepath.Join(g.gadgetPath, "functions", funcName)
	err := g.mkdir(funcPath)
	return funcPath, err
}

// startFunction 启动功能
func (g *Gadget) startFunction(funcName string, eps int) error {
	if g.epsMax-g.epsUsed >= eps {
		funcPath := filepath.Join(g.gadgetPath, "functions", funcName)
		destPath := filepath.Join(g.profilePath, funcName)
		if _, err := os.Stat(destPath); err == nil {
			err = os.Remove(destPath)
			g.logger("RM ------ %s", destPath)
		}

		err := g.symlink(funcPath, destPath)
		if err != nil {
			return err
		}
		g.epsUsed += eps
	} else {
		g.logger("Will not be started: No available endpoints")
	}
	return nil
}

// createMeta 创建元数据
func (g *Gadget) createMeta(funcName, desc string, eps int) error {
	meta := map[string]interface{}{
		"function":    funcName,
		"description": desc,
		"endpoints":   eps,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(g.metaPath, funcName+"@meta.json")
	return g.writeBytes(path, data)
}
