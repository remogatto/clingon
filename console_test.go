package clingon

import (
	pt "github.com/remogatto/prettytest"
	"testing"
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

func (s *consoleTestSuite) Before() {
	console = NewConsole(nil)
	console.SetRenderer(NewTestRenderer())
}

func (s *consoleTestSuite) TestNewConsole() {
	console := NewConsole(nil)
	console.SetRenderer(NewTestRenderer())
	s.Not(s.Nil(console))
	s.Not(s.Nil(console.lines))
}

func (s *consoleTestSuite) TestString() {
	console.PutCommand("foo")
	console.PutCommand("bar")
	s.Equal("console> foo\nconsole> bar\nconsole> ", console.String())
}

func (s *consoleTestSuite) TestPutUnicode() {
	console.PutUnicode('a')
	console.PutUnicode('b')
	console.PutUnicode('c')
	s.Equal("a", console.commandLine.content[0])
	s.Equal("b", console.commandLine.content[1])
	s.Equal("c", console.commandLine.content[2])
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) TestPrint() {
	console.Print("foo\nbar\nbaz")
	s.Equal("foo", console.lines[0])
	s.Equal("bar", console.lines[1])
	s.Equal("baz", console.lines[2])
	s.Equal("console> ", console.commandLine.toString())
}

func (s *consoleTestSuite) TestPrintLines() {
	console.PrintLines([]string{"foo", "bar", "baz"})
	s.Equal("foo", console.lines[0])
	s.Equal("bar", console.lines[1])
	s.Equal("baz", console.lines[2])
	s.Equal("console> ", console.commandLine.toString())
}

func (s *consoleTestSuite) TestPutCommand() {
	console.PutCommand("foo")
	s.Equal("console> foo", console.lines[0])
}

func (s *consoleTestSuite) TestPutString() {
	console.PutString("foo")
	s.Equal("console> foo", console.commandLine.toString())
}

func (s *consoleTestSuite) TestPutReadline() {
	console.PutString("foo")
	console.carriageReturn()
	console.PutString("bar")
	console.carriageReturn()
	console.PutString("biz")
	console.carriageReturn()

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> bar", console.commandLine.toString())

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> foo", console.commandLine.toString())

	console.PutReadline(HISTORY_NEXT)
	s.Equal("console> bar", console.commandLine.toString())

	console.PutReadline(HISTORY_NEXT)
	s.Equal("console> biz", console.commandLine.toString())

	console.PutReadline(HISTORY_NEXT)
	s.Equal("console> ", console.commandLine.toString())

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())

	console.PutString("bar")
	console.carriageReturn()

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> bizbar", console.commandLine.toString())

	console.PutReadline(HISTORY_PREV)
	s.Equal("console> biz", console.commandLine.toString())
}

func (s *consoleTestSuite) TestMoveCursor() {
	console.PutString("bar")

	console.PutReadline(CURSOR_LEFT)
	s.Equal(2, console.commandLine.cursorPosition)

	console.PutReadline(CURSOR_RIGHT)
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) TestCommandline() {
	console.PutString("bar")
	s.Equal("console> bar", console.Commandline())
}

func (s *consoleTestSuite) TestSetPrompt() {
	console.SetPrompt("foo> ")
	console.PutString("abc")
	s.Equal("foo> abc", console.commandLine.toString())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		new(consoleTestSuite),
	)
}
