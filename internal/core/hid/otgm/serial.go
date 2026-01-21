package otgm

type SerialInterface interface {
	// AddMSD() error
	// Bind(absoltePath, cdrom string) error
	// Remove() error
}

type Serial struct {
	GadgetInterface
	hidInstance int // 命名规范要求使用数字来索引HID实例
	funcName    string
	funcPath    string
	cdrom       string
	ro          string
}

func NewSerial(gadget GadgetInterface) SerialInterface {
	return &Serial{
		GadgetInterface: gadget,
	}
}
