package container

import (
	"fmt"
	"sync"
)

type Container struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

func New() *Container {
	return &Container{
		services: make(map[string]interface{}),
	}
}

func (c *Container) Register(key string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.services[key]; exists {
		panic(fmt.Sprintf("Сервис с ключом '%s' уже зарегистрирован", key))
	}
	c.services[key] = service
}

func (c *Container) Resolve(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[key]
	if !exists {
		panic(fmt.Sprintf("Сервис с ключом '%s' не найден", key))
	}
	return service
}
