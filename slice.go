package freesync

import (
	"sync"
	"sync/atomic"

	"github.com/wencan/freesync/lockfree"
)

// Slice 并发安全的Slice结构。
type Slice struct {
	// mux 锁。
	mu sync.Mutex

	// store 实质存储数据。内部结构为*lockfree.Slice。
	// slice增长时，需要加锁。
	store atomic.Value
}

// Append 在末尾追加一个元素。返回下标。
func (slice *Slice) Append(p interface{}) int {
	store, _ := slice.store.Load().(*lockfree.Slice)
	if store != nil {
		if index, ok := store.Append(p); ok {
			return index
		}
	}

	slice.mu.Lock()
	defer slice.mu.Unlock()

	if store == nil {
		// 初始化
		store, _ = slice.store.Load().(*lockfree.Slice)
		if store == nil {
			store = &lockfree.Slice{}
			slice.store.Store(store)
		}
	} else {
		previous := store
		store = slice.store.Load().(*lockfree.Slice)
		if store != previous {
			if index, ok := store.Append(p); ok {
				return index
			}
		}
	}

	// 增加容量后再append
	newStore, _ := store.Grow()
	index, ok := newStore.Append(p)
	if !ok {
		panic("impossibility")
	}
	slice.store.Store(newStore)

	return index
}

// Load 取得下标位置上的值。
func (slice *Slice) Load(index int) interface{} {
	store, _ := slice.store.Load().(*lockfree.Slice)
	if store == nil {
		panic("empty slice")
	}
	return store.Load(index)
}

// Range 遍历。
func (slice *Slice) Range(f func(index int, p interface{}) (stopIteration bool)) {
	store, _ := slice.store.Load().(*lockfree.Slice)
	if store == nil {
		return
	}
	store.Range(f)
}

// Length 长度。
func (slice *Slice) Length() int {
	var length int
	slice.Range(func(index int, p interface{}) (stopIteration bool) {
		length++
		return false
	})
	return length
}

// UpdateAt 更新下标位置上的值，返回旧值。
func (slice *Slice) UpdateAt(index int, p interface{}) (old interface{}) {
	store, _ := slice.store.Load().(*lockfree.Slice)
	if store == nil {
		panic("empty slice")
	}
	return store.UpdateAt(index, p)
}
