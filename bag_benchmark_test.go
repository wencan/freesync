package freesync

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkBagAdd(b *testing.B) {
	bag := NewBag()

	var number uint64

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint64(&number, 1)
			bag.Add(i)
		}
	})
}

func BenchmarkSyncMapAdd(b *testing.B) {
	var mapping sync.Map

	var number uint64

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint64(&number, 1)

			mapping.Store(i, i)
		}
	})
}

func BenchmarkMutexMapAdd(b *testing.B) {
	mapping := make(map[uint64]uint64)
	var mu sync.Mutex

	var number uint64

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			i := atomic.AddUint64(&number, 1)

			mu.Lock()
			mapping[i] = i
			mu.Unlock()
		}
	})
}

func BenchmarkBagWrite(b *testing.B) {
	bag := NewBag()

	ch := make(chan int, 10000000)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			index := bag.Add(i)
			ch <- index

			i++

			index = <-ch
			bag.DeleteAt(index)
		}
	})
}

func BenchmarkSyncMapWrite(b *testing.B) {
	var mapping sync.Map

	ch := make(chan int, 10000000)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			mapping.Store(i, i)
			ch <- i

			i++

			delI := <-ch
			mapping.Delete(delI)
		}
	})
}

func BenchmarkMutexMapWrite(b *testing.B) {
	var mu sync.Mutex
	mapping := make(map[int]int)

	ch := make(chan int, 10000000)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			mu.Lock()
			mapping[i] = i
			mu.Unlock()

			ch <- i

			i++

			i = <-ch

			mu.Lock()
			delete(mapping, i)
			mu.Unlock()
		}
	})
}

func BenchmarkBagRange(b *testing.B) {
	bag := NewBag()

	for i := 0; i < 10000; i++ {
		bag.Add(i)
	}

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bag.Range(func(index int, p interface{}) (stopIteration bool) {
				return false
			})
		}
	})
}

func BenchmarkSyncMapRange(b *testing.B) {
	var mapping sync.Map

	for i := 0; i < 10000; i++ {
		mapping.Store(i, i)
	}

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			mapping.Range(func(key, value any) bool {
				return true
			})
		}
	})
}
