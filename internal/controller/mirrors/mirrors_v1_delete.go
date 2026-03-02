package mirrors

import (
	"context"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	ids := req.Ids

	deletedCount, failedIds, err := c.mirrors.SVC.MirrorsManagerService.Delete(ids)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteRes{
		DeletedCount: deletedCount,
		FailedIds:    failedIds,
	}, nil
}
