package heap

// FIXME add channel sizes (min and max pending)

// PriorityChannel copies values from sourceCh to destCh, prioritizing the
// values according to lessFunc (lesser values are copied first), without any
// limit on buffering. If sourceCh is closed then it writes all remaining values
// to destCh, closes destCh, and returns.
func PriorityChannel[T any](destCh chan<- T, sourceCh <-chan T, lessFunc func(T, T) bool) {
	defer close(destCh)
	heap := NewHeap(lessFunc)

	var valueToSend T
	valueToSendValid := false

	for {
		// If we do not already have a value to send, get one. If the heap is
		// empty then read one from sourceCh, otherwise chose the highest
		// priority value from heap.
		if !valueToSendValid {
			var ok bool
			valueToSend, ok = heap.Pop()
			if !ok {
				valueToSend, ok = <-sourceCh
				if !ok {
					// sourceCh was closed so we are done.
					return
				}
			}
			valueToSendValid = true //nolint:wastedassign
		}

		// Either write valueToSend to destCh or read a new value from sourceCh
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
