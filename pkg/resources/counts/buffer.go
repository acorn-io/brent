package counts

import (
	"time"

	"github.com/acorn-io/brent/pkg/types"
)

func buffer(c chan types.APIEvent) chan types.APIEvent {
	result := make(chan types.APIEvent)
	go func() {
		defer close(result)
		debounce(result, c)
	}()
	return result
}

func debounce(result, input chan types.APIEvent) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	var (
		labrentnt *types.APIEvent
	)
	for {
		select {
		case event, ok := <-input:
			if ok {
				labrentnt = &event
			} else {
				return
			}
		case <-t.C:
			if labrentnt != nil {
				result <- *labrentnt
				labrentnt = nil
			}
		}
	}
}
