package svc

import (
	"context"
	"errors"
	"fmt"
	"opskvm/internal/conf"
	"opskvm/internal/core/files"
	"opskvm/internal/core/usbgadget"
	"opskvm/internal/core/video"
	"opskvm/internal/utils/jwt"
	"opskvm/internal/utils/resp"
	"os"
)

type SvcContext struct {
	Conf      *conf.Config
	Files     *files.LinuxFilesManager
	Video     *video.VideoStreamer
	USBGadget *usbgadget.LinuxUSBGadget
	RESP      *resp.Resp
	JWT       *jwt.JwtService
}

func New(ctx context.Context) *SvcContext {
	s := &SvcContext{}

	// 加载配置
	config, v, _ := conf.New()
	config.WatchConfigChanges(v)
	s.Conf = config

	// 初始化文件服务
	// 判断目录是否存在，不存在则创建
	path := "/opt/opskvm"
	err := EnsureDirExists(path)
	if err != nil {
		panic(err)
	}
	f, err := files.New(path)
	if err != nil {
		panic(err)
	}
	s.Files = f

	// 初始化视频服务
	video, err := video.New("/dev/video0", 1920, 1080)
	if err != nil {
		panic(err)
	}
	s.Video = video

	// 初始化USB Gadget服务
	s.USBGadget = usbgadget.New("", "test-gadget", 8)

	// 响应
	R := resp.New()
	s.RESP = R

	// 初始化JWT服务
	jwt := jwt.New(&config.JWT)
	s.JWT = jwt

	return s
}

// EnsureDirExists 检查目录是否存在，不存在则创建（支持多级目录）
// dirPath: 目标目录路径
// 返回值: 成功返回nil，失败返回具体错误（如权限不足、路径是文件等）
func EnsureDirExists(dirPath string) error {
	// 1. 获取目录状态
	stat, err := os.Stat(dirPath)
	if err != nil {
		// 2. 目录不存在则创建（MkdirAll支持多级目录，如 "a/b/c"）
		if errors.Is(err, os.ErrNotExist) {
			// 权限说明：0o755 表示所有者可读可写可执行，其他用户可读可执行
			// Go 1.15+ 推荐用 0o 前缀表示八进制（替代旧的 0 前缀）
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}
			fmt.Printf("目录创建成功: %s\n", dirPath)
			return nil
		}
		// 其他错误（如权限不足、路径无效等）
		return fmt.Errorf("stat directory failed: %w", err)
	}

	// 3. 路径存在但不是目录（比如是同名文件）
	if !stat.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dirPath)
	}

	fmt.Printf("目录已存在: %s\n", dirPath)
	return nil
}
