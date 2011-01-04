package clingon

import (
	"os"
	"time"
	"strings"
	"container/vector"
)

const (
	CURSOR_BLINK_TIME = 5e8

	HISTORY_NEXT = iota
	HISTORY_PREV
	CURSOR_LEFT
	CURSOR_RIGHT
)

type Evaluator interface {
	Run(console *Console, command string) os.Error
}

type UpdateConsoleEvent struct {
	console *Console
}

type InsertCharEvent struct {
	console *Console
	char    uint16
}

type BackspaceEvent struct {
	console *Console
}

type NewlineEvent struct {
	console *Console
}
 
type MoveCursorEvent struct {
	console        *Console
	cursorPosition int
}

type UpdateCursorEvent struct {
	commandLine *commandLine
	enabled     bool
}

type UpdateCommandLineEvent struct {
	console *Console
	commandLine *commandLine
}

type SendChar struct {
	Char uint16
	Done chan bool
}

type SendLines struct {
	Lines string
	Done  chan bool
}

type SendReadlineCommand struct {
	Command int
	Done    chan bool
}

type Renderer interface {
	// Receive Event objects from the console
	EventCh() chan<- interface{}
}

type commandLine struct {
	prompt string

	// the content of the prompt line as a string vector
	content *vector.StringVector

	// the history of the command line
	history *vector.Vector

	// Cursor position on current line
	cursorPosition int

	historyCount int
}

func newCommandLine(prompt string) *commandLine {
	return &commandLine{
		prompt:         prompt,
		content:        new(vector.StringVector),
		history:        new(vector.Vector),
		cursorPosition: 0,
		historyCount:   0,
	}
}

// Check if the command line is empty
func (commandLine *commandLine) empty() bool {
	return len(commandLine.toString()) == 0
}

// Insert a string in the prompt line
func (commandLine *commandLine) insertChar(str string) {
	commandLine.content.Insert(commandLine.cursorPosition, str)
	commandLine.incCursorPosition(len(str))
}

// Delete the character on the left of the current cursor position
func (commandLine *commandLine) backSpace() {
	commandLine.decCursorPosition(1)
	if commandLine.content.Len() > 0 {
		commandLine.content.Delete(commandLine.cursorPosition)
	}
}

// Clear the prompt line
func (commandLine *commandLine) clear() {
	commandLine.content = new(vector.StringVector)
	commandLine.cursorPosition = 0
	commandLine.historyCount = commandLine.history.Len()
}

// Browse history on the command line
func (commandLine *commandLine) browseHistory(direction int) {
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
		commandLine.clear()
		return
	}
	newContent := commandLine.history.At(commandLine.historyCount).(*vector.StringVector).Copy()
	commandLine.content = &newContent
	commandLine.incCursorPosition(len(commandLine.toString()))
}

// Push current command line on the history and reset the history counter
func (commandLine *commandLine) push() string {
	if commandLine.toString() != "" && commandLine.notInHistory(commandLine.toString()) {
		commandLine.history.Push(commandLine.content)
	}
	line := commandLine.toString()
	commandLine.clear()
	return line
}

// Return the current prompt line as a single string (including prompt)
func (commandLine *commandLine) toString() string {
	return commandLine.prompt + commandLine.stringVectorToString(commandLine.content)
}

// Move the cursor left/right on the command line
func (commandLine *commandLine) moveCursor(dir int) {
	if dir == CURSOR_LEFT {
		commandLine.decCursorPosition(1)
	} else {
		commandLine.incCursorPosition(1)
	}
}

func (commandLine *commandLine) stringVectorToString(v *vector.StringVector) string {
	var currLine string
	for _, str := range *v {
		currLine += str
	}
	return currLine
}

func (commandLine *commandLine) decCursorPosition(dec int) {
	commandLine.cursorPosition -= dec
	if commandLine.cursorPosition < 0 {
		commandLine.cursorPosition = 0
	}
}

func (commandLine *commandLine) incCursorPosition(inc int) {
	commandLine.cursorPosition += inc
	if commandLine.cursorPosition > commandLine.content.Len() {
		commandLine.cursorPosition = commandLine.content.Len()
	}
}

