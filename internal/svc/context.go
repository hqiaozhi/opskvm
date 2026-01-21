package svc

import (
	"context"
	"log"
	"opskvm/internal/conf"
	"opskvm/internal/core/dhcp"
	"opskvm/internal/core/files"
	"opskvm/internal/core/hid"
	"opskvm/internal/core/hid/ch9329"
	"opskvm/internal/core/hid/otgm"
	"opskvm/internal/core/video"
	"opskvm/internal/utils/jwt"
	"opskvm/internal/utils/resp"
	"os"
	"os/signal"
	"time"
)

type SvcContext struct {
	Conf     *conf.Config
	Files    *files.LinuxFilesManager
	Camera   video.Camera
	Streamer video.Streamer
	RESP     *resp.Resp
	JWT      *jwt.JwtService
	Gadget   otgm.GadgetInterface
	KMHID    hid.KMHIDController
	MSD      otgm.MSDInterface
	NET      otgm.NetInterface
	Done     chan struct{} // 用于通知主程序中断信号已处理
}

func New(ctx context.Context) *SvcContext {
	s := &SvcContext{}

	// 加载配置
	s.loadConfig()

	// 初始化文件服务
	s.initFiles()

	// 初始化视频服务
	s.initVideo()

	// 响应
	R := resp.New()
	s.RESP = R

	// 初始化JWT服务
	jwt := jwt.New(&s.Conf.JWT)
	s.JWT = jwt

	// 初始OTG
	s.initOTG()

	// 启动服务中断处理
	// 创建一个通道，用于通知主程序中断信号已处理
	done := make(chan struct{})
	go s.handleInterrupt(s.Camera, s.Streamer, s.Gadget, done)

	// 将通道存储在上下文中，以便主程序等待
	s.Done = done

	return s
}

func (s *SvcContext) initNet() {

}

func (s *SvcContext) initOTG() {
	// 获取键盘和鼠标控制器
	switch s.Conf.App.KMhidMode {
	case "otg":
		// 初始化USB Gadget服务
		udcName, err := otgm.FindUDC()
		if err != nil {
			panic(err)
		}
		s.Gadget = otgm.New(s.Conf.App.Name, udcName)
		_, err = s.Gadget.InitConfig()
		if err != nil {
			panic(err)
		}

		// 初始化键盘和鼠标
		km := otgm.NewKM(s.Gadget)
		if err := km.AddKeyboard(); err != nil {
			panic(err)
		}
		// 鼠标绝对模式（先创建，占用/dev/hidg1）
		if err := km.AddMouse(true, false); err != nil {
			panic(err)
		}
		// 鼠标相对模式（后创建，占用/dev/hidg2）
		if err := km.AddMouse(false, false); err != nil {
			panic(err)
		}

		// 添加大CD/DVD服务
		// s.MSD = otgm.NewMSD(s.Gadget)
		// if err := s.MSD.AddMSD(); err != nil {
		// 	panic(err)
		// }

		// 添加网络功能
		s.NET = otgm.NewNET(s.Gadget)
		_, err = s.NET.AddNET()
		if err != nil {
			panic(err)
		}

		// 网卡用于启动DHCP服务器(后台启动)
		go dhcp.Start("usb0")

		// 启动USB Gadget服务
		err = s.Gadget.StartUDC()
		if err != nil {
			panic(err)
		}

		s.KMHID = otgm.NewOTGKMHIDControl()
		s.KMHID.SetAbsoluteMouse(true) // 与CH9329保持一致，默认使用绝对鼠标模式
		if err := s.KMHID.Open(); err != nil {
			log.Printf("Failed to open OTG HID: %v", err)
		}
	case "ch9329":
		s.KMHID = ch9329.NewCH9329(s.Conf.App.Ch9329Path)
		s.KMHID.SetAbsoluteMouse(true)
		if err := s.KMHID.Open(); err != nil {
			log.Printf("Failed to open CH9329: %v", err)
		}
	}
}

func (s *SvcContext) initFiles() {
	// 判断目录是否存在，不存在则创建
	path := "/opt/opskvm"
	f, err := files.New(path)
	if err != nil {
		panic(err)
	}
	s.Files = f
}

// 初始化视频服务
func (s *SvcContext) initVideo() {
	camera := video.NewV4LCamera()
	if err := camera.Open(s.Conf.App.VideoPath); err != nil {
		panic(err)
	}
	// 应用配置到摄像头
	if err := camera.ApplyConfig(s.Conf.App.VideoWidth, s.Conf.App.VideoHeight, uint32(s.Conf.App.VideoFPS)); err != nil {
		log.Printf("Warning: Failed to apply camera config: %v, using device default", err)
	}
	s.Camera = camera

	// 初始化流分发器
	s.Streamer = video.NewMJPEGStreamer(s.Conf.App.VideoWidth, s.Conf.App.VideoHeight)

	// 启动帧采集goroutine
	go func() {
		for {
			if s.Streamer.IsStopped() {
				break
			}

			// 检查摄像头状态，如果关闭则暂停采集
			if !s.Camera.GetStatus() {
				time.Sleep(500 * time.Millisecond) // 等待500ms后再次检查
				continue
			}

			frame, err := s.Camera.Capture()
			if err != nil {
				log.Println("Capture error:", err)
				// 短暂重试，避免设备重启导致的瞬时错误
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if len(frame) < 5 {
				continue
			}
			// 广播JPEG帧
			s.Streamer.Broadcast(frame)
		}
	}()
}

// 加载配置
func (s *SvcContext) loadConfig() {
	config, v, _ := conf.New()
	config.WatchConfigChanges(v)
	s.Conf = config
}

// handleInterrupt 处理中断信号
func (s *SvcContext) handleInterrupt(camera video.Camera, streamer video.Streamer, gadget otgm.GadgetInterface, done chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	log.Println("Stopping server...")
	// 移除旧的Gadget目录
	if gadget != nil {
		if err := gadget.Remove(); err != nil {
			log.Printf("Error removing Gadget: %v", err)
		}
	}
	// 停止流分发器
	streamer.Stop()
	// 关闭摄像头
	camera.Close()
	// 等待流分发器停止
	streamer.Wait()

	// 所有清理操作完成后，关闭通道通知主程序
	log.Println("All cleanup operations completed, exiting...")
	close(done)
}
