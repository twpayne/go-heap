package prioritychannel

// FIXME add min heap size option
// FIXME add max heap size option

import "github.com/twpayne/go-heap"

// A PriorityChannelOption sets an option on a PriorityChannel.
//
//nolint:revive
type PriorityChannelOption[T any] func(*PriorityChannel[T])

// WithLessFunc sets the lessFunc on a PriorityChannel.
func WithLessFunc[T any](lessFunc heap.LessFunc[T]) PriorityChannelOption[T] {
	return func(c *PriorityChannel[T]) {
		c.lessFunc = lessFunc
	}
}

// A PriorityChannel copies values from one channel to another, prioritizing
// values according to a LessFunc. Lower values have priority.
type PriorityChannel[T any] struct {
	lessFunc heap.LessFunc[T]
}

// NewPriorityChannel returns a new PriorityChannel with the given options.
func NewPriorityChannel[T any](options ...PriorityChannelOption[T]) *PriorityChannel[T] {
	var c PriorityChannel[T]
	for _, option := range options {
		option(&c)
	}
	return &c
}

// Run copies values from destCh to sourceCh, prioritizing the values according
// to c's LessFunc. It returns after the last value received from sourceCh is
// sent to destCh.
func (c *PriorityChannel[T]) Run(destCh chan<- T, sourceCh <-chan T) {
	heap := heap.NewHeap(c.lessFunc)

	var valueToSend T
	valueToSendValid := false

	for {
		// If we do not already have a value to send, get one. If the heap is
		// empty then read one from sourceCh, otherwise chose the highest
		// priority value from heap.
		if !valueToSendValid {
			if heap.Empty() {
				var ok bool
				valueToSend, ok = <-sourceCh
				if !ok {
					// sourceCh was closed so we are done.
					return
				}
			} else {
				valueToSend = heap.MustPop()
			}
			valueToSendValid = true //nolint:wastedassign
		}

		// Either send valueToSend to destCh or read a new value from sourceCh
		// and update valueToSend.
		select {
		case destCh <- valueToSend:
			// As valueToSend was sent, we need a new one.
			valueToSendValid = false
		case value, ok := <-sourceCh:
			// As valueToSend was not sent, push it back onto the heap.
			heap.Push(valueToSend)

			// If sourceCh was closed then send the remaining values to destCh
			// and return.
			if !ok {
				for _, value := range heap.PopAll() {
					destCh <- value
				}
				return
			}

			// Otherwise, add value to the heap and get the new value to send.
			valueToSend = heap.PushPop(value)
			valueToSendValid = true
		}
	}
}
