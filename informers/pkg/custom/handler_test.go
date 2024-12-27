package custom

import (
	"informers/pkg/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"testing"
)

func TestPodHandler(t *testing.T) {
	client := config.InitClient()
	listAndWatch := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	informer := NewInformer(listAndWatch, &v1.Pod{}, &PodHandler{})
	informer.Start(make(chan struct{}))
}
