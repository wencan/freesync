package lockfree

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSList(t *testing.T) {
	slist := NewSinglyLinkedList()

	p, ok := slist.LeftPop()
	assert.Nil(t, p)
	assert.False(t, ok)

	slist.RightPush(1)
	p = slist.RightPeek()
	num := p.(int)
	assert.Equal(t, 1, num)

	slist.RightPush(2)
	p = slist.RightPeek()
	num = p.(int)
	assert.Equal(t, 2, num)

	p, ok = slist.LeftPop()
	if assert.True(t, ok) {
		num = p.(int)
		assert.Equal(t, 1, num)
	}

	// 实现限制，不能pop最后一个
}

func TestSList_ConcurrentlyRightPush(t *testing.T) {
	big := 200 * 10000
	rand.Seed(time.Now().UnixNano())
	ch := make(chan int, big)
	for _, num := range rand.Perm(big) {
		ch <- num
	}
	close(ch)

	slist := NewSinglyLinkedList()

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for num := range ch {
				slist.RightPush(num)
			}
		}()
	}
	wg.Wait()

	mapping := make(map[int]int, big)
	for {
		p, ok := slist.LeftPop()
		if !ok {
			break
		}
		num := p.(int)
		mapping[num] = 1
	}
	missMapping := make(map[int]int, big)
	for i := 0; i < big; i++ {
		if mapping[i] != 1 {
			missMapping[i] = 1
		}
	}
	if len(missMapping) == 1 {
		// 实现限制必须要留一个元素
	} else {
		for idx := range missMapping {
			t.Errorf("not found %d", idx)
		}
	}
}

func TestSList_ConcurrentlyPushAndPop(t *testing.T) {
	// 同时并发push和pop
	big := 200 * 10000

	rand.Seed(time.Now().UnixNano())
	ch := make(chan int, big)
	for _, num := range rand.Perm(big) {
		ch <- num
	}
	close(ch)

	slist := NewSinglyLinkedList()

	var wg sync.WaitGroup
	var count uint64
	var mapping sync.Map
	wg.Add(500)
	for i := 0; i < 250; i++ {
		go func() {
			defer wg.Done()

			for num := range ch {
				slist.RightPush(num)
				atomic.AddUint64(&count, 1)
			}
		}()
	}
	for i := 0; i < 250; i++ {
		go func() {
			defer wg.Done()

			for {
				p, ok := slist.LeftPop()
				if !ok {
					if atomic.LoadUint64(&count) == uint64(big) {
						break
					}
					runtime.Gosched()
					continue
				}

				num := p.(int)
				mapping.Store(num, 1)
			}
		}()
	}
	wg.Wait()

	missMapping := make(map[int]int, big)
	for i := 0; i < big; i++ {
		p, _ := mapping.Load(i)
		num, _ := p.(int)
		if num != 1 {
			missMapping[i] = 1
		}
	}
	if len(missMapping) == 1 {
		// 实现限制必须要留一个元素
	} else {
		for idx := range missMapping {
			t.Errorf("not found %d", idx)
		}
	}
}
