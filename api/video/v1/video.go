package v1

// VideoRequest 视频请求参数
type VideoRequest struct {
	Width     int    `json:"Width"`     // 分辨率宽度
	Height    int    `json:"Height"`    // 分辨率高度
	Framerate int    `json:"Framerate"` // 帧率
	Device    string `json:"Device"`    // 视频设备路径
	Action    string `json:"Action"`    // 操作类型 (start/stop)
}

// VideoResponse 视频响应参数
type VideoResponse struct {
	FrameData string `json:"FrameData"` // 视频帧数据 (Base64编码)
	Width     int    `json:"Width"`     // 实际宽度
	Height    int    `json:"Height"`    // 实际高度
	Error     string `json:"Error"`     // 错误信息
}

// VideoConfigRequest 视频配置请求参数
type VideoConfigRequest struct {
	Width     int    `json:"Width"`     // 分辨率宽度
	Height    int    `json:"Height"`    // 分辨率高度
	Framerate int    `json:"Framerate"` // 帧率
	Device    string `json:"Device"`    // 视频设备路径
}

// VideoConfigResponse 视频配置响应参数
type VideoConfigResponse struct {
	Width     int    `json:"Width"`     // 实际宽度
	Height    int    `json:"Height"`    // 实际高度
	Framerate int    `json:"Framerate"` // 实际帧率
	Device    string `json:"Device"`    // 当前设备路径
}

// DeviceListResponse 设备列表响应参数
type DeviceListResponse struct {
	Devices []string `json:"Devices"` // 设备路径列表
}

// SupportedConfigsResponse 支持的配置响应参数
type SupportedConfigsResponse struct {
	Configs []ConfigOption `json:"Configs"` // 支持的配置列表
}

// ConfigOption 配置选项
type ConfigOption struct {
	Width     int `json:"Width"`     // 宽度
	Height    int `json:"Height"`    // 高度
	Framerate int `json:"Framerate"` // 帧率
}
