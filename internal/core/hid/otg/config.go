package otg

// GadgetConfig 定义USB Gadget配置
type GadgetConfig struct {
	VendorID      uint16        `json:"vendor_id"`
	ProductID     uint16        `json:"product_id"`
	USBVersion    uint16        `json:"usb_version"`
	DeviceVersion int           `json:"device_version"`
	Manufacturer  string        `json:"manufacturer"`
	Product       string        `json:"product"`
	Serial        string        `json:"serial"`
	Config        string        `json:"config"`
	MaxPower      int           `json:"max_power"`
	RemoteWakeup  bool          `json:"remote_wakeup"`
	InitDelay     int           `json:"init_delay"`
	User          string        `json:"user"`
	Meta          string        `json:"meta"`
	Endpoints     int           `json:"endpoints"`
	UDC           string        `json:"udc"`
	Gadget        string        `json:"gadget"`
	Devices       DevicesConfig `json:"devices"`
	KVMD          KVMDConfig    `json:"kvmd"`
	SysFSPrefix   string        `json:"sysfs_prefix"`
}

// DevicesConfig 定义设备配置
type DevicesConfig struct {
	HID      HIDConfig      `json:"hid"`
	MSD      MSDConfig      `json:"msd"`
	Drives   DrivesConfig   `json:"drives"`
	Ethernet EthernetConfig `json:"ethernet"`
	Serial   SerialConfig   `json:"serial"`
	Audio    AudioConfig    `json:"audio"`
}

// HIDConfig 定义HID设备配置
type HIDConfig struct {
	Keyboard HIDDeviceConfig `json:"keyboard"`
	Mouse    HIDDeviceConfig `json:"mouse"`
	MouseAlt HIDDeviceConfig `json:"mouse_alt"`
}

// HIDDeviceConfig 定义单个HID设备配置
type HIDDeviceConfig struct {
	Start bool `json:"start"`
}

// MSDConfig 定义MSD设备配置
type MSDConfig struct {
	Start   bool             `json:"start"`
	Default MSDDefaultConfig `json:"default"`
}

// MSDDefaultConfig 定义MSD默认配置
type MSDDefaultConfig struct {
	Stall         bool                `json:"stall"`
	CDROM         bool                `json:"cdrom"`
	RW            bool                `json:"rw"`
	Removable     bool                `json:"removable"`
	FUA           bool                `json:"fua"`
	InquiryString InquiryStringConfig `json:"inquiry_string"`
}

// InquiryStringConfig 定义查询字符串配置
type InquiryStringConfig struct {
	CDROM InquiryStringDeviceConfig `json:"cdrom"`
	Flash InquiryStringDeviceConfig `json:"flash"`
}

// InquiryStringDeviceConfig 定义设备查询字符串配置
type InquiryStringDeviceConfig struct {
	Vendor   string `json:"vendor"`
	Product  string `json:"product"`
	Revision string `json:"revision"`
}

// DrivesConfig 定义驱动器配置
type DrivesConfig struct {
	Enabled bool `json:"enabled"`
	Count   int  `json:"count"`
	Start   bool `json:"start"`
}

// EthernetConfig 定义以太网配置
type EthernetConfig struct {
	Enabled bool   `json:"enabled"`
	Driver  string `json:"driver"`
	HostMAC string `json:"host_mac"`
	KVMMAC  string `json:"kvm_mac"`
	Start   bool   `json:"start"`
}

// SerialConfig 定义串口配置
type SerialConfig struct {
	Enabled bool `json:"enabled"`
	Start   bool `json:"start"`
}

// AudioConfig 定义音频配置
type AudioConfig struct {
	Enabled bool `json:"enabled"`
	Start   bool `json:"start"`
}

// KVMDConfig 定义KVMD配置
type KVMDConfig struct {
	HID HIDKVMDConfig `json:"hid"`
	MSD MSDKVMDConfig `json:"msd"`
}

// HIDKVMDConfig 定义HID KVMD配置
type HIDKVMDConfig struct {
	Type     string         `json:"type"`
	Mouse    MouseConfig    `json:"mouse"`
	MouseAlt MouseAltConfig `json:"mouse_alt"`
}

// MouseConfig 定义鼠标配置
type MouseConfig struct {
	Absolute        bool `json:"absolute"`
	HorizontalWheel bool `json:"horizontal_wheel"`
}

// MouseAltConfig 定义备用鼠标配置
type MouseAltConfig struct {
	Device string `json:"device"`
}

// MSDKVMDConfig 定义MSD KVMD配置
type MSDKVMDConfig struct {
	Type string `json:"type"`
}

// Hid 定义HID设备结构
type Hid struct {
	Protocol         int    `json:"protocol"`
	Subclass         int    `json:"subclass"`
	ReportLength     int    `json:"report_length"`
	ReportDescriptor []byte `json:"report_descriptor"`
}
