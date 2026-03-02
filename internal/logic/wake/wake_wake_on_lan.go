package wake

import (
	"context"
)

func (w *Wake) WakeOnLan(ctx context.Context, broadcastIp string, port int, macAddr string) error {
	return w.svc.WOL.WakeUp(broadcastIp, port, macAddr)
}
