package webshell

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/gogf/gf/v2/frame/g"
)

type Pty struct {
	Ptmx *os.File
	Pty  *os.File
	Cmd  *exec.Cmd
	once sync.Once
	ch   chan []byte
	done chan struct{}
}

func NewPty() *Pty {
	return &Pty{
		ch:   make(chan []byte, 100),
		done: make(chan struct{}),
	}
}

func (p *Pty) Start() {
	cmd := exec.Command("bash", "-c", "cd /root && bash -l")
	cmd.Env = []string{
		"TERM=xterm-256color",
		"HOME=/root",
		"USER=root",
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		g.Log().Errorf(context.TODO(), "pty.Start error: %v", err)
		return
	}

	p.Ptmx = ptmx
	p.Cmd = cmd
	p.Pty = ptmx

	go p.readOutput()
}

func (p *Pty) readOutput() {
	buf := make([]byte, 8192)
	for {
		n, err := p.Ptmx.Read(buf)
		if err != nil {
			close(p.ch)
			return
		}
		data := bytes.Clone(buf[:n])
		select {
		case p.ch <- data:
		case <-p.done:
			return
		}
	}
}

func (p *Pty) Write(data []byte) (int, error) {
	if p.Ptmx == nil {
		return 0, nil
	}
	return p.Ptmx.Write(data)
}

func (p *Pty) ReadCh() <-chan []byte {
	return p.ch
}

func (p *Pty) Resize(cols, rows uint16) {
	if p.Ptmx == nil {
		return
	}
	pty.Setsize(p.Ptmx, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
}

func (p *Pty) Close() {
	p.once.Do(func() {
		close(p.done)
		if p.Cmd != nil {
			p.Cmd.Process.Kill()
			p.Cmd.Wait()
		}
		if p.Ptmx != nil {
			p.Ptmx.Close()
		}
	})
}
