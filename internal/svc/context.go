package svc

import (
	"context"
	"log"
	"opskvm/internal/conf"
	"opskvm/internal/core/files"
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

	// 启动视频服务中断处理
	go handleInterrupt(s.Camera, s.Streamer)

	// 初始化USB Gadget服务

	// 响应
	R := resp.New()
	s.RESP = R

	// 初始化JWT服务
	jwt := jwt.New(&config.JWT)
	s.JWT = jwt

	return s
}

// handleInterrupt 处理中断信号
func handleInterrupt(camera video.Camera, streamer video.Streamer) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	log.Println("Stopping server...")
	streamer.Stop()
	camera.Close()
	streamer.Wait()
	os.Exit(0)
}
