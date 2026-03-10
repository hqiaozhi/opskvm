package otg

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	gLog "github.com/gogf/gf/v2/frame/g"
)

var (
	UDCWatcherEnabled   bool
	UDCWatcherStopCh    chan struct{}
	UDCWatcherRestartCh chan struct{}
)

// WaitForUDC 等待 UDC 设备插入
// maxWaitTime: 最大等待时间，0 表示无限等待
// checkInterval: 检查间隔
func WaitForUDC(maxWaitTime time.Duration, checkInterval time.Duration) (string, error) {
	startTime := time.Now()

	for {
		udcList, err := os.ReadDir("/sys/class/udc")
		if err != nil {
			return "", fmt.Errorf("Gadget: Failed to read UDC list: %w", err)
		}

		if len(udcList) > 0 {
			udcNames := make([]string, len(udcList))
			for i, udc := range udcList {
				udcNames[i] = udc.Name()
			}
			sort.Strings(udcNames)
			gLog.Log().Infof(context.Background(), "Found UDC device: %s", udcNames[0])
			return udcNames[0], nil
		}

		if maxWaitTime > 0 && time.Since(startTime) >= maxWaitTime {
			return "", fmt.Errorf("Gadget: Timeout waiting for UDC device")
		}

		gLog.Log().Infof(context.Background(), "Waiting for UDC device to be inserted...")
		time.Sleep(checkInterval)
	}
}

// CheckUDCExists 检查 UDC 设备是否存在
func CheckUDCExists(udcName string) bool {
	udcPath := fmt.Sprintf("/sys/class/udc/%s", udcName)
	_, err := os.Stat(udcPath)
	return err == nil
}

// StartUDCWatcher 启动 UDC 设备状态监控器
// 当设备移除时自动触发回调
func StartUDCWatcher(udcName string, onLost func(), onReconnect func()) {
	UDCWatcherEnabled = true
	UDCWatcherStopCh = make(chan struct{})
	UDCWatcherRestartCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-UDCWatcherStopCh:
				gLog.Log().Info(context.Background(), "UDC watcher stopped")
				return
			case <-UDCWatcherRestartCh:
				gLog.Log().Info(context.Background(), "UDC watcher restarting...")
				continue
			default:
				if !CheckUDCExists(udcName) {
					gLog.Log().Warningf(context.Background(), "UDC device %s lost!", udcName)
					if onLost != nil {
						onLost()
					}

					for {
						select {
						case <-UDCWatcherStopCh:
							return
						default:
							if CheckUDCExists(udcName) {
								gLog.Log().Infof(context.Background(), "UDC device %s reconnected!", udcName)
								if onReconnect != nil {
									onReconnect()
								}
								goto WaitNext
							}
							time.Sleep(1 * time.Second)
						}
					}
				WaitNext:
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	gLog.Log().Infof(context.Background(), "UDC watcher started for device: %s", udcName)
}

// StopUDCWatcher 停止 UDC 设备监控器
func StopUDCWatcher() {
	if UDCWatcherEnabled {
		close(UDCWatcherStopCh)
		UDCWatcherEnabled = false
	}
}

// FindUDC 查找可用的UDC控制器
func FindUDC() (string, error) {
	udcList, err := os.ReadDir("/sys/class/udc")
	if err != nil {
		return "", fmt.Errorf("Gadget: Failed to read UDC list: %w", err)
	}

	if len(udcList) == 0 {
		return "", fmt.Errorf("Gadget: No UDC devices found")
	}

	udcNames := make([]string, len(udcList))
	for i, udc := range udcList {
		udcNames[i] = udc.Name()
	}
	sort.Strings(udcNames)

	return udcNames[0], nil
}
