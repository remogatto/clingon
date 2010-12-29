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
	commandLine *CommandLine
	enabled     bool
}

type UpdateCommandLineEvent struct {
	commandLine *CommandLine
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

		Prompt:         prompt,
		content:        new(vector.StringVector),
		history:        new(vector.Vector),
		cursorPosition: 0,
		historyCount:   0,
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

// Browse history on the command line
func (commandLine *CommandLine) BrowseHistory(direction int) {
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
	// An instance of the command-line
	CommandLine *CommandLine
	// Pause/unpause the console
	Paused bool
	// Greeting message
	GreetingText string

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
		CommandLine: NewCommandLine("console> "),
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
	console.CommandLine.Prompt = prompt
}

// Push a carriage-return. Return the command string.
func (console *Console) Return() string {
	commandLine := console.CommandLine.Push()
	console.lines.Push(console.CommandLine.Prompt + commandLine)
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
func (console *Console) GetCommandLine() string {
	return console.CommandLine.String()
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
					console.CommandLine.BackSpace()
					if console.renderer != nil {
						console.renderer.EventCh() <- UpdateCommandLineEvent{console.CommandLine}
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
					console.CommandLine.Insert(string(receivedChar.Char))
					if console.renderer != nil {
						console.renderer.EventCh() <- UpdateCommandLineEvent{console.CommandLine}
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
					console.CommandLine.BrowseHistory(receivedReadlineCmd.Command)

				case CURSOR_LEFT, CURSOR_RIGHT:
					console.CommandLine.MoveCursor(receivedReadlineCmd.Command)
				}
				if console.renderer != nil {
					console.renderer.EventCh() <- UpdateCommandLineEvent{console.CommandLine}
				}

				receivedReadlineCmd.Done <- true

			case <-ticker.C: // Blink cursor
				if console.renderer != nil {
					console.renderer.EventCh() <- UpdateCursorEvent{console.CommandLine, toggleCursor}
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
