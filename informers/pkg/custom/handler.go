package custom

import (
	"fmt"
	"k8s.io/client-go/tools/cache"
)

// 给informer回调使用

var _ cache.ResourceEventHandler = &PodHandler{}

type PodHandler struct{}

func NewPodHandler() *PodHandler {
	return &PodHandler{}
}

func (h *PodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	fmt.Println("OnAdd pod handler", obj)
}
func (h *PodHandler) OnUpdate(oldObj, newObj interface{}) {
	fmt.Println("OnUpdate pod handler", oldObj, newObj)
}

func (h *PodHandler) OnDelete(obj interface{}) {
	fmt.Println("OnDelete pod handler", obj)
}
