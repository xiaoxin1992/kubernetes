package main

import (
	"fmt"
	"informers/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"time"
)

func main() {
	// 创建 Kubernetes 客户端

	clientset := config.InitClient()

	// 创建 SharedInformer 工厂
	factory := informers.NewSharedInformerFactory(clientset, time.Minute)
	// 获取 Pod 的 SharedInformer
	podInformer := factory.Core().V1().Pods().Informer()

	// 注册事件处理程序
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("Pod added: %s/%s\n", pod.Namespace, pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			//newPod := newObj.(*corev1.Pod)
			fmt.Printf("Pod updated: %s/%s\n", oldPod.Namespace, oldPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	// 启动 SharedInformer
	stopCh := make(chan struct{})
	defer close(stopCh)

	go podInformer.Run(stopCh)

	// 等待缓存同步完成
	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		panic("Failed to sync cache")
	}

	// 阻塞以保持程序运行
	select {}
}
