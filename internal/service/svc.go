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
	VideoPath                 string
	VideoWidth                int
	VideoHeight               int
	VideoFPS                  uint32
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
	Ch9329Path := g.Cfg().MustGetWithCmd(s.CTX, `ch9329path`, "").String()

	switch hidMode {
	case "otg":
		go func() {
			udcControllerName, err := otg.WaitForUDC(0, 2*time.Second)
			if err != nil {
				g.Log().Errorf(context.Background(), "Failed to wait for UDC: %v", err)
				return
			}

			s.Gadget = otg.New(udcName, udcControllerName)
			_, err = s.Gadget.InitConfig()
			if err != nil {
				g.Log().Errorf(context.Background(), "Failed to init gadget: %v", err)
				return
			}

			km := otg.NewKM(s.Gadget)
			if err := km.AddKeyboard(); err != nil {
				g.Log().Errorf(context.Background(), "Failed to add keyboard: %v", err)
				return
			}
			if err := km.AddMouse(true, false); err != nil {
				g.Log().Errorf(context.Background(), "Failed to add absolute mouse: %v", err)
				return
			}
			if err := km.AddMouse(false, false); err != nil {
				g.Log().Errorf(context.Background(), "Failed to add relative mouse: %v", err)
				return
			}

			s.MSD = otg.NewMSD(s.Gadget)
			if err := s.MSD.AddMSD(); err != nil {
				g.Log().Errorf(context.Background(), "Failed to add MSD: %v", err)
				return
			}

			err = s.Gadget.StartUDC()
			if err != nil {
				g.Log().Errorf(context.Background(), "Failed to start UDC: %v", err)
				return
			}

			s.HID = otg.NewOTGKMHIDController()
			s.HID.SetAbsoluteMouse(true)
			if err := s.HID.Open(); err != nil {
				g.Log().Errorf(context.Background(), "Failed to open OTG HID: %v", err)
				return
			}

			g.Log().Info(context.Background(), "OTG device initialized successfully")

			otg.StartUDCWatcher(udcControllerName, func() {
				g.Log().Warning(context.Background(), "UDC device lost, stopping HID...")
				s.HID.Close()
				s.Gadget.CloseUDC()
			}, func() {
				g.Log().Info(context.Background(), "UDC device reconnected, reinitializing...")
				time.Sleep(2 * time.Second)
				_, err := s.Gadget.InitConfig()
				if err != nil {
					g.Log().Errorf(context.Background(), "Failed to reinitialize gadget: %v", err)
					return
				}
				km := otg.NewKM(s.Gadget)
				if err := km.AddKeyboard(); err != nil {
					g.Log().Errorf(context.Background(), "Failed to re-add keyboard: %v", err)
					return
				}
				if err := km.AddMouse(true, false); err != nil {
					g.Log().Errorf(context.Background(), "Failed to re-add absolute mouse: %v", err)
					return
				}
				if err := km.AddMouse(false, false); err != nil {
					g.Log().Errorf(context.Background(), "Failed to re-add relative mouse: %v", err)
					return
				}
				if err := s.MSD.AddMSD(); err != nil {
					g.Log().Errorf(context.Background(), "Failed to re-add MSD: %v", err)
					return
				}
				if err := s.Gadget.StartUDC(); err != nil {
					g.Log().Errorf(context.Background(), "Failed to restart UDC: %v", err)
					return
				}
				if err := s.HID.Open(); err != nil {
					g.Log().Errorf(context.Background(), "Failed to reopen HID: %v", err)
					return
				}
				g.Log().Info(context.Background(), "UDC device reinitialized successfully")
			})
		}()
	case "ch9329":
		go func() {
			actualDevicePath, err := ch9329.WaitForCH9329(Ch9329Path, 0, 2*time.Second)
			if err != nil {
				g.Log().Errorf(context.Background(), "Failed to wait for CH9329: %v", err)
				return
			}
			s.HID = ch9329.NewCH9329(actualDevicePath)
			s.HID.SetAbsoluteMouse(true)
			if err := s.HID.Open(); err != nil {
				g.Log().Errorf(context.Background(), "Failed to open CH9329: %v", err)
				return
			}

			g.Log().Info(context.Background(), "CH9329 device initialized successfully")

			ch9329.StartCH9329Watcher(actualDevicePath, func() {
				g.Log().Warning(context.Background(), "CH9329 device lost, closing...")
				s.HID.Close()
			}, func(newDevicePath string) {
				g.Log().Info(context.Background(), "CH9329 device reconnected: %s, reinitializing...", newDevicePath)
				s.HID = ch9329.NewCH9329(newDevicePath)
				s.HID.SetAbsoluteMouse(true)
				time.Sleep(2 * time.Second)
				if err := s.HID.Open(); err != nil {
					g.Log().Errorf(context.Background(), "Failed to reopen CH9329: %v", err)
					return
				}
				g.Log().Info(context.Background(), "CH9329 device reinitialized successfully")
			})
		}()
	}
}

