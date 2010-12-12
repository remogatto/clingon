package clingon

import (
	"testing"
	pt "prettytest"
)

var console *Console

// Renderer mock
type testRenderer struct {
	dummyConsoleCh chan *Console
	dummyBoolCh chan bool
}

func (renderer *testRenderer) RenderCommandLineCh() chan<- *Console { return renderer.dummyConsoleCh }
func (renderer *testRenderer) RenderConsoleCh() chan<- *Console { return renderer.dummyConsoleCh }
func (renderer *testRenderer) RenderCursorCh() chan<- *Console { return renderer.dummyConsoleCh }
func (renderer *testRenderer) EnableCursorCh() chan<- bool { return renderer.dummyBoolCh }

func NewTestRenderer() *testRenderer {
	r := &testRenderer{make(chan *Console), make(chan bool)}
	go func() {
		for {
			select {
			case <-r.dummyConsoleCh:
			case <-r.dummyBoolCh:
			}
		}
	}()
	return r
}

type consoleTestSuite struct { pt.Suite }

func (s *consoleTestSuite) before() {
	console = NewConsole(NewTestRenderer(), nil)
}

func (s *consoleTestSuite) testInit() {
	console := NewConsole(NewTestRenderer(), nil)
	s.NotNil(console)
	s.NotNil(console.lines)
}

func (s *consoleTestSuite) testCharCh() {
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	s.Equal("a", console.CommandLine.content.At(0))
	s.Equal("b", console.CommandLine.content.At(1))
	s.Equal("c", console.CommandLine.content.At(2))
	s.Equal(3, console.CommandLine.cursorPosition)
}

func (s *consoleTestSuite) testLinesCh() {
	lines := "foo\nbar\nbaz"
	console.LinesCh() <- lines
	
	s.Equal("foo", console.lines.At(0))
	s.Equal("bar", console.lines.At(1))
	s.Equal("baz", console.lines.At(2))
	s.Equal("console> ", console.CommandLine.String())
}

func (s *consoleTestSuite) testHistoryCh() {
	console.CommandLine.Insert("f")
	console.CommandLine.Insert("o")
	console.CommandLine.Insert("o")
	console.Return()

	console.CommandLine.Insert("b")
	console.CommandLine.Insert("a")
	console.CommandLine.Insert("r")
	console.Return()

	console.CommandLine.Insert("b")
	console.CommandLine.Insert("i")
	console.CommandLine.Insert("z")
	console.Return()

	console.CommandLine.Insert("b")
	console.CommandLine.Insert("i")
	console.CommandLine.Insert("z")
	console.Return()

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> bar", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> foo", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> foo", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> bar", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> ", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.Insert("b")
	console.CommandLine.Insert("a")
	console.CommandLine.Insert("r")
	console.Return()

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> bizbar", console.CommandLine.String())

	console.CommandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.CommandLine.String())
}

func (s *consoleTestSuite) testMoveCursor() {
	console.CommandLine.Insert("b")
	console.CommandLine.Insert("a")
	console.CommandLine.Insert("r")

	console.CommandLine.MoveCursor(CURSOR_LEFT)
	s.Equal(2, console.CommandLine.cursorPosition)

	console.CommandLine.MoveCursor(CURSOR_RIGHT)
	s.Equal(3, console.CommandLine.cursorPosition)
}

func (s *consoleTestSuite) testGetCommandLine() {
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	s.Equal("console> abc", console.CommandLine.String())
}

func (s *consoleTestSuite) testSetPrompt() {
	console.SetPrompt("foo> ")
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	s.Equal("foo> abc", console.CommandLine.String())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		new(consoleTestSuite),
	)
}


