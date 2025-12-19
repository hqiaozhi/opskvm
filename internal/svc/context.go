package svc

import (
	"context"
	"opskvm/internal/conf"
	"opskvm/internal/core/files"
	"opskvm/internal/core/usbgadget"
	"opskvm/internal/core/video"
	"opskvm/internal/utils/jwt"
	"opskvm/internal/utils/resp"
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
