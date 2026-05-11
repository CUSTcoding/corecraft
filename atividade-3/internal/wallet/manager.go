package wallet

import (
	"errors"
	"sync"
)



type Manager struct {
	mu       sync.RWMutex
	active   string
}


var ErrNoWallet = errors.New("nenhuma wallet selecionada")


func New(initial string) *Manager {
	return &Manager{active: initial}
}


func (m *Manager) Active() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.active == "" {
		return "", ErrNoWallet
	}
	return m.active, nil
}


func (m *Manager) Set(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = name
}
