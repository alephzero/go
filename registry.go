package alephzero

import (
	"sync"
)

type registryDef struct {
	mutex  sync.Mutex
	objs   map[uintptr]interface{}
	nextId uintptr
}

func newGlobalCallbackRegistry() *registryDef {
	return &registryDef{
		objs: make(map[uintptr]interface{}),
	}
}

var registry = &registryDef{
	objs: make(map[uintptr]interface{}),
}

func (r *registryDef) Register(obj interface{}) (id uintptr) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	id = r.nextId
	r.nextId++
	r.objs[id] = obj
	return
}

func (r *registryDef) Unregister(id uintptr) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.objs, id)
}

func (r *registryDef) Get(id uintptr) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.objs[id]
}
