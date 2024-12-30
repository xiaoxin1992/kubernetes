package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

func main2Test() {
	p := &v1.Pod{}
	fmt.Println(reflect.TypeOf(p).Elem().Name())
	s := runtime.NewScheme()
	fmt.Println(s.AllKnownTypes())
	// scheme.AddToScheme(s)  所有类型
	v1.AddToScheme(s) // core v1 的类型 注册gvk
	fmt.Println(s.AllKnownTypes())
	fmt.Println(s.ObjectKinds(&v1.Pod{}))

}

func main1Test() {
	mgr, err := manager.New(k8sConfig(),
		manager.Options{
			Logger: logf.Log.WithName("test"),
			NewClient: func(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
				return cluster.DefaultNewClient(cache, config, options, &v1.Pod{})
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mgr.GetScheme().ObjectKinds(&v1.Pod{}))
	fmt.Println("starting manager")
	go func() {
		time.Sleep(2 * time.Second)
		pod := &v1.Pod{}
		// 设置缓存
		fmt.Printf("%T\n", mgr.GetClient())
		// get 是从缓存中读取
		err = mgr.GetClient().Get(
			context.TODO(),
			types.NamespacedName{Namespace: "default", Name: "test"},
			pod,
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(pod)
	}()
	err = mgr.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}

}
