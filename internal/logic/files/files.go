package files

import "opskvm/internal/service"

type FILES struct {
	SVC *service.SVC
}

func New() *FILES {
	return &FILES{
		SVC: service.Svc,
	}
}
