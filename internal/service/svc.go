package service

import (
	"context"
	"opskvm/internal/service/files"
	"opskvm/internal/service/hid"
	"opskvm/internal/service/hid/ch9329"
	"opskvm/internal/service/hid/otg"
	"opskvm/internal/service/mirrors"
	"opskvm/internal/service/sqlite"
	"opskvm/internal/service/video"
	"opskvm/internal/service/wol"
	"os"
	"os/signal"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

var Svc = &SVC{}

type SVC struct {
	Camera                    video.Camera
	Streamer                  video.Streamer
	Gadget                    otg.GadgetInterface
	HID                       hid.KMHIDController
	MSD                       otg.MSDInterface
	CTX                       g.Ctx
	ChunkUpload               files.IChunkUploadService
	FileManager               files.IFileManagerService
	MirrorsChunkUploadService mirrors.IMirrorsChunkUploadService
	MirrorsManagerService     mirrors.IMirrorsManagerService
	SQL                       sqlite.Sqliter
	WOL                       wol.WOLInterface
	RootPath                  string
	Done                      chan struct{}
}

func New(rootPath string, debug bool) {
	Svc = &SVC{}
	Svc.RootPath = rootPath

	// 初始化OTG
	Svc.initOTG()
	// 初始化视频
	Svc.initVideo()
	// 初始化文件管理
	Svc.initFilesManager(rootPath)
	// 初始化ISO镜像管理
	Svc.initIsoManager(rootPath)
	// 初始化数据库
	Svc.initSqlite(rootPath, debug)
	// 初始化WOL
	Svc.WOL = wol.NewWOL()
	// 启动服务中断处理
	done := make(chan struct{})
	go Svc.handleInterrupt(Svc.Camera, Svc.Streamer, Svc.Gadget, done)

	// 将通道存储在上下文中，以便主程序等待
	Svc.Done = done
}

func (s *SVC) initSqlite(rootPath string, debug bool) {
	s.SQL = sqlite.NewSqlite(rootPath, debug)
}

func (s *SVC) initFilesManager(rootPath string) {
	s.ChunkUpload = files.GetChunkUploadService(rootPath)
	s.FileManager = files.GetFileManagerService(rootPath)
}

func (s *SVC) initIsoManager(rootPath string) {
	s.MirrorsChunkUploadService = mirrors.GetMirrorsChunkUploadService(rootPath)
	s.MirrorsManagerService = mirrors.GetMirrorsManagerService(rootPath)
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
		s.HID.SetAbsoluteMouse(true)
		if err := s.HID.Open(); err != nil {
			g.Log().Errorf(context.Background(), "Failed to open OTG HID: %v", err)
		}
	case "ch9329":
		s.HID = ch9329.NewCH9329(Ch9329Path)
		s.HID.SetAbsoluteMouse(true)
		if err := s.HID.Open(); err != nil {
			g.Log().Errorf(context.Background(), "Failed to open CH9329: %v", err)
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
	if err := camera.ApplyConfig(VideoWidth, VideoHeight, VideoFPS); err != nil {
		g.Log().Warningf(context.Background(), "Failed to apply camera config: %v, using device default", err)
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
				g.Log().Errorf(context.Background(), "Capture error: %v", err)
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

	g.Log().Info(context.Background(), "Stopping server...")
	if gadget != nil {
		if err := gadget.Remove(); err != nil {
			g.Log().Errorf(context.Background(), "Error removing Gadget: %v", err)
		}
	}
	streamer.Stop()
	camera.Close()
	streamer.Wait()

	g.Log().Info(context.Background(), "All cleanup operations completed, exiting...")
	close(done)
}
