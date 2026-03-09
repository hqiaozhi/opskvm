package files

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

func (d *FILES) MountDevice(ctx context.Context, devPath string) (string, error) {
	mountPoint := filepath.Join(d.SVC.RootPath, "uploads")

	g.Log().Debugf(ctx, "MountDevice: devPath=%s, mountPoint=%s", devPath, mountPoint)

	if devPath == "" {
		if err := d.unmountTarget(ctx, mountPoint); err != nil {
			return "", nil
		}
		return mountPoint, nil
	}

	if !strings.HasPrefix(devPath, "/dev/") {
		g.Log().Warningf(ctx, "MountDevice: invalid devPath, must start with /dev/")
		return "", nil
	}

	cmd := exec.Command("mount", devPath, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		g.Log().Errorf(ctx, "files mount failed: %s, output: %s", err.Error(), string(output))
		return "", err
	}

	g.Log().Infof(ctx, "[MountDevice] mounted %s to %s", devPath, mountPoint)
	return mountPoint, nil
}

func (d *FILES) unmountTarget(ctx context.Context, targetPath string) error {
	cmd := exec.Command("umount", targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		g.Log().Errorf(ctx, "files unmount failed: %s, output: %s", err.Error(), string(output))
		return err
	}
	g.Log().Infof(ctx, "[unmountTarget] unmounted %s", targetPath)
	return nil
}
