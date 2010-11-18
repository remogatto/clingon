package console

import (
	"testing"
	pt "spectrum/prettytest"
)

// Renderer mock
type testRenderer struct {
	updateCh chan *Console
}

func (renderer *testRenderer) UpdateCh() chan *Console { return renderer.updateCh }

func (renderer *testRenderer) loop() { 
	for {
		select {
		case <-renderer.updateCh: break
		}
	}
}

func NewTestRenderer() *testRenderer {
	renderer := &testRenderer{

	updateCh: make(chan *Console),

	}

	go renderer.loop()
	
	return renderer
}

func beforeConsole(t *pt.T) {
	console = NewConsole(NewTestRenderer())
}

func testInit(t *pt.T) {
	console := NewConsole(NewTestRenderer())
	t.NotNil(console)
	t.NotNil(console.lines)
}

func testPush(t *pt.T) {
	console.Push("a")
	console.Push("b")
	console.Push("c")

	t.Equal("a", console.currLine.At(0))
	t.Equal("b", console.currLine.At(1))
	t.Equal("c", console.currLine.At(2))
	t.Equal(3, console.cursorPosition)
}

func testGetCurrentLine(t *pt.T) {
	console.Push("a")
	console.Push("b")
	console.Push("c")

	t.Equal("abc", console.GetCurrentLine())
}

func testSetPrompt(t *pt.T) {
	console.SetPrompt("console>")
	console.Push("a")
	console.Push("b")
	console.Push("c")

	t.Equal("console> abc", console.GetCurrentLine())
}

func TestConsole(t *testing.T) {
	pt.Run(
		t,
		testInit,
		testPush,
		testGetCurrentLine,
		testSetPrompt,

		beforeConsole,
	)
}


