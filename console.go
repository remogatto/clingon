package console

import (
	"time"
	"strings"
	"container/vector"
)

const (
	CURSOR_BLINK_TIME = 5e8
	HISTORY_NEXT = 1
	HISTORY_PREV = 2
	CURSOR_LEFT = 1
	CURSOR_RIGHT = 2
)

type Evaluator interface {
	Run(command string) string
}

type Renderer interface {
	RenderCommandLine(commandLine *CommandLine)
	RenderVisibleLines(console *Console)
	EnableCursor(enable bool)
}

type CommandLine struct {
	Prompt string

	// the content of the prompt line as a string vector
	content *vector.StringVector

	// the history of the command line
	history *vector.Vector

	// Cursor position on current line
	cursorPosition int

	historyCount int
}

func NewCommandLine(prompt string) *CommandLine {
	return &CommandLine{

	Prompt: prompt, 
	content: new(vector.StringVector), 
	history: new(vector.Vector), 
	cursorPosition: 0, 
	historyCount: 0,

	}
}

// Check if the command line is empty
func (commandLine *CommandLine) Empty() bool {
	return len(commandLine.toString()) == 0
}

// Insert a string in the prompt line
func (commandLine *CommandLine) Insert(str string) {
	commandLine.content.Insert(commandLine.cursorPosition, str)
	commandLine.incCursorPosition(len(str))
}

// Delete the character on the left of the current cursor position
func (commandLine *CommandLine) BackSpace() {
	commandLine.decCursorPosition(1)
	if commandLine.content.Len() > 0 {
		commandLine.content.Delete(commandLine.cursorPosition)
	}
}

// Clear the prompt line
func (commandLine *CommandLine) Clear() {
	commandLine.content = new(vector.StringVector)
	commandLine.cursorPosition = 0
	commandLine.historyCount = commandLine.history.Len()
}

// Cycle history on the command line
func (commandLine *CommandLine) CycleHistory(direction int) {
	if direction == HISTORY_NEXT {
		commandLine.historyCount++
	} else {
		commandLine.historyCount--
	}
	if commandLine.historyCount < 0 {
		commandLine.historyCount = 0
		return
	}
	if commandLine.historyCount >= commandLine.history.Len() {
		commandLine.Clear()
		return
	}
	newContent := commandLine.history.At(commandLine.historyCount).(*vector.StringVector).Copy()
	commandLine.content = &newContent
	commandLine.incCursorPosition(len(commandLine.toString()))
}

// Push current command line on the history and reset the history counter
func (commandLine *CommandLine) Push() string {
	if commandLine.toString() != "" && commandLine.notInHistory(commandLine.toString()) {
		commandLine.history.Push(commandLine.content)
	}
	line := commandLine.toString()
	commandLine.Clear()
	return line
}

// Return the current prompt line as a single string (including prompt)
func (commandLine *CommandLine) String() string {
	return commandLine.Prompt + commandLine.toString()
}

// Move the cursor left/right on the command line
func (commandLine *CommandLine) MoveCursor(dir int) {
	if dir == CURSOR_LEFT {
		commandLine.decCursorPosition(1)
	} else {
		commandLine.incCursorPosition(1)
	}
}

func (commandLine *CommandLine) toString() string {
	return commandLine.stringVectorToString(commandLine.content)
}

func (commandLine *CommandLine) stringVectorToString(v *vector.StringVector) string {
	var currLine string
	for _, str := range *v {
		currLine += str
	}
	return currLine
}

func (commandLine *CommandLine) decCursorPosition(dec int) {
	commandLine.cursorPosition -= dec
	if commandLine.cursorPosition < 0 {
		commandLine.cursorPosition = 0
	}
}

func (commandLine *CommandLine) incCursorPosition(inc int) {
	commandLine.cursorPosition += inc
	if commandLine.cursorPosition > commandLine.content.Len() {
		commandLine.cursorPosition = commandLine.content.Len()
	}
}

