package broadcastchannel

import (
	"reflect"
	"sync"
	"testing"
)

func TestBroadcast(t *testing.T) {
	b := New[int](0)
	defer b.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Create two subscribers.
	r1 := []int{}
	ch1 := make(chan int)
	b.Subscribe(ch1)
	go func() {
		for v := range ch1 {
			r1 = append(r1, v)
		}
		wg.Done()
	}()

	r2 := []int{}
	ch2 := make(chan int)
	b.Subscribe(ch2)
	go func() {
		for v := range ch2 {
			r2 = append(r2, v)
		}
		wg.Done()
	}()

	// Send a value.
	b.Chan() <- 1
	b.Close()

	wg.Wait()
	if !reflect.DeepEqual(r1, []int{1}) {
		t.Errorf("r1 = %v, want %v", r1, []int{1})
	}
	if !reflect.DeepEqual(r2, []int{1}) {
		t.Errorf("r2 = %v, want %v", r2, []int{1})
	}
}

func TestWithWrongReciever(t *testing.T) {
	b := New[int](0)
	defer b.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Create two subscribers.
	r1 := []int{}
	ch1 := make(chan int)
	b.Subscribe(ch1)
	go func() {
		for v := range ch1 {
			r1 = append(r1, v)
		}
		wg.Done()
	}()

	r2 := []int{}
	ch2 := make(chan int)
	b.Subscribe(ch2)
	go func() {
		// Here, ch2 is read only onece.
		r2 = append(r2, <-ch2)
		wg.Done()
	}()

	// Send two value.
	b.Chan() <- 1
	b.Chan() <- 2
	b.Close()

	wg.Wait()
	if !reflect.DeepEqual(r1, []int{1, 2}) {
		t.Errorf("r1 = %v, want %v", r1, []int{1, 2})
	}
	if !reflect.DeepEqual(r2, []int{1}) {
		t.Errorf("r2 = %v, want %v", r2, []int{1})
	}
}
