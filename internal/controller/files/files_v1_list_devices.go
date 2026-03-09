package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) ListDevices(ctx context.Context, req *v1.ListDevicesReq) (res *v1.ListDevicesRes, err error) {
	devices, err := c.Files.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	var deviceList []v1.DeviceInfo
	for _, d := range devices {
		deviceList = append(deviceList, v1.DeviceInfo{
			DevPath:     d.DevPath,
			MountPoint:  d.MountPoint,
			Name:        d.Name,
			Model:       d.Model,
			Capacity:    d.Capacity,
			Partition:   d.Partition,
			FsType:      d.FsType,
			Size:        d.Size,
			Used:        d.Used,
			Available:   d.Available,
			UsedPercent: d.UsedPercent,
			IsPartition: d.IsPartition,
			IsMounted:   d.IsMounted,
		})
	}

	return &v1.ListDevicesRes{
		Devices: deviceList,
	}, nil
}
