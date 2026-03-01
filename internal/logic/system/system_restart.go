package system

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (s *System) RestartSys(ctx context.Context) (err error) {
	var restartCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		restartCmd = exec.Command("bash", "-c", "systemctl reboot || /sbin/reboot")
	case "windows":
		restartCmd = exec.Command("shutdown", "/r", "/t", "0")
	case "darwin":
		restartCmd = exec.Command("osascript", "-e", `tell app "System Events" to restart`)
	default:
		return gerror.NewCode(gcode.CodeNotImplemented, "unsupported operating system: "+runtime.GOOS)
	}

	restartCmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	restartCmd.Stdout = os.Stdout
	restartCmd.Stderr = os.Stderr

	if err = restartCmd.Start(); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "failed to start restart command")
	}

	go func() {
		restartCmd.Wait()
		os.Exit(0)
	}()

	return nil
}
