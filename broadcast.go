package broadcastchannel

import (
	"fmt"
)

type Broadcaster[T any] struct {
	valCh  chan T
	subCh  chan chan<- T
	outChs []chan<- T
}

// New creates a new Broadcaster.
// It takes a buffer size for the channel.
func New[T any](n int) *Broadcaster[T] {
	b := &Broadcaster[T]{
		valCh: make(chan T, n),
		subCh: make(chan chan<- T),
	}
	go b.run()
	return b
}

func (b *Broadcaster[T]) run() {
	for {
		select {
		case v, ok := <-b.valCh:
			if !ok {
				close(b.subCh)
				for _, outCh := range b.outChs {
					close(outCh)
				}
				return
			}
			for _, outCh := range b.outChs {
				outCh <- v
			}
		case outCh := <-b.subCh:
			b.outChs = append(b.outChs, outCh)
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
