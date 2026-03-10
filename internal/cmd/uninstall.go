package cmd

import (
	"context"
	"os"
	"os/exec"

	"github.com/gogf/gf/v2/frame/g"
)

var Uninstall = uninstall{}

type uninstall struct {
	g.Meta `name:"uninstall" brief:"uninstall from system (need root permission)"`
}

type UninstallInput struct {
	g.Meta `name:"uninstall"`
}
type UninstallOutput struct{}

func (i uninstall) Index(ctx context.Context, in UninstallInput) (out *UninstallOutput, err error) {

	// 检查是否以root权限运行
	if os.Geteuid() != 0 {
		g.Log().Error(ctx, "Please run as root (sudo)")
		os.Exit(1)
	}

	// 服务文件路径
	servicePath := "/lib/systemd/system/opskvm.service"

	// 停止服务（如果正在运行）
	g.Log().Info(ctx, "Stopping opskvm service...")
	if err := exec.Command("systemctl", "stop", "opskvm.service").Run(); err != nil {
		g.Log().Warning(ctx, "Failed to stop opskvm service (might not be running): ", err)
	}

	// 禁用服务
	g.Log().Info(ctx, "Disabling opskvm service...")
	if err := exec.Command("systemctl", "disable", "opskvm.service").Run(); err != nil {
		g.Log().Warning(ctx, "Failed to disable opskvm service (might not be enabled): ", err)
	}

	// 删除服务文件
	g.Log().Info(ctx, "Removing systemd service file...")
	if err := os.Remove(servicePath); err != nil {
		if !os.IsNotExist(err) {
			g.Log().Error(ctx, "Failed to remove systemd service file: ", err)
			os.Exit(1)
		}
	}

	// 重新加载systemd守护进程
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		g.Log().Error(ctx, "Failed to reload systemd daemon: ", err)
		os.Exit(1)
	}

	// 删除可执行文件
	g.Log().Info(ctx, "Removing executable file...")
	execPath := "/usr/local/bin/opskvm"
	if err := os.Remove(execPath); err != nil {
		if !os.IsNotExist(err) {
			g.Log().Error(ctx, "Failed to remove executable ", execPath, ": ", err)
			os.Exit(1)
		}
	}

	// 重置systemd失效状态
	if err := exec.Command("systemctl", "reset-failed").Run(); err != nil {
		g.Log().Warning(ctx, "Failed to reset failed systemd services: ", err)
	}

	g.Log().Info(ctx, "opskvm service uninstalled successfully")
	return
}
