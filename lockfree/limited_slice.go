package lockfree

import (
	"sync/atomic"
)

// LimitedSliceEntry 包装要保存的数据。
// LimitedSliceEntry指针为nil，表示还未初始化；LimitedSliceEntry.p为nil，表示用户存了一个数据nil。
// atomic.Value不支持存nil，不支持更新不同类型的值，使用LimitedSliceEntry，绕过了atomic.Value的限制。
type LimitedSliceEntry struct {
	p interface{}
}

// LimitedSlice 长度受限的Slice。
type LimitedSlice struct {
	// array 值为*LimitedSliceEntry。LimitedSliceEntry内p存的是保存的数据。
	array []atomic.Value

	// entites 预分配的LimitedSliceEntry
	entites []LimitedSliceEntry

	// capacity 容量。
	// 容量不会发生变化。
	capacity int

	// nextAppendIndex 下次append元素的位置。无并发场景下，等于长度。
	nextAppendIndex uint64
}

// NewLimitedSlice 新建一个长度受限的Slice。
func NewLimitedSlice(capacity int) *LimitedSlice {
	return &LimitedSlice{
		array:           make([]atomic.Value, capacity),
		entites:         make([]LimitedSliceEntry, capacity),
		capacity:        capacity,
		nextAppendIndex: 0,
	}
}

// Capacity 容量。
func (slice *LimitedSlice) Capacity() int {
	return slice.capacity
}

// Append 追加新元素。
// 如果成功，返回下标。
// 如果已满，返回false。
func (slice *LimitedSlice) Append(p interface{}) (int, bool) {
	for {
		index := atomic.LoadUint64(&slice.nextAppendIndex)
		if index+1 > uint64(slice.capacity) {
			return 0, false
		}

		if atomic.CompareAndSwapUint64(&slice.nextAppendIndex, index, index+1) {
			// 这里需要警惕，length增长了，但数据还没存进去。
			// 等到Store完成，才算Append结束。
			entry := &slice.entites[index]
			entry.p = p
			slice.array[index].Store(entry)
			return int(index), true
		}
	}
}

// Load 根据下标取回一个元素。
func (slice *LimitedSlice) Load(index int) interface{} {
	entry := slice.array[index].Load().(*LimitedSliceEntry)
	return entry.p
}

// UpdateAt 更新下标位置上的元素，返回旧值。
func (slice *LimitedSlice) UpdateAt(index int, p interface{}) (old interface{}) {
	newEntry := &LimitedSliceEntry{p: p}
	oldVal := slice.array[index].Swap(newEntry)
	// 不能回收Swap返回的entry。
	// 因为可能另一个过程刚刚拿到这个entry。
	oldEntry := oldVal.(*LimitedSliceEntry)
	return oldEntry.p
}

// Range 遍历。
func (slice *LimitedSlice) Range(f func(index int, p interface{}) (stopIteration bool)) {
	length := int(atomic.LoadUint64(&slice.nextAppendIndex))
	for index := 0; index < length; index++ {
		val := slice.array[index].Load()
		if val == nil {
			// nextAppendIndex增长了，但数据还没存进去
			continue
		}

		entry := val.(*LimitedSliceEntry)
		stopIteration := f(index, entry.p)
		if stopIteration {
			break
		}
	}
}

// Length 长度。
func (slice *LimitedSlice) Length() int {
	var length int
	slice.Range(func(index int, p interface{}) (stopIteration bool) {
		length++
		return false
	})
	return length
}
