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

func (s *consoleTestSuite) testPushUnicode() {
	console.PushUnicode(uint16([]int("a")[0]))
	console.PushUnicode(uint16([]int("b")[0]))
	console.PushUnicode(uint16([]int("c")[0]))
	s.Equal("a", console.commandLine.content.At(0))
	s.Equal("b", console.commandLine.content.At(1))
	s.Equal("c", console.commandLine.content.At(2))
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) testPrint() {
	console.Print("foo\nbar\nbaz")
	s.Equal("foo", console.lines.At(0))
	s.Equal("bar", console.lines.At(1))
	s.Equal("baz", console.lines.At(2))
	s.Equal("console> ", console.commandLine.toString())
}

func (s *consoleTestSuite) testPushString() {
	console.PushString("foo")
	s.Equal("console> foo", console.commandLine.toString())
}

func (s *consoleTestSuite) testPushReadline() {
	console.PushString("foo")
	console.carriageReturn()
	console.PushString("bar")
	console.carriageReturn()
	console.PushString("biz")
	console.carriageReturn()

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> bar", console.commandLine.toString())

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.PushReadline(HISTORY_NEXT)
	s.Equal("console> bar", console.commandLine.toString())

	console.PushReadline(HISTORY_NEXT)
	s.Equal("console> biz", console.commandLine.toString())

	console.PushReadline(HISTORY_NEXT)
	s.Equal("console> ", console.commandLine.toString())

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.PushString("bar")
	console.carriageReturn()

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> bizbar", console.commandLine.toString())

	console.PushReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())
}

func (s *consoleTestSuite) testMoveCursor() {
	console.PushString("bar")

	console.PushReadline(CURSOR_LEFT)
	s.Equal(2, console.commandLine.cursorPosition)

	console.PushReadline(CURSOR_RIGHT)
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) testCommandline() {
	console.PushString("bar")
	s.Equal("console> bar", console.Commandline())
}

func (s *consoleTestSuite) testSetPrompt() {
	console.SetPrompt("foo> ")
	console.PushString("abc")
	s.Equal("foo> abc", console.commandLine.toString())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		new(consoleTestSuite),
	)
}
