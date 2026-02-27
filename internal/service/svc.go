package service

import (
	"log"
	"opskvm/internal/service/hid"
	"opskvm/internal/service/hid/ch9329"
	"opskvm/internal/service/hid/otg"
	"opskvm/internal/service/video"
	"os"
	"os/signal"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

var Svc = &SVC{}

type SVC struct {
	Camera   video.Camera
	Streamer video.Streamer
	Gadget   otg.GadgetInterface
	HID      hid.KMHIDController
	MSD      otg.MSDInterface
	CTX      g.Ctx
	Done     chan struct{}
}

func New() {
	Svc = &SVC{}
	Svc.initOTG()
	Svc.initVideo()

	// 启动服务中断处理
	done := make(chan struct{})
	go Svc.handleInterrupt(Svc.Camera, Svc.Streamer, Svc.Gadget, done)

	// 将通道存储在上下文中，以便主程序等待
	Svc.Done = done
}

func (s *SVC) initOTG() {

	hidMode := g.Cfg().MustGetWithCmd(s.CTX, `hidmode`, "otg").String()
	udcName := g.Cfg().MustGetWithCmd(s.CTX, `udcname`, "opskvm").String()
	Ch9329Path := g.Cfg().MustGetWithCmd(s.CTX, `ch9329path`, "/dev/ttyUSB0").String()

	// 获取键盘和鼠标控制器
	switch hidMode {
	case "otg":
		// 初始化USB Gadget服务
		udcControllerName, err := otg.FindUDC()
		if err != nil {
			panic(err)
		}
		s.Gadget = otg.New(udcName, udcControllerName)
		_, err = s.Gadget.InitConfig()
		if err != nil {
			panic(err)
		}

		// 初始化键盘和鼠标
		km := otg.NewKM(s.Gadget)
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
		s.MSD = otg.NewMSD(s.Gadget)
		if err := s.MSD.AddMSD(); err != nil {
			panic(err)
		}

		// 启动USB Gadget服务
		err = s.Gadget.StartUDC()
		if err != nil {
			panic(err)
		}

		s.HID = otg.NewOTGKMHIDController()
		s.HID.SetAbsoluteMouse(true) // 与CH9329保持一致，默认使用绝对鼠标模式
		if err := s.HID.Open(); err != nil {
			log.Printf("Failed to open OTG HID: %v", err)
		}
	case "ch9329":
		s.HID = ch9329.NewCH9329(Ch9329Path)
		s.HID.SetAbsoluteMouse(true)
		if err := s.HID.Open(); err != nil {
			log.Printf("Failed to open CH9329: %v", err)
		}
	}
}

// 初始化视频服务
func (s *SVC) initVideo() {
	VideoPath := g.Cfg().MustGetWithCmd(s.CTX, `videopath`, "").String()
	VideoWidth := g.Cfg().MustGetWithCmd(s.CTX, `videowidth`, "1920").Int()
	VideoHeight := g.Cfg().MustGetWithCmd(s.CTX, `videoheight`, "1080").Int()
	VideoFPS := g.Cfg().MustGetWithCmd(s.CTX, `videofps`, "30").Uint32()

	camera := video.NewV4LCamera()
	if err := camera.Open(VideoPath); err != nil {
		panic(err)
	}
	// 应用配置到摄像头
	if err := camera.ApplyConfig(VideoWidth, VideoHeight, VideoFPS); err != nil {
		log.Printf("Warning: Failed to apply camera config: %v, using device default", err)
	}
	s.Camera = camera

	// 初始化流分发器
	s.Streamer = video.NewMJPEGStreamer(VideoWidth, VideoHeight)

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

// handleInterrupt 处理中断信号
func (s *SVC) handleInterrupt(camera video.Camera, streamer video.Streamer, gadget otg.GadgetInterface, done chan struct{}) {
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
