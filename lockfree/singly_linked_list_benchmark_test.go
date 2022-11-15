package lockfree

import (
	"sync"
	"testing"
)

func BenchmarkSinglyLinkedList_pushAndPop(b *testing.B) {
	slist := NewSinglyLinkedList()
	for i := 0; i < 10000; i++ {
		slist.RightPush(i)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			n := i
			slist.RightPush(n)
			slist.LeftPop()

			i++
		}
	})
}

func BenchmarkChannel_pushAndPop(b *testing.B) {
	ch := make(chan int, 10000000)
	for i := 0; i < 10000; i++ {
		ch <- i
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			n := i
			ch <- n
			<-ch

			i++
		}
	})
}

func BenchmarkSyncPool_putAndGet(b *testing.B) {
	var pool sync.Pool

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		var i int
		for p.Next() {
			n := i
			pool.Put(&n)
			pool.Get()

			i++
		}
	})
}
