package system

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shirou/gopsutil/v4/cpu"
)

type SystemInfo struct {
	Version  string
	Hostname string
	Distro   string
	Kernel   string
	Arch     string
	HostIP   string
	BootTime string
	Uptime   string
	CpuModel string
}

func (s *System) SysInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{}

	info.Version = s.SVC.Version

	hostname, err := os.Hostname()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "failed to get hostname")
	}
	info.Hostname = hostname

	distro, err := getDistro()
	if err == nil {
		info.Distro = distro
	}

	kernel, err := exec.Command("uname", "-r").Output()
	if err == nil {
		info.Kernel = strings.TrimSpace(string(kernel))
	}

	info.Arch = runtime.GOARCH

	hostIP, err := getHostIP()
	if err == nil {
		info.HostIP = hostIP
	}

	bootTime, uptime, err := getBootTimeAndUptime()
	if err == nil {
		info.BootTime = bootTime
		info.Uptime = uptime
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CpuModel = cpuInfo[0].ModelName
	}

	return info, nil
}

func getDistro() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}

	var prettyName string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = strings.TrimPrefix(line, "PRETTY_NAME=")
			prettyName = strings.Trim(prettyName, `"`)
			return prettyName, nil
		}
	}
	return "", io.EOF
}

func getHostIP() (string, error) {
	cmd := exec.Command("bash", "-c", "hostname -I | awk '{print $1}'")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getBootTimeAndUptime() (string, string, error) {
	uptimeBytes, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", "", err
	}

	fields := strings.Fields(string(uptimeBytes))
	if len(fields) < 1 {
		return "", "", gerror.New("invalid uptime file")
	}

	var uptimeSeconds float64
	_, err = fmt.Sscanf(fields[0], "%f", &uptimeSeconds)
	if err != nil {
		return "", "", err
	}

	uptimeDuration := time.Duration(uptimeSeconds) * time.Second
	uptimeStr := formatUptime(uptimeDuration)

	bootTime := time.Now().Add(-uptimeDuration)
	bootTimeStr := bootTime.Format("2006-01-02 15:04:05")

	return bootTimeStr, uptimeStr, nil
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	parts = append(parts, fmt.Sprintf("%ds", seconds))

	return strings.Join(parts, " ")
}
