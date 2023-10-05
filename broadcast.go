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

// WithTimeout sets a timeout for broadcasting.
// If a subscriber is blocked for the timeout, it will be removed from the Broadcaster.
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
	// A channel maybe already closed. So, ignore panics.
	ignorePanic(func() { close(b.subCh) })
	for _, c := range b.sendCases {
		ignorePanic(func() { c.Chan.Close() })
	}
}

func ignorePanic(f func()) {
	defer func() {
		if r := recover(); r != nil {
			// Ignore.
		}
	}()
	f()
}

func (b *Broadcaster[T]) broadcast(v T) {
	val := reflect.ValueOf(v)
	for i := range b.sendCases {
		b.sendCases[i].Send = val
	}

	timeout := time.NewTimer(b.timeout)
	defer timeout.Stop()

	// Construct select cases like this:
	// select {
	// case <-timeout.C:
	// case ch1 <- v:
	// case ch2 <- v:
	// ...
	// }
	cases := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(timeout.C)}}
	cases = append(cases, b.sendCases...)

	for len(cases) > 1 {
		timeout := false
		cases, timeout = b.doSelect(cases)
		if timeout {
			break
		}
	}
	for _, c := range cases[1:] {
		// If there are remaining cases, they are timed out.
		// So, remove them from the broadcaster.
		for i := range b.sendCases {
			if b.sendCases[i].Chan.Equal(c.Chan) {
				b.sendCases = append(b.sendCases[:i], b.sendCases[i+1:]...)
				ignorePanic(func() { c.Chan.Close() })
				break
			}
		}
	}
}

func (b *Broadcaster[T]) doSelect(c []reflect.SelectCase) (cases []reflect.SelectCase, timeout bool) {
	cases = c
	defer func() {
		if r := recover(); r != nil {
			// If you are here, reflect.Select below paniced because one of channels is closed.
			// However, which we don't know which one is closed. So, try one by one.
			for i := 1; i < len(cases); {
				switch trySend(cases[i]) {
				case trySendDone:
					cases = append(cases[:i], cases[i+1:]...)
				case trySendBlocked:
					i++
				case trySendClosed:
					// Remove the channel from the Broadcaster.
					for j := range b.sendCases {
						if b.sendCases[j].Chan.Equal(cases[i].Chan) {
							b.sendCases = append(b.sendCases[:j], b.sendCases[j+1:]...)
							break
						}
					}
					cases = append(cases[:i], cases[i+1:]...)
				}
			}
		}
	}()
	// Wait for one of the cases to be ready.
	chosen, _, _ := reflect.Select(cases)
	if chosen == 0 {
		// Timeout.
		return cases, true
	}
	// Remove the case that was done.
	cases = append(cases[:chosen], cases[chosen+1:]...)
	return cases, false
}

type trySendResult int

const (
	trySendBlocked trySendResult = iota
	trySendDone
	trySendClosed
)

func trySend(c reflect.SelectCase) (ret trySendResult) {
	defer func() {
		if r := recover(); r != nil {
			// If you are here, reflect.Select paniced because, that is, the channel is closed.
			ret = trySendClosed
		}
	}()
	chesen, _, _ := reflect.Select([]reflect.SelectCase{c, {Dir: reflect.SelectDefault}})
	if chesen == 0 {
		return trySendDone
	}
	return trySendBlocked
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
