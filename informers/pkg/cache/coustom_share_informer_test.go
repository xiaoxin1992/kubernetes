package cache

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"informers/pkg/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"testing"
	"time"
)

func TestCustomShareInformer1(t *testing.T) {
	lis := newProcessListener(&PodHandler{}, 0, 0, time.Now(), 0, nil)
	go func() {
		count := 0
		for {
			time.Sleep(1 * time.Second)
			pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("pod-%d", count)}}
			lis.addCh <- addNotification{
				newObj:          pod,
				isInInitialList: false,
			}
			count++
		}
	}()
	go func() {
		lis.pop()
	}()
	lis.run()
}

func TestCustomShareInformer2(t *testing.T) {
	// 不带缓存的调用
	client := config.InitClient()
	listAndWatch := NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	csi := NewCustomShareInformer(listAndWatch, &v1.Pod{})
	csi.AddEventHandler(&PodHandler{})
	csi.Start(make(chan struct{}))
}

func TestCustomShareInformer3(t *testing.T) {
	//  indexer, 基本使用, 通过函数过滤pod
	indexers := Indexers{"app": func(obj interface{}) ([]string, error) {
		meta, err := meta.Accessor(obj)
		if err != nil {
			return []string{""}, fmt.Errorf("object does not have metadata %v", err)
		}
		return []string{meta.GetLabels()["app"]}, nil
	}}
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns1",
			Labels: map[string]string{
				"app": "l1",
			},
		},
	}
	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "ns2",
			Labels: map[string]string{
				"app": "l2",
			},
		},
	}
	myIndex := NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers)
	myIndex.Add(pod1)
	myIndex.Add(pod2)
	fmt.Println(myIndex.IndexKeys("app", "l2"))
}

func TestCustomShareInformer4(t *testing.T) {
	// 缓存的调用
	// indexer缓存的调用
	indexers := Indexers{NamespaceIndex: MetaNamespaceIndexFunc}
	myIndex := NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers)
	go func() {
		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			data, err := myIndex.IndexKeys(NamespaceIndex, "default")
			if err != nil {
				fmt.Println("index error:", err)
			}
			c.JSON(200, gin.H{"data": data})
		})
		r.Run(":8080")
	}()
	client := config.InitClient()
	listAndWatch := NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	csi := NewCustomShareInformerIndexer(listAndWatch, &v1.Pod{}, myIndex)
	csi.AddEventHandler(&PodHandler{})
	csi.Start(wait.NeverStop)
}
