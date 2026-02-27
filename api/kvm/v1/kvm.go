package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config 简化的配置结构体（API返回用）
type Config struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	FPS    float64 `json:"fps"`
}

// StreamReq WebSocket 流请求
type StreamReq struct {
	g.Meta `path:"video" method:"get" sm:"视频流" tags:"视频"`
}

// StreamRes WebSocket 流响应
type StreamRes struct {
}

// TurnOnReq 开启摄像头请求
type TurnOnReq struct {
	g.Meta `path:"video/on" method:"post" sm:"开启摄像头" tags:"视频"`
}

// TurnOnRes 开启摄像头响应
type TurnOnRes struct {
}

// TurnOffReq 关闭摄像头请求
type TurnOffReq struct {
	g.Meta `path:"video/off" method:"post" sm:"关闭摄像头" tags:"视频"`
}

// TurnOffRes 关闭摄像头响应
type TurnOffRes struct {
}

// GetConfigReq 获取配置请求
type GetConfigReq struct {
	g.Meta `path:"video/config" method:"get" sm:"当前配置" tags:"视频"`
}

// GetConfigRes 获取配置响应
type GetConfigRes struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	FPS    float64 `json:"fps"`
}

// GetConfigsReq 获取支持的配置列表请求
type GetConfigsReq struct {
	g.Meta `path:"video/configs" method:"get" sm:"支持配置" tags:"视频"`
}

// GetConfigsRes 获取支持的配置列表响应
type GetConfigsRes struct {
	Configs []Config `json:"configs,omitempty" dc:"支持的配置列表"`
}

// UpdateConfigReq 更新配置请求
type UpdateConfigReq struct {
	g.Meta `path:"video/config" method:"post" sm:"更新配置" tags:"视频"`
	Width  uint16  `json:"width" v:"required" dc:"宽度"`  // 宽度
	Height uint16  `json:"height" v:"required" dc:"高度"` // 高度
	FPS    float64 `json:"fps" v:"required" dc:"帧率"`    // 帧率
}

type UpdateConfigRes struct {
	Width  int    `json:"width" v:"required" dc:"宽度"`  // 宽度
	Height int    `json:"height" v:"required" dc:"高度"` // 高度
	FPS    uint32 `json:"fps" v:"required" dc:"帧率"`    // 帧率
}

// 压缩管理
type GetCompressReq struct {
	g.Meta `path:"video/compress" method:"get" sm:"压缩状态" tags:"视频"`
}
type GetCompressRes struct {
	Enabled bool `json:"enabled"` // 当前压缩状态
	Quality int  `json:"quality"` // 当前压缩质量
}

type SetCompressReq struct {
	g.Meta  `path:"video/compress" method:"post" sm:"设置压缩" tags:"视频"`
	Enabled bool `json:"enabled"` // 当前压缩状态
	Quality int  `json:"quality"` // 当前压缩质量
}

type SetCompressRes struct {
	Enabled bool `json:"enabled"` // 当前压缩状态
	Quality int  `json:"quality"` // 当前压缩质量
}