// 初始化视频服务
func (s *SVC) initVideo() {
	s.VideoPath = g.Cfg().MustGetWithCmd(s.CTX, `videopath`, "").String()
	s.VideoWidth = g.Cfg().MustGetWithCmd(s.CTX, `videowidth`, "1920").Int()
	s.VideoHeight = g.Cfg().MustGetWithCmd(s.CTX, `videoheight`, "1080").Int()
	s.VideoFPS = g.Cfg().MustGetWithCmd(s.CTX, `videofps`, "30").Uint32()

	go func() {
		actualVideoPath, err := video.WaitForCamera(s.VideoPath, 30*time.Second, 2*time.Second)
		if err != nil {
			g.Log().Warningf(context.Background(), "No camera found in 30s, will retry later: %v", err)
		} else {
			s.VideoPath = actualVideoPath

			camera := video.NewV4LCamera()
			if err := camera.Open(actualVideoPath); err != nil {
				g.Log().Errorf(context.Background(), "Failed to open camera: %v", err)
			} else {
				s.Camera = camera

				s.Streamer = video.NewMJPEGStreamer(s.VideoWidth, s.VideoHeight)

				g.Log().Info(context.Background(), "Camera initialized successfully")

				go func() {
					for {
						if s.Streamer.IsStopped() {
							break
						}

						if s.Camera == nil {
							time.Sleep(500 * time.Millisecond)
							continue
						}

						if !s.Camera.GetStatus() {
							time.Sleep(500 * time.Millisecond)
							continue
						}

						frame, err := s.Camera.Capture()
						if err != nil {
							time.Sleep(100 * time.Millisecond)
							continue
						}
						if len(frame) < 5 {
							continue
						}
						s.Streamer.Broadcast(frame)
					}
				}()
			}
		}

		video.StartCameraWatcher(func() {
			g.Log().Warning(context.Background(), "Camera device lost, stopping streamer...")
			s.Streamer.Stop()
			s.Camera.Close()
		}, func(newDevicePath string) {
			g.Log().Info(context.Background(), "Camera device reconnected: %s, reinitializing...", newDevicePath)
			s.VideoPath = newDevicePath
			time.Sleep(2 * time.Second)
			camera := video.NewV4LCamera()
			if err := camera.Open(newDevicePath); err != nil {
				g.Log().Errorf(context.Background(), "Failed to reopen camera: %v", err)
				return
			}
			s.Camera = camera
			s.Streamer = video.NewMJPEGStreamer(s.VideoWidth, s.VideoHeight)

			go func() {
				for {
					if s.Streamer.IsStopped() {
						break
					}
					if s.Camera == nil {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					if !s.Camera.GetStatus() {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					frame, err := s.Camera.Capture()
					if err != nil {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					if len(frame) < 5 {
						continue
					}
					s.Streamer.Broadcast(frame)
				}
			}()

			g.Log().Info(context.Background(), "Camera device reinitialized successfully")
		})
	}()
}

// handleInterrupt 处理中断信号
func (s *SVC) handleInterrupt(camera video.Camera, streamer video.Streamer, gadget otg.GadgetInterface, done chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	g.Log().Info(context.Background(), "Stopping server...")

	g.Log().Info(context.Background(), "Stopping OTG watcher...")
	otg.StopUDCWatcher()
	g.Log().Info(context.Background(), "OTG watcher stopped")

	g.Log().Info(context.Background(), "Stopping CH9329 watcher...")
	ch9329.StopCH9329Watcher()
	g.Log().Info(context.Background(), "CH9329 watcher stopped")

	g.Log().Info(context.Background(), "Stopping camera watcher...")
	video.StopCameraWatcher()
	g.Log().Info(context.Background(), "Camera watcher stopped")

	if gadget != nil {
		g.Log().Info(context.Background(), "Removing gadget...")
		if err := gadget.Remove(); err != nil {
			g.Log().Errorf(context.Background(), "Error removing Gadget: %v", err)
		}
	}

	if camera != nil {
		go func() {
			camera.Close()
		}()
		time.Sleep(100 * time.Millisecond)
	}

	if streamer != nil {
		g.Log().Info(context.Background(), "Stopping streamer...")
		streamer.Stop()
		g.Log().Info(context.Background(), "Streamer stopped")
	}

	g.Log().Info(context.Background(), "All cleanup operations completed, exiting...")
	close(done)
}
