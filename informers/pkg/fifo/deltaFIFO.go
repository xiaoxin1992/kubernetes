package fifo

import (
	"fmt"
	"k8s.io/client-go/tools/cache"
)

/*
这是一个先先出的队列，主要用于在client-go中，reflector的list和watch中数据存储
reflector 通过list或watch把数据写入到此队列，
此队列提供Add, Update, Delete, Pop等方法来操作队列
deltaFIFO底层通过加锁的map实现
*/

func keyFunc(obj interface{}) (string, error) {
	/*
		用于获取key， 比如pod的name作为DeltaFIFO的key，所以通过这个函数进行返回
	*/
	return obj.(pod).Name, nil
}

type pod struct {
	Name  string
	Value float64
}

func deltaFIFO() {
	/*
		KnownObjects: 存储一份数据，主要用于判断当前数据是否被删除或是否需要替换
	*/
	store := cache.NewStore(keyFunc)
	df := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KeyFunction:  keyFunc,
		KnownObjects: store,
	})
	pod1 := pod{"pod1", 1.0}
	pod2 := pod{"pod2", 2.0}
	df.Add(pod1)
	df.Add(pod2)
	pod1.Value = 3.0
	df.Update(pod1)

	/*
		pod1 1 Added
		pod1 3 Updated
		df.Pop函数会弹出存入的数据，数据是先进先出，多个数据需要for循环进行弹出
		同一个数据多次更新或删除会一次性弹出,
		事件一般是: Added, Update, Delete, Sync
	*/
	df.Pop(func(obj interface{}, isInInitialList bool) error {
		for _, delta := range obj.(cache.Deltas) {
			fmt.Printf("EventType: %v, dataName: %v, dataValue: %v\n",
				delta.Type,
				delta.Object.(pod).Name,
				delta.Object.(pod).Value,
			)
		}
		return nil
	})
}
