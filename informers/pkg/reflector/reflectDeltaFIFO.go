package reflector

import (
	"fmt"
	"informers/pkg/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

/*
	从reflector获取到的数据需要存放到DeltaFIFO队列中, 下面实现下
*/

func reflectDelta() {
	client := config.InitClient()
	listAndWatch := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	/*
			cache.MetaNamespaceKeyFunc 最终实现判断namespace是否为空，
			如果为空则返回{Namespace: "", Name: obj.GetName()}
		    否则返回{Namespace: obj.GetNamespace(), Name: obj.GetName()}
			KnownObjects 用于存储 Reflector 已同步的资源对象帮助 Reflector 处理增量更新和保持缓存一致性。
			DeletionHandlingMetaNamespaceKeyFunc 在删除事件中，资源对象可能已经不存在，只有它的 DeletedFinalStateUnknown 包装器传递给回调函数。DeletionHandlingMetaNamespaceKeyFunc 能够正确处理这种情况，确保键的生成不会出错
	*/
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction:  cache.MetaNamespaceKeyFunc,
		KnownObjects: store,
	})
	/*
	   lw cache.ListerWatcher,       // List 和 Watch 机制的实现
	   expectedType runtime.Object, // 预期的资源类型
	   store cache.Store,           // 本地缓存，用于存储资源对象
	   resyncPeriod time.Duration,  // 重新同步周期
	*/
	rf := cache.NewReflector(listAndWatch, &v1.Pod{}, df, 0)
	ch := make(chan struct{})
	go func() {
		rf.Run(ch)
	}()
	for {
		df.Pop(func(obj interface{}, isInInitialList bool) error {
			for _, delta := range obj.(cache.Deltas) {
				fmt.Printf("deltaEvent: %v podName: %v podStaus: %v isInInitialList: %v\n", delta.Type, delta.Object.(*v1.Pod).Name, delta.Object.(*v1.Pod).Status.Phase, isInInitialList)
				switch delta.Type {
				case cache.Sync, cache.Added:
					store.Add(delta.Object)
				case cache.Updated:
					store.Update(delta.Object)
				case cache.Deleted:
					store.Delete(delta.Object)
				}
			}
			return nil
		})
	}
}
