package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/gogf/gf/v2/frame/g"
)

var Install = install{}

type install struct {
	g.Meta `name:"install" brief:"install to system (need root permission)"`
}

type installInput struct {
	g.Meta `name:"install"`
}

type InstallOutput struct{}

func (i install) Index(ctx context.Context, in installInput) (out *InstallOutput, err error) {

	// 检查是否以root权限运行
	if os.Geteuid() != 0 {
		g.Log().Error(ctx, "Please run as root (sudo)")
		os.Exit(1)
	}

	// 获取当前可执行文件路径
	currentExecPath, err := os.Executable()
	if err != nil {
		g.Log().Error(ctx, "Failed to get executable path: ", err)
		os.Exit(1)
	}

	// 服务文件路径
	servicePath := "/lib/systemd/system/opskvm.service"
	if _, err := os.Stat(servicePath); err == nil {
		g.Log().Info(ctx, "stop old opskvm service...")
		if err := exec.Command("systemctl", "stop", "opskvm.service").Run(); err != nil {
			g.Log().Warning(ctx, "Failed to stop opskvm service (might not be running): ", err)
		}
		os.RemoveAll(servicePath)
	}

	// 目标安装路径
	targetExecPath := "/usr/local/bin/opskvm"
	if _, err := os.Stat(targetExecPath); err == nil {
		os.RemoveAll(targetExecPath)
	}
	// 复制可执行文件到目标路径
	g.Log().Infof(ctx, "Copying executable to %s...", targetExecPath)
	currentContent, err := os.ReadFile(currentExecPath)
	if err != nil {
		g.Log().Error(ctx, "Failed to read current executable: ", err)
		os.Exit(1)
	}
	if err := os.WriteFile(targetExecPath, currentContent, 0755); err != nil {
		g.Log().Error(ctx, "Failed to copy executable to ", targetExecPath, ": ", err)
		os.Exit(1)
	}

	// 构建systemd服务单元文件内容
	serviceContent := []*unit.UnitOption{
		{Section: "Unit", Name: "Description", Value: "opskvm service"},
		{Section: "Unit", Name: "After", Value: "network.target"},
		{Section: "Service", Name: "Type", Value: "simple"},
		{Section: "Service", Name: "ExecStart", Value: fmt.Sprintf("%s --hidmode %s", targetExecPath, detectHIDMode(ctx))},
		{Section: "Service", Name: "Restart", Value: "on-failure"},
		{Section: "Service", Name: "RestartSec", Value: "5"},
		{Section: "Service", Name: "User", Value: "root"},
		{Section: "Install", Name: "WantedBy", Value: "multi-user.target"},
	}

	// 转换为单元文件字符串
	serviceStr := ""
	for _, opt := range serviceContent {
		serviceStr += fmt.Sprintf("[%s]\n%s=%s\n", opt.Section, opt.Name, opt.Value)
	}

	// 写入服务文件
	if err := os.WriteFile(servicePath, []byte(serviceStr), 0644); err != nil {
		g.Log().Error(ctx, "Failed to write systemd service file: ", err)
		os.Exit(1)
	}
	g.Log().Info(ctx, "Systemd service file created at: ", servicePath)

	// 重新加载systemd守护进程
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		g.Log().Error(ctx, "Failed to reload systemd daemon: ", err)
		os.Exit(1)
	}
	g.Log().Info(ctx, "Systemd daemon reloaded")

	// 启用服务
	if err := exec.Command("systemctl", "enable", "opskvm.service").Run(); err != nil {
		g.Log().Error(ctx, "Failed to enable opskvm service: ", err)
		os.Exit(1)
	}
	g.Log().Info(ctx, "opskvm service enabled")

	// 启动服务
	if err := exec.Command("systemctl", "start", "opskvm.service").Run(); err != nil {
		g.Log().Error(ctx, "Failed to start opskvm service: ", err)
		os.Exit(1)
	}
	g.Log().Info(ctx, "opskvm service started successfully")

	// 输出服务状态
	statusCmd := exec.Command("systemctl", "status", "opskvm.service", "--no-pager")
	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr
	if err := statusCmd.Run(); err != nil {
		g.Log().Warning(ctx, "Failed to get service status: ", err)
	}
	return
}

func detectHIDMode(ctx context.Context) string {
	otgExists := false
	ch9329Exists := false

	udcList, err := filepath.Glob("/sys/class/udc/*")
	if err == nil && len(udcList) > 0 {
		otgExists = true
	}

	ch9329Matches, err := filepath.Glob("/dev/ttyUSB*")
	if err == nil && len(ch9329Matches) > 0 {
		ch9329Exists = true
	}

	if otgExists || ch9329Exists {
		if !otgExists && ch9329Exists {
			g.Log().Info(ctx, "Only CH9329 device found, using --hidmode ch9329")
			return "ch9329"
		}
		g.Log().Info(ctx, "OTG device found or both devices exist, using --hidmode otg")
		return "otg"
	}

	g.Log().Warning(ctx, "No HID device found, defaulting to --hidmode otg")
	return "otg"
}
