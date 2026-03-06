package kvm

import "sync"

var (
	sessions     = make(map[string]bool)
	sessionsLock sync.RWMutex
)
