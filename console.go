package console

import (
	"container/vector"
)

type Renderer interface {
	UpdateCh() chan *Console
}

type Console struct {
	lines *vector.StringVector

	// the content of the current line as a string slice
	currLine *vector.StringVector

	// Cursor position on current line
	cursorPosition int

	// prompt string (empty by default)
	prompt string

	// Interface
	renderer Renderer

	// Channels
	charCh chan string
	backspaceCh, returnCh, renderingDone chan bool
}

func NewConsole(renderer Renderer) *Console {
	console :=  &Console{

	lines: new(vector.StringVector),
	currLine: new(vector.StringVector),
	charCh: make(chan string),
	backspaceCh: make(chan bool),
	returnCh: make(chan bool),
	renderingDone: make(chan bool),
	renderer: renderer,

	}

	go console.loop()

	return console
}

// Push a new string on the current line
func (console *Console) Push(str string) {
	console.currLine.Push(str)
	console.cursorPosition += len(str)
}

// Push a carriage-return
func (console *Console) Return() {
	console.lines.Push(console.GetCurrentLine())
	// Clear the current line
	console.currLine = new(vector.StringVector)
	console.cursorPosition = 0
}

// Pop a character from the current line
func (console *Console) Pop() {
	console.currLine.Pop()
	console.cursorPosition--
}

// Return the character channel (receive only)
func (console *Console) CharCh() chan<- string {
	return console.charCh
}

// Return the return channel (receive only)
func (console *Console) ReturnCh() chan<- bool {
	return console.returnCh
}

// Return the current line string
func (console *Console) GetCurrentLine() string {
	var currLine string

	if console.prompt != "" {
		currLine = console.prompt + " "
	}

	for _, str := range *console.currLine {
		currLine += str
	}

	return currLine
}

// Set the prompt string
func (console *Console) SetPrompt(prompt string) {
	console.prompt = prompt
}

func (console *Console) loop() {
	for {
		select {
		case char := <-console.charCh:
			console.Push(char)
			console.renderer.UpdateCh() <- console

		case <-console.backspaceCh:
			console.Pop()

		case <-console.returnCh:
			console.Return()
		}
	}
}