func (commandLine *CommandLine) notInHistory(line string) bool {
	for _, v := range *commandLine.history {
		strVector := v.(*vector.StringVector)
		historyEntry := commandLine.stringVectorToString(strVector)
		if line == historyEntry {
			return false 
		}
	}
	return true
}

type Console struct {
	CommandLine *CommandLine

	lines *vector.StringVector

	// Interfaces
	renderer Renderer
	evaluator Evaluator

	// Channels
	charCh chan uint16
	linesCh chan string
	historyCh, cursorCh chan int

	backspaceCh, returnCh, renderingDone chan bool
}

func NewConsole(renderer Renderer, evaluator Evaluator) *Console {
	console :=  &Console{

	lines: new(vector.StringVector),
	CommandLine: NewCommandLine("console> "),
	charCh: make(chan uint16),
	linesCh: make(chan string),
	backspaceCh: make(chan bool),
	returnCh: make(chan bool),
	historyCh: make(chan int),
	cursorCh: make(chan int),
	renderingDone: make(chan bool),
	renderer: renderer,
	evaluator: evaluator,
	}

	go console.loop()

	console.renderer.RenderCommandLine(console.CommandLine)

	return console
}

// Set the prompt string
func (console *Console) SetPrompt(prompt string) {
	console.CommandLine.Prompt = prompt
}

// Push a carriage-return. Return the command string.
func (console *Console) Return() string {
	commandLine := console.CommandLine.Push()
	console.lines.Push(console.CommandLine.Prompt + commandLine)
	return commandLine
}

// Push lines
func (console *Console) PushLines(lines []string) {
	for _, line := range lines {
		console.lines.Push(line)
	}
}

// Get the current command line
func (console *Console) GetCommandLine() string {
	return console.CommandLine.String()
}

// Return the character channel (receive only)
func (console *Console) CharCh() chan<- uint16 {
	return console.charCh
}

// Return the return channel (receive only)
func (console *Console) ReturnCh() chan<- bool {
	return console.returnCh
}

// Return the lines channel (receive only).
func (console *Console) LinesCh() chan<- string {
	return console.linesCh
}

// Return the history channel (receive only)
func (console *Console) HistoryCh() chan<- int {
	return console.historyCh
}

// Return the move cursor channel (receive only). Values sent to this
// channel trigger the cursor movement on the command line.
func (console *Console) CursorCh() chan<- int {
	return console.cursorCh
}

func (console *Console) loop() {
	var toggleCursor bool
	ticker := time.NewTicker(CURSOR_BLINK_TIME)
	for {
		select {
		case char := <-console.charCh:
			switch char {
			case 0x0008: // BACKSPACE
				console.CommandLine.BackSpace()
				console.renderer.EnableCursor(true)
				console.renderer.RenderCommandLine(console.CommandLine)
			case 0x000d: // RETURN
				command := console.Return()
				if console.evaluator != nil && command != "" {
					result := console.evaluator.Run(command)
					lines := strings.Split(result, "\n", -1)
					console.PushLines(lines)
				}
				console.renderer.EnableCursor(true)
				console.renderer.RenderVisibleLines(console)
			default:
				console.CommandLine.Insert(string(char))
				console.renderer.EnableCursor(true)
				console.renderer.RenderCommandLine(console.CommandLine)
			}
 
		case str := <-console.linesCh: // Receive lines of text
			lines := strings.Split(str, "\n", -1)
			console.PushLines(lines)
			console.renderer.RenderVisibleLines(console)
			
		case historyDirection := <-console.historyCh: // Browse history
			console.CommandLine.CycleHistory(historyDirection)
			console.renderer.EnableCursor(true)
			console.renderer.RenderCommandLine(console.CommandLine)

		case cursorDirection := <-console.cursorCh: // Move cursor on the command line
			console.CommandLine.MoveCursor(cursorDirection)
			console.renderer.EnableCursor(true)
			console.renderer.RenderCommandLine(console.CommandLine)

		case <-ticker.C: // Blink cursor
			console.renderer.EnableCursor(toggleCursor)
			console.renderer.RenderCommandLine(console.CommandLine)
			toggleCursor = !toggleCursor

		}
	}
}
