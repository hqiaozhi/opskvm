package files

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type DeviceInfo struct {
	DevPath     string `json:"dev_path"`
	MountPoint  string `json:"mount_point"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	Capacity    string `json:"capacity"`
	Partition   string `json:"partition"`
	FsType      string `json:"fs_type"`
	Size        int64  `json:"size"`
	Used        int64  `json:"used"`
	Available   int64  `json:"available"`
	UsedPercent int    `json:"used_percent"`
	IsPartition bool   `json:"is_partition"`
	IsMounted   bool   `json:"is_mounted"`
}

type LsblkOutput struct {
	Blockdevices []LsblkDevice `json:"blockdevices"`
}

type LsblkDevice struct {
	Alignment    int           `json:"alignment"`
	DiscAln      int           `json:"disc-aln"`
	Dax          bool          `json:"dax"`
	DiscGran     string        `json:"disc-gran"`
	DiscMax      string        `json:"disc-max"`
	DiscZero     bool          `json:"disc-zero"`
	Fsavail      string        `json:"fsavail"`
	Fsroots      []string      `json:"fsroots"`
	Fssize       string        `json:"fssize"`
	Fstype       string        `json:"fstype"`
	Fsused       string        `json:"fsused"`
	Fsuse        string        `json:"fsuse%"`
	Fsver        string        `json:"fsver"`
	Group        string        `json:"group"`
	Hctl         string        `json:"hctl"`
	Hotplug      bool          `json:"hotplug"`
	Kname        string        `json:"kname"`
	Label        string        `json:"label"`
	LogSec       int           `json:"log-sec"`
	MajMin       string        `json:"maj:min"`
	MinIo        int           `json:"min-io"`
	Mode         string        `json:"mode"`
	Model        string        `json:"model"`
	Name         string        `json:"name"`
	OptIo        int           `json:"opt-io"`
	Owner        string        `json:"owner"`
	Partflags    string        `json:"partflags"`
	Partlabel    string        `json:"partlabel"`
	Parttype     string        `json:"parttype"`
	Parttypename string        `json:"parttypename"`
	Partuuid     string        `json:"partuuid"`
	Path         string        `json:"path"`
	PhySec       int           `json:"phy-sec"`
	Pkname       string        `json:"pkname"`
	Pttype       string        `json:"pttype"`
	Ptuuid       string        `json:"ptuuid"`
	Ra           int           `json:"ra"`
	Rand         bool          `json:"rand"`
	Rev          string        `json:"rev"`
	Rm           bool          `json:"rm"`
	Ro           bool          `json:"ro"`
	Rota         bool          `json:"rota"`
	RqSize       int           `json:"rq-size"`
	Sched        string        `json:"sched"`
	Serial       string        `json:"serial"`
	Size         string        `json:"size"`
	Start        int           `json:"start"`
	State        string        `json:"state"`
	Subsystems   string        `json:"subsystems"`
	Mountpoint   string        `json:"mountpoint"`
	Mountpoints  []string      `json:"mountpoints"`
	Tran         string        `json:"tran"`
	Type         string        `json:"type"`
	Uuid         string        `json:"uuid"`
	Vendor       string        `json:"vendor"`
	Wsame        string        `json:"wsame"`
	Wwn          string        `json:"wwn"`
	Zoned        string        `json:"zoned"`
	ZoneSz       string        `json:"zone-sz"`
	ZoneWgran    string        `json:"zone-wgran"`
	ZoneApp      string        `json:"zone-app"`
	ZoneNr       int           `json:"zone-nr"`
	ZoneOmax     int           `json:"zone-omax"`
	ZoneAmax     int           `json:"zone-amax"`
	Children     []LsblkDevice `json:"children"`
}

func (d *FILES) ListDevices(ctx context.Context) ([]DeviceInfo, error) {
	cmd := exec.Command("lsblk", "-J", "-O")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行lsblk失败: %v, output: %s", err, string(output))
	}

	g.Log().Debugf(ctx, "ListDevices: lsblk output length: %d", len(output))

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析lsblk JSON失败: %v", err)
	}

	var devices []DeviceInfo
	for _, dev := range result.Blockdevices {
		deviceInfo := parseLsblkDevice(dev)
		if deviceInfo == nil {
			continue
		}
		devices = append(devices, *deviceInfo)

		for _, child := range dev.Children {
			childInfo := parseLsblkDevice(child)
			if childInfo == nil {
				continue
			}
			devices = append(devices, *childInfo)
		}
	}

	return devices, nil
}

func parseLsblkDevice(dev LsblkDevice) *DeviceInfo {
	if dev.Type == "disk" {
		if strings.HasPrefix(dev.Name, "mmcblk2boot") ||
			strings.HasPrefix(dev.Name, "mmcblk2rpmb") ||
			strings.HasPrefix(dev.Name, "zram") {
			return nil
		}
	}

	mountpoint := dev.Mountpoint
	if mountpoint == "/" || mountpoint == "/boot" {
		return nil
	}

	isPartition := dev.Type == "part"

	return &DeviceInfo{
		DevPath:     dev.Path,
		MountPoint:  mountpoint,
		Name:        dev.Name,
		Model:       dev.Model,
		Capacity:    dev.Size,
		Partition:   extractPartitionNumber(dev.Name),
		FsType:      dev.Fstype,
		IsPartition: isPartition,
		IsMounted:   mountpoint != "" && mountpoint != "[SWAP]",
	}
}

func extractPartitionNumber(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] >= '0' && name[i] <= '9' {
			return name[i:]
		}
	}
	return ""
}
