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
	console = NewConsole(nil)
	console.SetRenderer(NewTestRenderer())
}

func (s *consoleTestSuite) testNewConsole() {
	console := NewConsole(nil)
	console.SetRenderer(NewTestRenderer())
	s.NotNil(console)
	s.NotNil(console.lines)
}

func (s *consoleTestSuite) testString() {
	console.PutCommand("foo")
	console.PutCommand("bar")
	s.Equal("console> foo\nconsole> bar\nconsole> ", console.String())
}

func (s *consoleTestSuite) testPutUnicode() {
	console.PutUnicode(uint16([]int("a")[0]))
	console.PutUnicode(uint16([]int("b")[0]))
	console.PutUnicode(uint16([]int("c")[0]))
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

func (s *consoleTestSuite) testPrintLines() {
	console.PrintLines([]string{"foo", "bar", "baz"})
	s.Equal("foo", console.lines.At(0))
	s.Equal("bar", console.lines.At(1))
	s.Equal("baz", console.lines.At(2))
	s.Equal("console> ", console.commandLine.toString())
}

func (s *consoleTestSuite) testPutCommand() {
	console.PutCommand("foo")
	s.Equal("console> foo", console.lines.At(0))
}

func (s *consoleTestSuite) testPutString() {
	console.PutString("foo")
	s.Equal("console> foo", console.commandLine.toString())
}

func (s *consoleTestSuite) testPutReadline() {
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

func (s *consoleTestSuite) testMoveCursor() {
	console.PutString("bar")

	console.PutReadline(CURSOR_LEFT)
	s.Equal(2, console.commandLine.cursorPosition)

	console.PutReadline(CURSOR_RIGHT)
	s.Equal(3, console.commandLine.cursorPosition)
}

func (s *consoleTestSuite) testCommandline() {
	console.PutString("bar")
	s.Equal("console> bar", console.Commandline())
}

func (s *consoleTestSuite) testSetPrompt() {
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
