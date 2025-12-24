/*
* 视频管理器接口定义了视频设备的操作方法。
* 包括获取视频设备路径、开启设备、获取设备配置、设置设备配置、视频流处理和关闭设备。
* 设备修改配置的时候需要关闭视频流，否则会导致视频流异常。
 */

package video

import (
	"context"
)

type VideoManager interface {
	// GetPath 获取所有视频设备路径
	GetPath() []string
	// 开启设备
	Open() error
	// ListConfigs 获取当前设备支持的配置
	ListConfigs() (map[string][]deviceSupportConfig, error)
	// SetConfig 设置设备配置
	SetConfig() error
	// StreamClient 获取视频流客户端
	StreamClient(ctx context.Context) (*Client, error)
	// 关闭设备
	Close() error
}
