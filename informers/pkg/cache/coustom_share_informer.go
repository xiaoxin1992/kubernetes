package cache

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"informers/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"time"
)

/*
 自定义实现share informer
*/

var _ ResourceEventHandler = &PodHandler{}

type PodHandler struct{}

func NewPodHandler() *PodHandler {
	return &PodHandler{}
}

func (h *PodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	fmt.Println("OnAdd shareInformer pod handler")
}
func (h *PodHandler) OnUpdate(oldObj, newObj interface{}) {
	fmt.Println("OnUpdate shareInformer pod handler")
}

func (h *PodHandler) OnDelete(obj interface{}) {
	fmt.Println("OnDelete shareInformer pod handler")
}

type CustomShareInformer struct {
	reflector *Reflector
	fifo      *DeltaFIFO
	store     Store
	processor *sharedProcessor
}

func (c *CustomShareInformer) AddEventHandler(handler ResourceEventHandler) {
	lis := newProcessListener(handler, 0, 0, time.Now(), initialBufferSize, nil)
	c.processor.addListener(lis)
}

func NewCustomShareInformerIndexer(listAnd *ListWatch, obj runtime.Object, indexer Indexer) *CustomShareInformer {
	fifo := NewDeltaFIFOWithOptions(DeltaFIFOOptions{
		KeyFunction:  MetaNamespaceKeyFunc,
		KnownObjects: indexer,
	})
	reflect := NewReflector(listAnd, obj, fifo, 0)
	return &CustomShareInformer{
		reflector: reflect,
		fifo:      fifo,
		processor: &sharedProcessor{},
		store:     indexer,
	}
}

func NewCustomShareInformer(listAnd *ListWatch, obj runtime.Object) *CustomShareInformer {
	store := NewStore(DeletionHandlingMetaNamespaceKeyFunc)
	fifo := NewDeltaFIFOWithOptions(DeltaFIFOOptions{
		KeyFunction:  MetaNamespaceKeyFunc,
		KnownObjects: store,
	})
	reflect := NewReflector(listAnd, obj, fifo, 0)
	return &CustomShareInformer{
		reflector: reflect,
		fifo:      fifo,
		store:     store,
		processor: &sharedProcessor{},
	}
}

func (c *CustomShareInformer) Start(stopCh <-chan struct{}) {
	go func() {
		for {
			c.fifo.Pop(func(obj interface{}, isInInitialList bool) error {
				for _, delta := range obj.(Deltas) {
					switch delta.Type {
					case Added, Sync:
						c.store.Add(delta.Object)
						c.processor.distribute(addNotification{newObj: delta.Object, isInInitialList: isInInitialList}, false)
					case Updated:
						if old, exists, err := c.store.Get(delta.Object); err == nil && exists {
							c.store.Update(delta.Object)
							c.processor.distribute(updateNotification{oldObj: old, newObj: delta.Object}, true)
						}
					case Deleted:
						c.store.Delete(delta.Object)
						c.processor.distribute(deleteNotification{delta.Object}, false)
					}
				}
				return nil
			})
		}
	}()
	go func() {
		c.reflector.Run(stopCh)
	}()
	c.processor.run(stopCh)
}

type MyFactory struct {
	client    *kubernetes.Clientset
	informers map[reflect.Type]SharedIndexInformer
}

func (my *MyFactory) PodInformer() SharedIndexInformer {
	if informer, ok := my.informers[reflect.TypeOf(v1.Pod{})]; ok {
		return informer
	}
	listAndWatch := NewListWatchFromClient(my.client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	indexers := Indexers{NamespaceIndex: MetaNamespaceIndexFunc}
	informer := NewSharedIndexInformer(listAndWatch, &v1.Pod{}, 0, indexers)
	my.informers[reflect.TypeOf(v1.Pod{})] = informer
	return informer
}

func (my *MyFactory) Start() {
	ch := wait.NeverStop
	for _, i := range my.informers {
		go func(informer SharedIndexInformer) {
			informer.Run(ch)
		}(i)
	}
}

func NewMyFactory(client *kubernetes.Clientset) *MyFactory {
	return &MyFactory{client: client, informers: map[reflect.Type]SharedIndexInformer{}}
}

func MyFactoryStart() {
	client := config.InitClient()
	fact := NewMyFactory(client)
	fact.PodInformer().AddEventHandler(ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInInitialList bool) {
			fmt.Println("AddFunc", obj.(metav1.Object).GetName())
		},
	})
	fact.Start()
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		data, err := fact.PodInformer().GetIndexer().IndexKeys(NamespaceIndex, "default")
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(200, gin.H{"data": data})
	})
	r.Run(":8080")

}
