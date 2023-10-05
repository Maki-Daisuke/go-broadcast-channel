package broadcastchannel

import (
	"reflect"
	"sync"
	"testing"
	"time"
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

func TestWithWrongReceiver(t *testing.T) {
	b := New[int](0).WithTimeout(1 * time.Second)
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
		// Here, ch2 is read only once.
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

func TestWithClosingReceiver(t *testing.T) {
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
		// Here, ch2 is read only once and is closed after the read.
		r2 = append(r2, <-ch2)
		close(ch2)
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

func TestCloseTwice(t *testing.T) {
	b := New[int](0)
	if err := b.Close(); err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if err := b.Close(); err == nil {
		t.Errorf("err = %v, want non-nil error", err)
	}
}
