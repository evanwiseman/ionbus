package handlers

import (
	"sync"

	"github.com/evanwiseman/ionbus/internal/models"
)

type Registry struct {
	handlers map[string]*models.Handler
	mu       sync.RWMutex
}

func (r *Registry) Add(handler *models.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.Method] = handler
}

func (r *Registry) Remove(handler *models.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, handler.Method)
}

func (r *Registry) Get(method string) (*models.Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[method]
	return h, ok
}
