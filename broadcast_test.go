package broadcastchannel

import (
	"sync"
	"testing"
)

func TestBroadcast(t *testing.T) {
	b := New[int](0)
	defer b.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Create two subscribers.
	ch1 := make(chan int, 0)
	b.Subscribe(ch1)
	go func() {
		for v := range ch1 {
			if v != 1 {
				t.Fatalf("expected %d, got %d", 1, v)
			}
		}
		wg.Done()
	}()

	ch2 := make(chan int, 0)
	b.Subscribe(ch2)
	go func() {
		for v := range ch2 {
			if v != 1 {
				t.Fatalf("expected %d, got %d", 1, v)
			}
		}
		wg.Done()
	}()

	// Send a value.
	b.Chan() <- 1
	b.Close()

	wg.Wait()
}
