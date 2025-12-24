package video

import (
	"context"
	"errors"
	"opskvm/internal/core/video"
	"opskvm/internal/svc"
	"sync"
)

// Logic 视频逻辑结构体
type Logic struct {
	svcCtx        *svc.SvcContext
	running       bool
	receiverCount int
	mu            sync.Mutex
}

// NewLogic 创建视频逻辑实例
func NewLogic(svcCtx *svc.SvcContext) *Logic {
	return &Logic{
		svcCtx:        svcCtx,
		running:       false,
		receiverCount: 0,
	}
}

// channelFrameReceiver 是一个简单的FrameReceiver实现，将帧发送到通道
type channelFrameReceiver struct {
	ch   chan []byte
	ctx  context.Context
	name string
}

// ReceiveFrame 接收视频帧并发送到通道
func (r *channelFrameReceiver) ReceiveFrame(frame []byte) error {
	select {
	case r.ch <- frame:
		return nil
	case <-r.ctx.Done():
		return r.ctx.Err()
	}
}

// Close 关闭接收器
func (r *channelFrameReceiver) Close() error {
	return nil
}

// VideoStream 启动视频流
func (l *Logic) StreamClient(ctx context.Context) (*video.Client, error) {
	clt, err := l.svcCtx.Video.StreamClient(ctx)
	if err != nil {
		return nil, err
	}
	if clt == nil {
		return nil, errors.New("video stream client is nil")
	}
	return clt, nil
}
