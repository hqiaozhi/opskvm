package system

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type NetSpeed struct {
	Upload   int64
	Download int64
}

var (
	netDevPath     = "/proc/net/dev"
	previousStats  = make(map[string]NetSpeed)
	lastSampleTime time.Time
)

func (s *System) GetNetSpeed(ctx context.Context) (*NetSpeed, error) {
	file, err := os.Open(netDevPath)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "failed to open /proc/net/dev")
	}
	defer file.Close()

	currentStats := make(map[string]NetSpeed)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 17 {
			continue
		}

		iface := strings.TrimSuffix(parts[0], ":")
		if iface == "lo" || iface == "docker0" || strings.HasPrefix(iface, "veth") || strings.HasPrefix(iface, "br-") {
			continue
		}

		receiveBytes, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		transmitBytes, err := strconv.ParseInt(parts[9], 10, 64)
		if err != nil {
			continue
		}

		currentStats[iface] = NetSpeed{
			Upload:   transmitBytes,
			Download: receiveBytes,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "failed to read /proc/net/dev")
	}

	elapsed := time.Since(lastSampleTime).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}

	var totalUpload, totalDownload int64

	if !lastSampleTime.IsZero() {
		for iface, stats := range currentStats {
			prev, exists := previousStats[iface]
			if !exists {
				continue
			}

			uploadDiff := stats.Upload - prev.Upload
			if uploadDiff < 0 {
				uploadDiff = stats.Upload
			}

			downloadDiff := stats.Download - prev.Download
			if downloadDiff < 0 {
				downloadDiff = stats.Download
			}

			totalUpload += int64(float64(uploadDiff) / elapsed)
			totalDownload += int64(float64(downloadDiff) / elapsed)
		}
	}

	previousStats = currentStats
	lastSampleTime = time.Now()

	return &NetSpeed{
		Upload:   totalUpload,
		Download: totalDownload,
	}, nil
}
