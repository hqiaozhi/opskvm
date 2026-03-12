package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/korandiz/v4l"
)

func (c Init) cmdLine(ctx context.Context, in CIintInput) error {
	// Version.
	if in.Version {
		Version.Index(ctx, cVersionInput{})
		os.Exit(0)
		return nil
	}

	// 设置Debug模式
	// set loglevel
	g.Log().SetTimeFormat("2006-01-02 15:04:05.999")
	g.SetDebug(in.Debug)
	if in.Debug {
		g.Log().SetLevelStr("debug")
	} else {
		err := g.Log().SetLevelStr(in.LogLevel)
		if err != nil {
			return err
		}
	}

	// Device
	if in.VideoDevice {
		for _, v := range v4l.FindDevices() {
			fmt.Printf("%s %s %v %s %s %d\n", v.Path, v.BusInfo, v.Camera, v.DeviceName, v.DriverName, v.DriverVersion)
		}
		os.Exit(0)
		return nil
	}

	return nil
}
