package system

import "opskvm/internal/service"

type System struct {
	SVC *service.SVC
}

func New() *System {
	return &System{
		SVC: service.Svc,
	}
}
