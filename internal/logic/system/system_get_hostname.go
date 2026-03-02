package system

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (s *System) GetHostname(ctx context.Context) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "failed to get hostname")
	}
	return hostname, nil
}
