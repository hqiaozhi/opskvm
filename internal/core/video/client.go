package video

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/korandiz/v4l"
)

var (
	mu      sync.Mutex
	clients []*Client
	stopped bool
	quit    = make(chan int, 1)
	blank   []byte
)

type Client struct {
	i  int
	CH chan []byte
}

func (clt *Client) Remove() {
	mu.Lock()
	defer mu.Unlock()
	i := clt.i
	last := len(clients) - 1
	clients[i] = clients[last]
	clients[i].i = i
	clients[last] = nil
	clients = clients[:last]
	clt.i = -1
	if stopped && len(clients) == 0 {
		quit <- 1
	}
}

func handleInterrupt() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Stopping...")
	mu.Lock()
	stopped = true
	if len(clients) == 0 {
		quit <- 1
	}
	// 不再直接关闭通道，而是等待客户端自然断开
	mu.Unlock()
	<-quit
	os.Exit(0)
}

func stream(cam *v4l.Device) {
	for {
		buf, err := cam.Capture()
		if err != nil {
			log.Println("Capture:", err)
			// 捕获失败时重试，而不是立即中断程序
			continue
		}
		b := make([]byte, buf.Size())
		buf.ReadAt(b, 0)
		mu.Lock()
		if stopped {
			mu.Unlock()
			break
		}
		for _, clt := range clients {
			select {
			case clt.CH <- b:
			case <-clt.CH:
				clt.CH <- b
			}
		}
		mu.Unlock()
	}
}
