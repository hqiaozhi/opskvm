package v1

import "opskvm/internal/core/video"

// ConfigRequest API请求体结构（用于修改配置）
type ConfigRequest struct {
	Width  int `json:"width" binding:"gte=0"`  // 宽度（0表示不修改），gte=0确保非负
	Height int `json:"height" binding:"gte=0"` // 高度（0表示不修改）
	FPS    int `json:"fps" binding:"gte=0"`    // 帧率（0表示不修改）
}

// ConfigResponse API响应体结构
type ConfigResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Config  *video.Config `json:"config,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// CompressRequest 压缩控制请求体结构
type CompressRequest struct {
	Enabled bool `json:"enabled"`          // 是否开启压缩
	Quality int  `json:"quality,omitempty"` // 压缩质量 (1-100)
}

// CompressResponse 压缩控制响应体结构
type CompressResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Enabled bool   `json:"enabled"`          // 当前压缩状态
	Quality int    `json:"quality"`          // 当前压缩质量
	Error   string `json:"error,omitempty"`
}
