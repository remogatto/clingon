package clingon

import (
	"os"
	"time"
	"strings"
	"container/vector"
	"sync"
)

const (
	CARRIAGE_RETURN = 0x000d
	BACKSPACE       = 0x0008
	DELETE          = 0x007f

	CURSOR_BLINK_TIME = 5e8

	HISTORY_NEXT = iota
	HISTORY_PREV
	CURSOR_LEFT
	CURSOR_RIGHT
)

type Evaluator interface {
	Run(console *Console, command string) os.Error
}

type Cmd_UpdateConsole struct {
	console *Console
}

type Cmd_UpdateCursor struct {
	commandLine *commandLine
	enabled     bool
}

type Cmd_UpdateCommandLine struct {
	console     *Console
	commandLine *commandLine
}

type Renderer interface {
	// Receive Cmd_* objects from the console
	EventCh() chan<- interface{}
}

type commandLine struct {
	prompt string

	// The content of the prompt line as a string vector
	content *vector.StringVector

	// The history of the command line.
	// This is a vector of '*vector.StringVector'.
	history vector.Vector

	// Cursor position on current line
	cursorPosition int

	historyPosition int
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
		prompt:          prompt,
		content:         new(vector.StringVector),
		cursorPosition:  0,
		historyPosition: 0,
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
func (commandLine *commandLine) key_backspace() {
	if commandLine.cursorPosition >= 1 {
		commandLine.decCursorPosition(1)
		commandLine.content.Delete(commandLine.cursorPosition)
	}
}

// Delete the character at the current cursor position
func (commandLine *commandLine) key_delete() {
	if commandLine.cursorPosition < commandLine.content.Len() {
		commandLine.content.Delete(commandLine.cursorPosition)
	}
}

// Clear the prompt line
func (commandLine *commandLine) clear() {
	commandLine.content = new(vector.StringVector)
	commandLine.cursorPosition = 0
	commandLine.historyPosition = commandLine.history.Len()
}

// Browse history on the command line
func (commandLine *commandLine) browseHistory(direction int) {
	if direction == HISTORY_NEXT {
		commandLine.historyPosition++
	} else {
		commandLine.historyPosition--
	}
	if commandLine.historyPosition < 0 {
		commandLine.historyPosition = 0
		return
	}
	if commandLine.historyPosition >= commandLine.history.Len() {
		commandLine.clear()
		return
	}
	newContent := commandLine.history.At(commandLine.historyPosition).(*vector.StringVector).Copy()
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
	for _, v := range commandLine.history {
		strVector := v.(*vector.StringVector)
		historyEntry := stringVectorToString(strVector)
		if line == historyEntry {
			return false
		}
	}
	return true
}

type Console struct {
	commandLine     *commandLine         // An instance of the commandline
	lines           *vector.StringVector // Lines of text above the commandline
	renderer_orNil  Renderer
	evaluator_orNil Evaluator
	mu              sync.RWMutex
}

// Initialize a new console object passing to it an evaluator
func NewConsole(evaluator_orNil Evaluator) *Console {
	console := &Console{
		lines:           new(vector.StringVector),
		commandLine:     newCommandLine("console> "),
		renderer_orNil:  nil,
		evaluator_orNil: evaluator_orNil,
	}
	go console.loop()
	return console
}

// Return the current renderer.
func (console *Console) RendererOrNil() Renderer {
	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	return renderer
}

// Replace the current renderer with a new one
func (console *Console) SetRenderer(renderer_orNil Renderer) {
	console.mu.Lock()
	console.renderer_orNil = renderer_orNil
	console.mu.Unlock()

	if renderer_orNil != nil {
		renderer_orNil.EventCh() <- Cmd_UpdateConsole{console}
	}
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
// The strings should not contain '\n'.
func (console *Console) PrintLines(strings []string) {
	if len(strings) > 0 {
		console.pushLines(strings)
	}

	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	if renderer != nil {
		renderer.EventCh() <- Cmd_UpdateConsole{console}
	}
}

// Print a string on the console and go to the next line.
// This function can handle strings containing '\n'.
func (console *Console) Print(str string) {
	if strings.HasSuffix(str, "\n") {
		str = str[0 : len(str)-1]
	}
	console.PrintLines(strings.Split(str, "\n"))
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

// Put an unicode value on the command-line at the current cursor position.
func (console *Console) PutUnicode(value uint16) {
	var event interface{}

	switch value {
	case BACKSPACE:
		console.commandLine.key_backspace()
		event = Cmd_UpdateCommandLine{console, console.commandLine}

	case DELETE:
		console.commandLine.key_delete()
		event = Cmd_UpdateCommandLine{console, console.commandLine}

	case CARRIAGE_RETURN:
		command := console.carriageReturn()
		if console.evaluator_orNil != nil {
			if err := console.evaluator_orNil.Run(console, command); err != nil {
				console.Print(err.String())
			}
		}
		event = Cmd_UpdateConsole{console}

	default:
		console.commandLine.insertChar(string(value))
		event = Cmd_UpdateCommandLine{console, console.commandLine}
	}

	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	if renderer != nil {
		renderer.EventCh() <- event
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

	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	if renderer != nil {
		renderer.EventCh() <- Cmd_UpdateCommandLine{console, console.commandLine}
	}
}

// Put the given string on the command line at the current cursor
// position.
func (console *Console) PutString(str string) {
	for _, c := range str {
		console.commandLine.insertChar(string(c))
	}

	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	if renderer != nil {
		renderer.EventCh() <- Cmd_UpdateCommandLine{console, console.commandLine}
	}
}

// Clear the commandline.
func (console *Console) ClearCommandline() {
	console.commandLine.clear()

	console.mu.RLock()
	renderer := console.renderer_orNil
	console.mu.RUnlock()
	if renderer != nil {
		renderer.EventCh() <- Cmd_UpdateCommandLine{console, console.commandLine}
	}
}

// Push lines of text on the console. The argument is a slice of
// strings.
func (console *Console) pushLines(lines []string) {
	for _, line := range lines {
		console.lines.Push(line)
	}
}

// Push the current commandline in the console history.
// Return it as a string.
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

		console.mu.RLock()
		renderer := console.renderer_orNil
		console.mu.RUnlock()
		if renderer != nil {
			renderer.EventCh() <- Cmd_UpdateCursor{console.commandLine, toggleCursor}
		}

		toggleCursor = !toggleCursor
	}
}
