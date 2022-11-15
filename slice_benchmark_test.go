package freesync

import (
	"sync"
	"testing"
)

func BenchmarkSlice_Append(b *testing.B) {
	var slice Slice

	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			slice.Append(i)

			i++
		}
	})
}

func BenchmarkMutexSlice_Append(b *testing.B) {
	var slice []int
	var mu sync.Mutex

	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			mu.Lock()
			slice = append(slice, i)
			mu.Unlock()

			i++
		}
	})
}

func BenchmarkRWMutexSlice_Append(b *testing.B) {
	var slice []int
	var mu sync.RWMutex

	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			mu.Lock()
			slice = append(slice, i)
			mu.Unlock()

			i++
		}
	})
}

func BenchmarkSlice_Load(b *testing.B) {
	var slice Slice

	for i := 0; i < 10000; i++ {
		slice.Append(i)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			if i >= 10000 {
				i = 0
			}
			slice.Load(i)
			i++
		}
	})
}

func BenchmarkMutexSlice_Load(b *testing.B) {
	var slice []int
	var mu sync.Mutex

	for i := 0; i < 10000; i++ {
		slice = append(slice, i)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			if i >= 10000 {
				i = 0
			}
			mu.Lock()
			_ = slice[i]
			mu.Unlock()
			i++
		}
	})
}

func BenchmarkRWMutexSlice_Load(b *testing.B) {
	var slice []int
	var mu sync.RWMutex

	for i := 0; i < 10000; i++ {
		slice = append(slice, i)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			if i >= 10000 {
				i = 0
			}
			mu.RLock()
			_ = slice[i]
			mu.RUnlock()
			i++
		}
	})
}
