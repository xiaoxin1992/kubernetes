package custom

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type Informer struct {
	listAndWatch cache.ListerWatcher
	objType      runtime.Object
	handler      cache.ResourceEventHandler

	reflector *cache.Reflector
	fifo      *cache.DeltaFIFO
	store     cache.Store
}

func NewInformer(listAndWatch *cache.ListWatch, objType runtime.Object, handler cache.ResourceEventHandler) *Informer {
	// listAndWatch List 和 Watch 机制的实现
	// objType 资源对象， 比如&v1.Pod{}
	// handler 当前资源的处理对象，需要实现, onAdd, onUpdate, onDelete函数
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction:  cache.MetaNamespaceKeyFunc,
		KnownObjects: store,
	})

	reflector := cache.NewReflector(listAndWatch, objType, fifo, 0)

	return &Informer{
		listAndWatch: listAndWatch,
		objType:      objType,
		handler:      handler,
		reflector:    reflector,
		fifo:         fifo,
		store:        store,
	}
}

func (i *Informer) Start(ch chan struct{}) {
	go func() {
		i.reflector.Run(ch)
	}()
	for {
		i.fifo.Pop(func(obj interface{}, isInInitialList bool) error {
			var err error
			for _, delta := range obj.(cache.Deltas) {
				switch delta.Type {
				case cache.Sync, cache.Added:
					err = i.store.Add(delta.Object)
					if err != nil {
						fmt.Println("error handling add event: ", err)
					}
					i.handler.OnAdd(delta.Object, isInInitialList)
				case cache.Updated:
					old, exists, err := i.store.Get(delta.Object)
					if err == nil && exists {
						err = i.store.Update(delta.Object)
						if err != nil {
							fmt.Println("Error updating cache: ", err)
						}
						i.handler.OnUpdate(old, delta.Object)
					}
				case cache.Deleted:
					err = i.store.Delete(delta.Object)
					if err != nil {
						fmt.Println("Error deleting cache: ", err)
					}
					i.handler.OnDelete(delta.Object)
				}
			}
			return err
		})
	}
}