func (commandLine *commandLine) notInHistory(line string) bool {
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
	// Pause/unpause the console
	Paused bool
	// Greeting message
	GreetingText string

	// An instance of the command-line
	commandLine *commandLine

	lines *vector.StringVector

	// Interfaces
	renderer  Renderer
	evaluator Evaluator

	// Channels
	charCh     chan SendChar
	linesCh    chan SendLines
	readlineCh chan SendReadlineCommand
}

// Initialize a new console object. Renderer and Evaluator objects can
// be nil.
func NewConsole(renderer Renderer, evaluator Evaluator) *Console {
	console := &Console{
		lines:       new(vector.StringVector),
		commandLine: newCommandLine("console> "),
		charCh:      make(chan SendChar),
		linesCh:     make(chan SendLines),
		readlineCh:  make(chan SendReadlineCommand),
		renderer:    renderer,
		evaluator:   evaluator,
	}
	go console.loop()
	return console
}

// Set the prompt string
func (console *Console) SetPrompt(prompt string) {
	console.commandLine.prompt = prompt
}

// Push the current commandline. Return it has a string.
func (console *Console) Return() string {
	commandLine := console.commandLine.push()
	console.lines.Push(console.commandLine.prompt + commandLine)
	return commandLine
}

// Print a string on the console
func (console *Console) Print(str string) {
	if str != "" {
		console.PushLines(strings.Split(str, "\n", -1))
	}
}

// Push lines of text
func (console *Console) PushLines(lines []string) {
	for _, line := range lines {
		console.lines.Push(line)
	}
}

// Get the current command line
func (console *Console) CommandLine() string {
	return console.commandLine.toString()
}

// Return the character channel (receive only)
func (console *Console) CharCh() chan<- SendChar {
	return console.charCh
}

// Return the lines channel (receive only).
func (console *Console) LinesCh() chan<- SendLines {
	return console.linesCh
}

// Client code can send readline-like command to this channel
// (e.g. history browsing, cursor movements, etc.). Readline's
// commands emulation is incomplete.
func (console *Console) ReadlineCh() chan<- SendReadlineCommand {
	return console.readlineCh
}

func (console *Console) loop() {
	var toggleCursor bool
	ticker := time.NewTicker(CURSOR_BLINK_TIME)

	// Render the prompt and the greeting text before starting the
	// loop
	if console.renderer != nil {
		console.Print(console.GreetingText)
		console.renderer.EventCh() <- UpdateConsoleEvent{console}
	}

	for {
		if !console.Paused {
			select {
			case receivedChar := <-console.charCh:
				switch receivedChar.Char {
				case 0x0008: // BACKSPACE
					console.commandLine.backSpace()
					if console.renderer != nil {
						console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
					}
				case 0x000d: // RETURN
					command := console.Return()
					if console.evaluator != nil {
						console.evaluator.Run(console, command)
					}
					if console.renderer != nil {
						console.renderer.EventCh() <- UpdateConsoleEvent{console}
					}
				default:
					console.commandLine.insertChar(string(receivedChar.Char))
					if console.renderer != nil {
						console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
					}
				}

				receivedChar.Done <- true

			case receivedLines := <-console.linesCh: // Receive lines of text
				lines := strings.Split(receivedLines.Lines, "\n", -1)
				console.PushLines(lines)
				if console.renderer != nil {
					console.renderer.EventCh() <- UpdateConsoleEvent{console}
				}
				receivedLines.Done <- true

			case receivedReadlineCmd := <-console.readlineCh: // Browse history
				switch receivedReadlineCmd.Command {
				case HISTORY_NEXT, HISTORY_PREV:
					console.commandLine.browseHistory(receivedReadlineCmd.Command)

				case CURSOR_LEFT, CURSOR_RIGHT:
					console.commandLine.moveCursor(receivedReadlineCmd.Command)
				}
				if console.renderer != nil {
					console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
				}

				receivedReadlineCmd.Done <- true

			case <-ticker.C: // Blink cursor
				if console.renderer != nil {
					console.renderer.EventCh() <- UpdateCursorEvent{console.commandLine, toggleCursor}
				}
				toggleCursor = !toggleCursor
			}
		} else {
			select {
			case <-console.charCh:
			case <-console.linesCh:
			case <-console.readlineCh:
			case <-ticker.C:
			}

		}
	}
}
