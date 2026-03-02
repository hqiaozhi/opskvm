package wake

import (
	"context"
	"opskvm/api/wake/v1"
	"strconv"
	"strings"
)

func (c *ControllerV1) DeleteWolBatch(ctx context.Context, req *v1.DeleteWolBatchReq) (res *v1.DeleteWolBatchRes, err error) {
	idsStr := strings.Split(req.Ids, ",")
	ids := make([]int, 0, len(idsStr))
	for _, idStr := range idsStr {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	deletedCount, err := c.Wake.DeleteWolBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteWolBatchRes{
		DeletedCount: deletedCount,
	}, nil
}
