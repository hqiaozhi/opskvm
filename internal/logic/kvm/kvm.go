package kvm

import "opskvm/internal/service"

type KVM struct {
	SVC *service.SVC
}

func New() *KVM {
	return &KVM{
		SVC: service.Svc,
	}
}
