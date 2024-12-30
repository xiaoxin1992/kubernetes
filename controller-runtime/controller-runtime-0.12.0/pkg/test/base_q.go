package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func newItem(name, ns string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: ns,
		},
	}
}

func main() {
	que := workqueue.New()
	go func() {
		for {
			item, _ := que.Get()
			fmt.Println(item.(reconcile.Request).Namespace)
			que.Done(item)
			time.Sleep(1 * time.Millisecond)
			que.ShutDown()
		}
	}()
	for {
		que.Add(newItem("abc", "default"))
		time.Sleep(1 * time.Millisecond)
	}
	time.Sleep(10 * time.Second)
}
