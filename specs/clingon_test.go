package specs

import (
	"⚛sdl"
	"testing"
	pt "prettytest"
	"clingon"
)

type specsSuite struct {
	pt.Suite
}

func (s *specsSuite) beforeAll() {
	initTest()
}

func (s *specsSuite) afterAll() {
	terminateRendering <- true
	sdl.Quit()
}

func (s *specsSuite) should_receive_commands() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "", 2e8)}))
	s.True(Interact([]Interactor{NewEnterCommand(console, "Hello gopher!", 2e8)}))
}

func (s *specsSuite) should_browse_history() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "foo", 2e8)}))
	s.True(Interact([]Interactor{NewEnterCommand(console, "bar", 2e8)}))
	s.True(Interact([]Interactor{NewEnterCommand(console, "biz", 2e8)}))
	s.True(Interact([]Interactor{NewBrowseHistory(console,
		[]int{
			clingon.HISTORY_PREV,
			clingon.HISTORY_PREV,
			clingon.HISTORY_PREV,
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
		},
		2e8)}))
}

func (s *specsSuite) should_move_cursor() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "fooooooooooooooooo!!!!", 2e8)}))
	s.True(Interact([]Interactor{NewBrowseHistory(console, []int{clingon.HISTORY_PREV}, 2e8)}))
	s.True(Interact([]Interactor{NewMoveCursor(console, []int{
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
	},
		2e8)}))
}

func (s *specsSuite) should_slidedown_slideup() {
	slideDown.Start()
	s.True(<-slideDown.FinishedCh())
	slideUp.Start()
	s.True(<-slideUp.FinishedCh())
}

func TestConsoleSpecs(t *testing.T) {
	pt.RunWithFormatter(
		t,
		&pt.BDDFormatter{"Console"},
		new(specsSuite),
	)
}
