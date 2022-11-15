package lockfree

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimitedSlice(t *testing.T) {
	slice := NewLimitedSlice(10)
	for i := 0; i < 10; i++ {
		index, ok := slice.Append(i)
		assert.True(t, ok)
		assert.Equal(t, i, index)
	}
	assert.Equal(t, 10, slice.Length())

	_, ok := slice.Append(11)
	assert.False(t, ok)

	for i := 0; i < 10; i++ {
		num, _ := slice.Load(i).(int)
		assert.Equal(t, i, num)
	}
}

func TestLimitedSlice_ConcurrentlyAppend(t *testing.T) {
	slice := NewLimitedSlice(500 * 10000)

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 10000; j++ {
				_, ok := slice.Append(123)
				assert.True(t, ok)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 500*10000, slice.Length())

	// 空间满后，失败的append
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				_, ok := slice.Append(123)
				assert.False(t, ok)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 500*10000, slice.Length())
}

func TestLimitedSlice_ConcurrentlyAppend2(t *testing.T) {
	// 一组顺序的数字，并发随机append
	big := 100 * 10000
	slice := NewLimitedSlice(big)

	rand.Seed(time.Now().UnixNano())
	ch := make(chan int, big)
	for _, num := range rand.Perm(big) {
		ch <- num
	}
	close(ch)

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for {
				num, ok := <-ch
				if !ok {
					break
				}

				_, ok = slice.Append(num)
				assert.True(t, ok)
			}
		}()
	}
	wg.Wait()

	// 检查
	var all = make(map[int]int, 1000000)
	for i := 0; i < big; i++ {
		value := slice.Load(i)
		num := value.(int)
		all[num] = 1
	}
	assert.Equal(t, big, len(all))

	for i := 0; i < big; i++ {
		assert.Equal(t, 1, all[i], "not found %d", i)
	}
}

func TestLimitedSlice_ConcurrentlyLoad(t *testing.T) {
	slice := NewLimitedSlice(10000)
	for i := 0; i < 10000; i++ {
		_, ok := slice.Append(i)
		assert.True(t, ok)
	}

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 10000; j++ {
				num := slice.Load(j).(int)
				if num != j {
					assert.Equal(t, j, num) // assert.Equal较慢
				}
			}
		}()
	}
	wg.Wait()
}

func TestLimitedSlice_concurrentlyAppendLoadUpdateRange(t *testing.T) {
	big := 1000000
	slice := NewLimitedSlice(big)

	ch := make(chan int, big)
	rand.Seed(time.Now().UnixNano())
	for _, num := range rand.Perm(big) {
		ch <- num
	}
	close(ch)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				num, ok := <-ch
				if !ok {
					return
				}

				index, ok := slice.Append(num)
				assert.True(t, ok)

				// 将value更新为index值，方便后面检查
				old := slice.UpdateAt(index, index)
				assert.Equal(t, old, num)

				got := slice.Load(index).(int)
				assert.Equal(t, index, got)
			}
		}()
	}

	closeFlag := make(chan struct{})
	go func() {
		wg.Wait()
		close(closeFlag)
	}()

	var wg2 sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()

			for {
				select {
				case <-closeFlag:
					return
				default:
				}

				slice.Range(func(index int, p interface{}) (stopIteration bool) {
					_ = p.(int)
					return false
				})
			}
		}()
	}
	wg2.Wait()

	// 最后再检查一遍
	slice.Range(func(index int, p interface{}) (stopIteration bool) {
		num := p.(int)
		assert.Equal(t, index, num)
		return false
	})
}
