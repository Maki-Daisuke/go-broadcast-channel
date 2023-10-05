# go-broadcast-channel

Broadcast values to multiple channels

## Usage

```golang
	b := broadcast.New[int](0).WithTimeout(1 * time.Second)
	defer b.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Create two subscribers.
	ch1 := make(chan int)
	b.Subscribe(ch1)
	go func() {
		for v := range ch1 {
			// Do something.
		}
		wg.Done()
	}()

	ch2 := make(chan int)
	b.Subscribe(ch2)
	go func() {
		for v := range ch2 {
			// Do another thing.
		}
		wg.Done()
	}()

	// Send two value.
	b.Chan() <- 1
	b.Chan() <- 2
	b.Close()

	wg.Wait()
```

## Design Goals

### Type safe

Use generics.

### Safe with stalling goroutine

"Naive" implementations of broadcast use loop of send statement like this:

```golang
for _, c := range subscribers {
	c <- v
}
```

This implementation blocks entire broadcasting when one of the receivers stalls.

To avoid that, use `select` statement and add (optional) timeout feature.

### Channel as Interface

Use channel both to send and to receive values, so that `select` 
statement can be used for non-blocking communication.

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/Maki-Daisuke/go-broadcast-channel).

## Author

Daisuke (yet another) Maki

## LICENSE

MIT License
