package specs

import (
	"testing"
	pt "spectrum/prettytest"
	"time"
)

func should_receive_chars(t *pt.T) {
	str := "Hello gopher!"
	count := 0
	ticker := time.NewTicker(1e8)

	for {
		select {
		case <-ticker.C:
			if count == len(str) {
				goto finish
			}
			console.CharCh() <- string(str[count])
			count++
			render()
		}
	}

finish:
	t.True(true)
}

func should_receive_return(t *pt.T) {
	str := "Hello gopher!"
	count := 0
	ticker := time.NewTicker(1e8)

	for {
		select {
		case <-ticker.C:
			if count == len(str) {
				goto finish
			}
			console.CharCh() <- string(str[count])
			count++
			render()
		}
	}

finish:
	console.ReturnCh() <- true
	console.ReturnCh() <- true
	console.ReturnCh() <- true

	render()

	for {}
	t.True(true)
}

func TestConsoleSpecs(t *testing.T) {
	pt.Describe(
		t,
		"Console",
//		should_receive_chars,
		should_receive_return,

		before,
		after,
	)
}
