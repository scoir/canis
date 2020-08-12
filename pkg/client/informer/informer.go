package informer

import "sync"

type ResourceInformer interface {
	AddEventHandler(ResourceEventHandler)
	Run(stopCh <-chan struct{})
}

type SharedResourceInformer struct {
	lock          sync.RWMutex
	eventHandlers []ResourceEventHandler
	adapter       StreamAdapter
}

func NewSharedResourceInformer(a StreamAdapter) *SharedResourceInformer {
	return &SharedResourceInformer{adapter: a}
}

func (r *SharedResourceInformer) AddEventHandler(handler ResourceEventHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.eventHandlers = append(r.eventHandlers, handler)
}

func (r *SharedResourceInformer) Run(stopCh <-chan struct{}) {
	add := r.adapter.AddCh()
	upd := r.adapter.UpdateCh()
	del := r.adapter.DeleteCh()
	for {
		select {
		case added := <-add:
			r.fireAdd(added)
		case updated := <-upd:
			r.fireUpdate(updated)
		case deleted := <-del:
			r.fireDeleted(deleted)
		case <-stopCh:
			return
		}
	}
}

func (r *SharedResourceInformer) fireAdd(added interface{}) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, handler := range r.eventHandlers {
		handler.OnAdd(added)
	}
}

func (r *SharedResourceInformer) fireUpdate(updated Update) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, handler := range r.eventHandlers {
		handler.OnUpdate(updated.Old, updated.New)
	}
}

func (r *SharedResourceInformer) fireDeleted(deleted interface{}) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, handler := range r.eventHandlers {
		handler.OnDelete(deleted)
	}
}

// ResourceEventHandler can handle notifications for events that
// happen to a resource. The events are informational only, so you
// can't return an error.  The handlers MUST NOT modify the objects
// received; this concerns not only the top level of structure but all
// the data structures reachable from it.
//  * OnAdd is called when an object is added.
//  * OnUpdate is called when an object is modified. Note that oldObj is the
//      last known state of the object-- it is possible that several changes
//      were combined together, so you can't use this to see every single
//      change. OnUpdate is also called when a re-list happens, and it will
//      get called even if nothing changed. This is useful for periodically
//      evaluating or syncing something.
//  * OnDelete will get the final state of the item if it is known, otherwise
//      it will get an object of type DeletedFinalStateUnknown. This can
//      happen if the watch is closed and misses the delete event and we don't
//      notice the deletion until the subsequent re-list.
type ResourceEventHandler interface {
	OnAdd(obj interface{})
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

// ResourceEventHandlerFuncs is an adaptor to let you easily specify as many or
// as few of the notification functions as you want while still implementing
// ResourceEventHandler.  This adapter does not remove the prohibition against
// modifying the objects.
type ResourceEventHandlerFuncs struct {
	AddFunc    func(obj interface{})
	UpdateFunc func(oldObj, newObj interface{})
	DeleteFunc func(obj interface{})
}

// OnAdd calls AddFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnAdd(obj interface{}) {
	if r.AddFunc != nil {
		r.AddFunc(obj)
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(oldObj, newObj)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnDelete(obj interface{}) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(obj)
	}
}
