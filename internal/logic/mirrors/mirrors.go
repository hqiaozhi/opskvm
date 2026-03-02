package mirrors

import (
	"opskvm/internal/service"
)

type Mirrors struct {
	SVC *service.SVC
}

func New() *Mirrors {
	return &Mirrors{
		SVC: service.Svc,
	}
}
