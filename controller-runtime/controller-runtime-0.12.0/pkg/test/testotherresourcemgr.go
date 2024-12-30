package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"log"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/*
监听其他资源
*/

func AddCmWatch(ctl controller.Controller) error {
	src := source.Kind{Type: &v1.ConfigMap{}}
	envet := handler.Funcs{
		CreateFunc: func(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
			limitingInterface.Add(reconcile.Request{
				types.NamespacedName{
					Name: "abc", Namespace: "default1",
				},
			})
		},
	}
	return ctl.Watch(&src, envet)
}

type Ctl1 struct{}

func (c *Ctl1) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	fmt.Println(req.NamespacedName)
	return controllerruntime.Result{}, nil
}

func mainmgr1() {
	mgr, err := manager.New(k8sConfig(), manager.Options{
		Logger:    logf.Log.WithName("mgr_test"),
		Namespace: "default",
	})
	if err != nil {
		log.Fatal(err)
	}
	ctl, err := controller.New("abc", mgr, controller.Options{
		Reconciler:              &Ctl1{},
		MaxConcurrentReconciles: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	src := source.Kind{Type: &v1.Pod{}}
	handler := &handler.EnqueueRequestForObject{}
	err = ctl.Watch(&src, handler)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(AddCmWatch(ctl))
	err = mgr.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
