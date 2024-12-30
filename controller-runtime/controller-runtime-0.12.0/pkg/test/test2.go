package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

func mainTest3() {
	mgr, err := manager.New(k8sConfig(),
		manager.Options{
			Logger:    logf.Log.WithName("test"),
			Namespace: v1.NamespaceAll,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mgr.GetScheme().ObjectKinds(&v1.Pod{}))
	fmt.Println("starting manager")
	go func() {
		time.Sleep(2 * time.Second)
		podInformer, _ := mgr.GetCache().GetInformer(context.Background(), &v1.Pod{})
		fmt.Printf("%T", podInformer)
		fmt.Println(podInformer.(cache.SharedIndexInformer).GetIndexer().ListKeys())
	}()
	err = mgr.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
