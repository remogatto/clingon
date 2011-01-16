package clingon

import (
	"os"
	"time"
	"strings"
	"container/vector"
)

const (
	CARRIAGE_RETURN = 0X000d
	BACKSPACE = 0x0008

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
	console     *Console
	commandLine *commandLine
}

// Pause/Unpause related event
type PauseEvent struct {
	paused  bool
	console *Console
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

func stringVectorToString(v *vector.StringVector) string {
	var currLine string
	for _, str := range *v {
		currLine += str
	}
	return currLine
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

func (commandLine *commandLine) setPrompt(prompt string) {
	commandLine.prompt = prompt
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
	if commandLine.contentToString() != "" && commandLine.notInHistory(commandLine.contentToString()) {
		commandLine.history.Push(commandLine.content)
	}
	line := commandLine.contentToString()
	commandLine.clear()
	return line
}

// Return the current prompt line as a single string (including prompt)
func (commandLine *commandLine) toString() string {
	return commandLine.prompt + commandLine.contentToString()
}

// Move the cursor left/right on the command line
func (commandLine *commandLine) moveCursor(dir int) {
	if dir == CURSOR_LEFT {
		commandLine.decCursorPosition(1)
	} else {
		commandLine.incCursorPosition(1)
	}
}

func (commandLine *commandLine) contentToString() string {
	return stringVectorToString(commandLine.content)
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
		historyEntry := stringVectorToString(strVector)
		if line == historyEntry {
			return false
		}
	}
	return true
}

type Console struct {
	paused      bool                 // Is the console paused?
	commandLine *commandLine         // An instance of the commandline
	lines       *vector.StringVector // Lines of text above the commandline 
	renderer    Renderer
	evaluator   Evaluator
}

// Initialize a new console object. Initially the console is
// paused. You have to unpause it in order to start rendering. The
// evaluator argument can be nil.
func NewConsole(renderer Renderer, evaluator Evaluator) *Console {
	console := &Console{
		lines:       new(vector.StringVector),
		commandLine: newCommandLine("console> "),
		renderer:    renderer,
		evaluator:   evaluator,
	}
	if renderer == nil {
		panic("Renderer can't be nil")
	}
	go console.loop()
	return console
}

// Dump the console content as a string.
func (console *Console) String() string {
	var result string
	for _, str := range *console.lines {
		result += str + "\n"
	}
	result += console.commandLine.toString()
	return result
}

// Set the prompt string
func (console *Console) SetPrompt(prompt string) {
	console.commandLine.setPrompt(prompt)
}

// Print a slice of strings on the console.
func (console *Console) PrintLines(strings []string) {
	if len(strings) > 0 {
		console.pushLines(strings)
	}
	if !console.paused {
		console.renderer.EventCh() <- UpdateConsoleEvent{console}
	}
}

// Print a string on the console. The string is splitted in lines by
// newline characters.
func (console *Console) Print(str string) {
	if str != "" {
		console.pushLines(strings.Split(str, "\n", -1))
	}
	if !console.paused {
		console.renderer.EventCh() <- UpdateConsoleEvent{console}
	}
}

// Get the current command line as string.
func (console *Console) Commandline() string {
	return console.commandLine.toString()
}

// Put a command and execute it
func (console *Console) PutCommand(command string) {
	console.PutString(command)
	console.PutUnicode(CARRIAGE_RETURN)
}

// Put an unicode value on the commandline at the current cursor
// position.
func (console *Console) PutUnicode(value uint16) {
	var event interface{}
	switch value {
	case BACKSPACE: // BACKSPACE
		console.commandLine.backSpace()
		event = UpdateCommandLineEvent{console, console.commandLine}
	case CARRIAGE_RETURN: // RETURN
		command := console.carriageReturn()
		if console.evaluator != nil {
			if err := console.evaluator.Run(console, command); err != nil {
				console.Print(err.String())
			}
		}
		event = UpdateConsoleEvent{console}
	default:
		console.commandLine.insertChar(string(value))
		event = UpdateCommandLineEvent{console, console.commandLine}
	}
	if !console.paused {
		console.renderer.EventCh() <- event
	}
}

// Put a readline-like command on the commandline (e.g. history
// browsing, cursor movements, etc.). Readline's commands emulation is
// incomplete.
func (console *Console) PutReadline(command int) {
	switch command {
	case HISTORY_NEXT, HISTORY_PREV:
		console.commandLine.browseHistory(command)
	case CURSOR_LEFT, CURSOR_RIGHT:
		console.commandLine.moveCursor(command)
	}
	if !console.paused {
		console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
	}
}

// Put the given string on the command line at the current cursor
// position.
func (console *Console) PutString(str string) {
	for _, c := range str {
		console.commandLine.insertChar(string(c))
	}
	if !console.paused {
		console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
	}
}

// Clear the commandline.
func (console *Console) ClearCommandline() {
	console.commandLine.clear()
	if !console.paused {
		console.renderer.EventCh() <- UpdateCommandLineEvent{console, console.commandLine}
	}
}

// Pause/Unpause the console. When the console is paused no events
// will be sent to the renderer. However, client-code could change the
// internal state of the console through the API.
func (console *Console) Pause(value bool) {
	console.paused = value
	console.renderer.EventCh() <- PauseEvent{value, console}
}

// Check if the console is paused.
func (console *Console) Paused() bool {
	return console.paused
}

// Push lines of text on the console. The argument is a slice of
// strings.
func (console *Console) pushLines(lines []string) {
	for _, line := range lines {
		console.lines.Push(line)
	}
}

// Push the current commandline in the console history. Return it has
// a string.
func (console *Console) carriageReturn() string {
	commandLine := console.commandLine.push()
	console.lines.Push(console.commandLine.prompt + commandLine)
	return commandLine
}

func (console *Console) loop() {
	var toggleCursor bool
	ticker := time.NewTicker(CURSOR_BLINK_TIME)
	for {
		<-ticker.C // Blink cursor
		if !console.paused {
			console.renderer.EventCh() <- UpdateCursorEvent{console.commandLine, toggleCursor}
			toggleCursor = !toggleCursor
		}
	}
}
