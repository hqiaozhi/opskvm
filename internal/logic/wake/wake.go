package wake

import "opskvm/internal/service"

type Wake struct {
	svc *service.SVC
}

func New() *Wake {
	return &Wake{
		svc: service.Svc,
	}
}
