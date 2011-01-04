package clingon

import (
	"testing"
	pt "prettytest"
)

var console *Console

// Renderer mock
type testRenderer struct {
	dummyEventCh chan interface{}
}

func (renderer *testRenderer) EventCh() chan<- interface{} { return renderer.dummyEventCh }

func NewTestRenderer() *testRenderer {
	r := &testRenderer{make(chan interface{})}
	go func() {
		for {
			select {
			case <-r.dummyEventCh:
			}
		}
	}()
	return r
}

type consoleTestSuite struct {
	pt.Suite
}

func (s *consoleTestSuite) before() {
	console = NewConsole(NewTestRenderer(), nil)
}

func (s *consoleTestSuite) testInit() {
	console := NewConsole(NewTestRenderer(), nil)
	s.NotNil(console)
	s.NotNil(console.lines)
}

func (s *consoleTestSuite) testCharCh() {
	done := make(chan bool)

	console.CharCh() <- SendChar{uint16([]int("a")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("b")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("c")[0]), done}
	<-done

	s.Equal("a", console.commandLine.content.At(0))
	s.Equal("b", console.commandLine.content.At(1))
	s.Equal("c", console.commandLine.content.At(2))
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) testLinesCh() {
	done := make(chan bool)
	lines := "foo\nbar\nbaz"

	console.LinesCh() <- SendLines{lines, done}
	<-done

	s.Equal("foo", console.lines.At(0))
	s.Equal("bar", console.lines.At(1))
	s.Equal("baz", console.lines.At(2))
	s.Equal("console> ", console.commandLine.toString())
}

func (s *consoleTestSuite) testHistoryCh() {
	console.commandLine.Insert("f")
	console.commandLine.Insert("o")
	console.commandLine.Insert("o")
	console.Return()

	console.commandLine.Insert("b")
	console.commandLine.Insert("a")
	console.commandLine.Insert("r")
	console.Return()

	console.commandLine.Insert("b")
	console.commandLine.Insert("i")
	console.commandLine.Insert("z")
	console.Return()

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> bar", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> bar", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> biz", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_NEXT)
	s.Equal("console> ", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.commandLine.Insert("b")
	console.commandLine.Insert("a")
	console.commandLine.Insert("r")
	console.Return()

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> bizbar", console.commandLine.toString())

	console.commandLine.BrowseHistory(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())
}

func (s *consoleTestSuite) testMoveCursor() {
	console.commandLine.Insert("b")
	console.commandLine.Insert("a")
	console.commandLine.Insert("r")

	console.commandLine.MoveCursor(CURSOR_LEFT)
	s.Equal(2, console.commandLine.cursorPosition)

	console.commandLine.MoveCursor(CURSOR_RIGHT)
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) testGetcommandLine() {
	done := make(chan bool)

	console.CharCh() <- SendChar{uint16([]int("a")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("b")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("c")[0]), done}
	<-done

	s.Equal("console> abc", console.commandLine.toString())
}

func (s *consoleTestSuite) testSetPrompt() {
	console.SetPrompt("foo> ")

	done := make(chan bool)

	console.CharCh() <- SendChar{uint16([]int("a")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("b")[0]), done}
	<-done

	console.CharCh() <- SendChar{uint16([]int("c")[0]), done}
	<-done

	s.Equal("foo> abc", console.commandLine.toString())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		new(consoleTestSuite),
	)
}
