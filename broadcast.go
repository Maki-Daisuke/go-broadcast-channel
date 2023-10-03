package broadcastchannel

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

type Broadcaster[T any] struct {
	valCh     chan T
	subCh     chan chan<- T
	sendCases []reflect.SelectCase
	timeout   time.Duration
}

// New creates a new Broadcaster.
// It takes a buffer size for the channel.
func New[T any](n int) *Broadcaster[T] {
	b := &Broadcaster[T]{
		valCh:   make(chan T, n),
		subCh:   make(chan chan<- T),
		timeout: math.MaxInt64, // Eventually, this will never timeout by default.
	}
	go b.run()
	return b
}

func (b *Broadcaster[T]) WithTimeout(d time.Duration) *Broadcaster[T] {
	b.timeout = d
	return b
}

func (b *Broadcaster[T]) run() {
	for {
		select {
		case v, ok := <-b.valCh:
			if !ok {
				b.destroy()
				return
			}
			b.broadcast(v)
		case ch := <-b.subCh:
			b.sendCases = append(b.sendCases, reflect.SelectCase{Dir: reflect.SelectSend, Chan: reflect.ValueOf(ch)})
		}
	}
}

func (b *Broadcaster[T]) destroy() {
	defer func() {
		if r := recover(); r != nil {
			// A channel maybe already closed. So, ignore it.
			return
		}
	}()
	close(b.subCh)
	for _, c := range b.sendCases {
		c.Chan.Close()
	}
}

func (b *Broadcaster[T]) broadcast(v T) {
	val := reflect.ValueOf(v)
	for i := range b.sendCases {
		b.sendCases[i].Send = val
	}

	timeout := time.NewTimer(b.timeout)
	defer timeout.Stop()

	cases := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(timeout.C)}}
	cases = append(cases, b.sendCases...)

	for len(cases) > 1 {
		// Wait for one of the cases to be ready.
		chosen, _, _ := reflect.Select(cases)
		if chosen == 0 {
			// Timeout.
			break
		}
		// Remove the case that was done.
		cases = append(cases[:chosen], cases[chosen+1:]...)
	}
	for _, c := range cases[1:] {
		// If there are remaining cases, they are timed out.
		// So, remove them from the broadcaster.
		for i := range b.sendCases {
			if b.sendCases[i].Chan.Equal(c.Chan) {
				b.sendCases = append(b.sendCases[:i], b.sendCases[i+1:]...)
				break
			}
		}
	}
}

// Chan returns a channel that can be used to send values to all subscribers.
func (b *Broadcaster[T]) Chan() chan<- T {
	return b.valCh
}

// Close closes the Broadcaster and all its subscribers.
// It is safe to call Close multiple times.
func (b *Broadcaster[T]) Close() {
	defer func() {
		if r := recover(); r != nil {
			// Already closed. So, ignore it.
			return
		}
	}()
	close(b.valCh)
}

// Subscribe subscribes to the Broadcaster.
// It returns an error if the Broadcaster is already closed.
// It returns error if the Broadcaster is already closed.
// Otherwise, it returns nil.
func (b *Broadcaster[T]) Subscribe(ch chan<- T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("can't Subscribe to Broadcaster: %w", r.(error))
		}
	}()
	b.subCh <- ch
	return nil
}
