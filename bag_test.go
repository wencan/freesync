package freesync

import (
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBag(t *testing.T) {
	bag := NewBag()

	getAll := func() []int {
		ints := []int{}
		bag.Range(func(index int, p interface{}) (stopIteration bool) {
			value := p.(int)
			ints = append(ints, value)
			return false
		})
		sort.Ints(ints)
		return ints
	}

	// 放 0-100
	delIndexes := make([]int, 0, 100)
	for i := 0; i < 100; i++ {
		index := bag.Add(i)
		delIndexes = append(delIndexes, index)
	}
	assert.Equal(t, 100, bag.Length())

	// 输出全部
	all := getAll()
	want := make([]int, 0, 100)
	for i := 0; i < 100; i++ {
		want = append(want, i)
	}
	sort.Ints(want)
	assert.Equal(t, want, all)

	// 删除中间的0、10、20...
	for _, index := range delIndexes {
		if index%10 == 0 {
			bag.DeleteAt(index)
		}
	}
	// 再对比
	all = getAll()
	want = make([]int, 0, 90)
	for i := 0; i < 100; i++ {
		if i%10 != 0 {
			want = append(want, i)
		}
	}
	sort.Ints(want)
	assert.Equal(t, want, all)
	assert.Equal(t, 90, bag.Length())

	// 继续添加
	for i := 100; i < 200; i++ {
		bag.Add(i)
	}
	// 再对比
	all = getAll()
	want = make([]int, 0, 190)
	for i := 0; i < 200; i++ {
		if i%10 == 0 && i < 100 {
			continue
		}
		want = append(want, i)
	}
	sort.Ints(want)
	assert.Equal(t, want, all)
	assert.Equal(t, 190, bag.Length())
}

func TestBagConcurrentlyUpdate(t *testing.T) {
	bag := NewBag()
	big := 50000

	// 并发添加/删除
	var wg sync.WaitGroup
	delIndexChans := make(chan int, big)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for _, num := range rand.Perm(big) {
				index := bag.Add(num)
				delIndexChans <- index

				index = <-delIndexChans
				bag.DeleteAt(index)
			}
		}()
	}

	// 这里添加的不删除
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, num := range rand.Perm(big) {
			bag.Add(num)
		}
	}()

	wg.Wait()

	// 检查长度
	assert.Equal(t, big, bag.Length())

	// 遍历检查
	all := make([]int, 0, big)
	bag.Range(func(index int, p interface{}) (stopIteration bool) {
		num := p.(int)
		all = append(all, num)
		return false
	})
	sort.Ints(all)

	// 检查剩下的
	want := make([]int, 0, big)
	for i := 0; i < big; i++ {
		want = append(want, i)
	}
	assert.Equal(t, want, all)
}

func TestBagConcurrentlyUpdateAndRange(t *testing.T) {
	bag := NewBag()
	big := 50000

	// 并发添加/删除
	var wg sync.WaitGroup
	delIndexChans := make(chan int, big)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for _, num := range rand.Perm(big) {
				index := bag.Add(num)
				delIndexChans <- index

				index = <-delIndexChans
				bag.DeleteAt(index)
			}
		}()
	}

	var done uint64
	go func() {
		wg.Wait()
		atomic.StoreUint64(&done, 1)
	}()

	var wg2 sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()

			for {
				if atomic.LoadUint64(&done) == 1 {
					return
				}

				bag.Range(func(index int, p interface{}) (stopIteration bool) {
					_ = p.(int)
					return false
				})
			}
		}()
	}
	wg2.Wait()

	assert.Equal(t, 0, bag.Length())
}
