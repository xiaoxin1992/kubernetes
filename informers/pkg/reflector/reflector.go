package reflector

import (
	"fmt"
	"informers/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

/*
	reflector 主要是用于做list或watch的操作
	从apiServer请求数据同步
*/

func printPodList(pods *v1.PodList) {
	for _, pod := range pods.Items {
		fmt.Printf("Name: %v, Labels: %v\n", pod.Name, pod.Labels)
	}
}

func reflector() {
	// 生成客户端
	client := config.InitClient()
	// fields.Everything() 表示不进行任何过滤
	//fieldSelector := fields.Everything()
	// 过滤执行的数据，我们选择fields.SelectorFromSet()来设置
	fieldSelector := fields.SelectorFromSet(fields.Set{"status.phase": "Running", "metadata.name": "test"})
	listAndWatch := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fieldSelector)
	pods, err := listAndWatch.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	printPodList(pods.(*v1.PodList))
	fmt.Printf("watcher starting...\n")
	watcher, err := listAndWatch.Watch(metav1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("watch error %v", err))
	}
	for {
		select {
		case data, ok := <-watcher.ResultChan():
			if ok {
				fmt.Printf("type: %v: podName: %v\n", data.Type, data.Object.(*v1.Pod).Name)
			}
		}
	}
}
