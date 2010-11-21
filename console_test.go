package cli

import (
	"testing"
	pt "spectrum/prettytest"
)

var console *Console

// Renderer mock
type testRenderer struct {}

func (renderer *testRenderer) RenderCommandLine(commandLine *CommandLine) {}
func (renderer *testRenderer) RenderVisibleLines(console *Console) {}
func (renderer *testRenderer) EnableCursor(enable bool) {}

func NewTestRenderer() *testRenderer {
	return &testRenderer{}
}

func before(t *pt.T) {
	console = NewConsole(NewTestRenderer(), nil)
}

func testInit(t *pt.T) {
	console := NewConsole(NewTestRenderer(), nil)
	t.NotNil(console)
	t.NotNil(console.lines)
}

func testCharCh(t *pt.T) {
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	t.Equal("a", console.CommandLine.content.At(0))
	t.Equal("b", console.CommandLine.content.At(1))
	t.Equal("c", console.CommandLine.content.At(2))
	t.Equal(3, console.CommandLine.cursorPosition)
}

func testLinesCh(t *pt.T) {
	lines := "foo\nbar\nbaz"
	console.LinesCh() <- lines
	
	t.Equal("foo", console.lines.At(0))
	t.Equal("bar", console.lines.At(1))
	t.Equal("baz", console.lines.At(2))
	t.Equal("console> ", console.CommandLine.String())
}

func testHistoryCh(t *pt.T) {
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

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> bar", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> foo", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> foo", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_NEXT)
	t.Equal("console> bar", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_NEXT)
	t.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_NEXT)
	t.Equal("console> ", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> biz", console.CommandLine.String())

	console.CommandLine.Insert("b")
	console.CommandLine.Insert("a")
	console.CommandLine.Insert("r")
	console.Return()

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> bizbar", console.CommandLine.String())

	console.CommandLine.CycleHistory(HISTORY_PREV)
	t.Equal("console> biz", console.CommandLine.String())
}

func testMoveCursor(t *pt.T) {
	console.CommandLine.Insert("b")
	console.CommandLine.Insert("a")
	console.CommandLine.Insert("r")

	console.CommandLine.MoveCursor(CURSOR_LEFT)
	t.Equal(2, console.CommandLine.cursorPosition)

	console.CommandLine.MoveCursor(CURSOR_RIGHT)
	t.Equal(3, console.CommandLine.cursorPosition)
}

func testGetCommandLine(t *pt.T) {
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	t.Equal("console> abc", console.CommandLine.String())
}

func testSetPrompt(t *pt.T) {
	console.SetPrompt("foo> ")
	console.CharCh() <- uint16([]int("a")[0])
	console.CharCh() <- uint16([]int("b")[0])
	console.CharCh() <- uint16([]int("c")[0])

	t.Equal("foo> abc", console.CommandLine.String())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		testInit,
		testCharCh,
		testLinesCh,
		testHistoryCh,
		testMoveCursor,
		testGetCommandLine,
		testSetPrompt,
		
		before,
	)
}


