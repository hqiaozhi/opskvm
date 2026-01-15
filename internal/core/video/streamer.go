package video

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"sync"
)

// Streamer 定义帧数据分发的接口
type Streamer interface {
	AddClient() *Client
	RemoveClient(clt *Client)
	Broadcast(frame []byte)
	Stop()
	IsStopped() bool
	Wait()
	Pause()
	Resume()

	// CompressFrame 对视频帧进行JPEG压缩 - 未启用，延时太高
	CompressFrame(frame []byte, quality int) ([]byte, error)
}

// 压缩配置
const (
	// JPEG压缩质量 (1-100), 值越低压缩率越高
	DefaultJPEGQuality = 70
)

// Client 表示流客户端
type Client struct {
	i  int
	Ch chan []byte
}

// MJPEGStreamer 是Streamer接口的MJPEG实现
type MJPEGStreamer struct {
	mu      sync.Mutex
	clients []*Client
	stopped bool
	paused  bool
	quit    chan int
	Blank   []byte
}

// NewMJPEGStreamer 创建MJPEGStreamer实例
func NewMJPEGStreamer(width, height int) *MJPEGStreamer {
	blankImg := image.NewRGBA(image.Rect(0, 0, width, height))
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, blankImg, nil)

	return &MJPEGStreamer{
		quit:  make(chan int, 1),
		Blank: buf.Bytes(),
	}
}

func (s *MJPEGStreamer) AddClient() *Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}

	clt := &Client{
		i:  len(s.clients),
		Ch: make(chan []byte, 1),
	}
	clt.Ch <- s.Blank
	s.clients = append(s.clients, clt)
	return clt
}

func (s *MJPEGStreamer) RemoveClient(clt *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clt.i == -1 || clt.i >= len(s.clients) {
		return
	}

	i := clt.i
	last := len(s.clients) - 1
	s.clients[i] = s.clients[last]
	s.clients[i].i = i
	s.clients[last] = nil
	s.clients = s.clients[:last]
	clt.i = -1

	if s.stopped && len(s.clients) == 0 {
		s.quit <- 1
	}
}

func (s *MJPEGStreamer) Broadcast(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped || s.paused {
		return
	}

	for _, clt := range s.clients {
		select {
		case clt.Ch <- frame:
		case <-clt.Ch:
			clt.Ch <- frame
		}
	}
}

func (s *MJPEGStreamer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stopped = true
	for _, clt := range s.clients {
		close(clt.Ch)
	}

	if len(s.clients) == 0 {
		s.quit <- 1
	}
}

func (s *MJPEGStreamer) IsStopped() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopped
}

func (s *MJPEGStreamer) Wait() {
	<-s.quit
}

func (s *MJPEGStreamer) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = true
	log.Println("Stream paused for config update")
}

func (s *MJPEGStreamer) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = false
	log.Println("Stream resumed after config update")
}

// CompressFrame 对视频帧进行JPEG压缩
func (s *MJPEGStreamer) CompressFrame(frame []byte, quality int) ([]byte, error) {
	// 解码原始JPEG帧
	img, err := jpeg.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}

	// 编码为新的JPEG帧，使用指定的压缩质量
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
