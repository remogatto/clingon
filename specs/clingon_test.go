package specs

import (
	"testing"
	pt "spectrum/prettytest"
	"clingon"
)

func should_receive_commands(t *pt.T) {
	t.True(Interact([]Interactor{NewEnterCommand(console, "Hello gopher!", 2e8)}))
}

func should_browse_history(t *pt.T) {
	t.True(Interact([]Interactor{NewEnterCommand(console, "foo", 2e8)}))
	t.True(Interact([]Interactor{NewEnterCommand(console, "bar", 2e8)}))
	t.True(Interact([]Interactor{NewEnterCommand(console, "biz", 2e8)}))
	t.True(Interact([]Interactor{NewBrowseHistory(console, 
			[]int{
			clingon.HISTORY_PREV, 
			clingon.HISTORY_PREV, 
			clingon.HISTORY_PREV, 
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
		}, 2e8)}))
}

func should_move_cursor(t *pt.T) {
	t.True(Interact([]Interactor{NewEnterCommand(console, "fooooooooooooooooo!!!!", 2e8)}))
	t.True(Interact([]Interactor{NewBrowseHistory(console, []int{clingon.HISTORY_PREV}, 2e8)}))
	t.True(Interact([]Interactor{NewMoveCursor(console, []int{
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
		}, 2e8)}))
}

func TestConsoleSpecs(t *testing.T) {
	pt.Describe(
		t,
		"Console",
		should_receive_commands,
		should_browse_history,
		should_move_cursor,

		beforeAll,
		afterAll,
	)
}
