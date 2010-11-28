package clingon

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"unsafe"
)

type metrics struct {
	// Surface width, height
	width, height uint16

	cursorWidth, cursorHeight uint16
	fontWidth, fontHeight int
	cursorY int16
	lastVisibleLine int
	commandLineRect *sdl.Rect
}

func (m *metrics) calcMetrics() {
	m.cursorWidth = uint16(m.fontWidth) 
	m.cursorHeight = uint16(m.fontHeight)
	m.lastVisibleLine = int(float(m.height) / float(m.fontHeight))

	m.commandLineRect = m.calcCommandLineRect()

	m.cursorY = m.commandLineRect.Y
}

func (m *metrics) calcCommandLineRect() *sdl.Rect {
	return &sdl.Rect{
		0,
		int16(m.height) - int16(m.fontHeight),
		m.width,
		m.height - uint16(m.fontHeight),
	}
}

type SDLRenderer struct {
	layout metrics
	surface *sdl.Surface
	font *ttf.Font
	cursorOn bool
}

func NewSDLRenderer(surface *sdl.Surface, font *ttf.Font) *SDLRenderer {
	rect := new(sdl.Rect)
	surface.GetClipRect(rect)

	renderer := &SDLRenderer{

	surface: surface,
	font: font,
		
	}

	fontWidth, fontHeight, _ := font.SizeText("A")	

	renderer.layout = metrics{

	width: rect.W,
	height: rect.H,
	fontWidth: fontWidth,
	fontHeight: fontHeight,
	}
	
	renderer.layout.calcMetrics()

	return renderer
}

func (renderer *SDLRenderer) RenderCommandLine(commandLine *CommandLine) {
	renderer.clearPrompt()
	renderer.renderLine(0, commandLine.String())
	renderer.renderCursor(commandLine)
}

func (renderer *SDLRenderer) RenderVisibleLines(console *Console) {
	renderer.clear()
	
	for i := console.lines.Len(); i > 0; i-- {
		if i < renderer.layout.lastVisibleLine {
			renderer.renderLine(i, console.lines.At(console.lines.Len() - i))
		} else {
			continue
		}
	}
	renderer.renderLine(0, console.CommandLine.String())
}

func (renderer *SDLRenderer) EnableCursor(enable bool) {
	renderer.cursorOn = enable
}

// Return the internal SDL surface
func (renderer *SDLRenderer) GetSurface() *sdl.Surface {
	return renderer.surface
}

func (renderer *SDLRenderer) clear() {
	renderer.surface.FillRect(&sdl.Rect{0, 0, renderer.layout.width, renderer.layout.height}, 0)
}

func (renderer *SDLRenderer) clearPrompt() {
	renderer.surface.FillRect(renderer.layout.commandLineRect, 0)
}

func (renderer *SDLRenderer) renderLine(pos int, line string) {
	white := sdl.Color{255, 255, 255, 0}
	textSurface := ttf.RenderUTF8_Blended(renderer.font, line, white)

	x := renderer.layout.commandLineRect.X
	y := int16(renderer.layout.commandLineRect.Y) - int16(renderer.layout.commandLineRect.H * uint16(pos))
	w := renderer.layout.commandLineRect.W 
	h := renderer.layout.commandLineRect.H

	if textSurface != nil {
		renderer.surface.Blit(&sdl.Rect{x, y, w, h}, textSurface, nil)
	}
}

// Return the address of pixel at (x,y)
func (renderer *SDLRenderer) getSurfaceAddr(x, y uint) uintptr {
	pixels := uintptr(unsafe.Pointer(renderer.surface.Pixels))
	offset := uintptr(y*uint(renderer.surface.Pitch) + x*uint(renderer.surface.Format.BytesPerPixel))
	return uintptr(unsafe.Pointer(pixels + offset))
}

func (renderer *SDLRenderer) renderXORRect(x0, y0 int16, w, h uint16, color uint32) {
	x1 := x0 + int16(w)
	y1 := y0 + int16(h)
	for y:=y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			addr := renderer.getSurfaceAddr(uint(x), uint(y))
			currColor := *(*uint32)(unsafe.Pointer(addr))
			*(*uint32)(unsafe.Pointer(addr)) = ^(currColor ^ color)
		}
	}
}

func (renderer *SDLRenderer) renderCursorRect(x int16) {
	var cursorColor uint32
	if renderer.cursorOn {
		cursorColor = 0
	} else {
		cursorColor = 0xffffffff
	}
	renderer.renderXORRect(x, renderer.layout.cursorY, renderer.layout.cursorWidth, renderer.layout.cursorHeight, cursorColor)
}

func (renderer *SDLRenderer) cursorX(commandLine *CommandLine) int16 {
	var (
		cursorX int = int(renderer.layout.commandLineRect.X)
		finalPos = commandLine.cursorPosition + len(commandLine.Prompt)
	)
	
	for pos, c := range commandLine.String() {
		_, _, _, _, advance, _ := renderer.font.GlyphMetrics(uint16(c))
		if pos < finalPos {
			cursorX += advance
		} else {
			break
		}
	}
	return int16(cursorX)
}

func (renderer *SDLRenderer) renderCursor(commandLine *CommandLine) {	
	renderer.renderCursorRect(renderer.cursorX(commandLine))
}
