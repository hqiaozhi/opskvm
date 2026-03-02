package mirrors

import (
	"context"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	total, List, err := c.mirrors.SVC.MirrorsManagerService.List(req.Path, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var list []v1.Info
	for _, item := range List {
		list = append(list, v1.Info{
			Id:        item.Id,
			Name:      item.Name,
			Path:      item.Path,
			Size:      item.Size,
			URL:       item.URL,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return &v1.ListRes{
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
		List:  list,
	}, nil
}
