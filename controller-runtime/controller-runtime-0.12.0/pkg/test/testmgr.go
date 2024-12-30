package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"log"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	cc "sigs.k8s.io/controller-runtime/pkg/internal/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/*
手动实现mgr，以及手动出发事件
*/
var _ manager.Runnable = &myWeb{}

type myWeb struct {
	event handler.EventHandler
	ctl   controller.Controller
}

func newMyWeb(ctl controller.Controller, event handler.EventHandler) *myWeb {
	return &myWeb{ctl: ctl, event: event}
}

func (m myWeb) Start(ctx context.Context) error {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		pod := &v1.Pod{}
		pod.Name = "mytestpod"
		pod.Namespace = "default"

		m.event.Create(event.CreateEvent{Object: pod}, m.ctl.(*cc.Controller).Queue)
	})
	return r.Run(":8081")
}

type Ctl struct{}

func (c *Ctl) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	fmt.Println(req.NamespacedName)
	return controllerruntime.Result{}, nil
}

func mainMgr() {
	mgr, err := manager.New(k8sConfig(), manager.Options{
		Logger:    logf.Log.WithName("mgr_test"),
		Namespace: "default",
	})
	if err != nil {
		log.Fatal(err)
	}
	ctl, err := controller.New("abc", mgr, controller.Options{
		Reconciler: &Ctl{},
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
	mgr.Add(newMyWeb(ctl, handler))
	err = mgr.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
