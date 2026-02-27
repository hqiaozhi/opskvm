package consts

// WSMessageType WebSocket消息类型
type WSMessageType int

const (
	WSMessageTypeVideo     WSMessageType = 0 // 视频流数据
	WSMessageTypeKeyboard  WSMessageType = 1 // 键盘事件
	WSMessageTypeMouse     WSMessageType = 2 // 鼠标事件
	WSMessageTypeMouseMode WSMessageType = 3 // 鼠标模式切换
)
