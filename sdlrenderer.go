package clingon

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"time"
	"unsafe"
)

const DEFAULT_SDL_RENDERER_FPS = 20.0

type metrics struct {
	// Surface width, height
	width, height uint16

	cursorWidth, cursorHeight uint16
	fontWidth, fontHeight     int
	cursorY                   int16
	lastVisibleLine           int
	commandLineRect           *sdl.Rect
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
		uint16(m.fontHeight),
	}
}

type SDLRenderer struct {
	// Frame per-second
	FPS float
	// Activate/deactivate blended text rendering. By default use
	// solid text rendering
	Blended bool
	// Set the foreground color of the font (white by default)
	Color sdl.Color
	// Set the font family
	Font *ttf.Font

	layout                                               metrics
	internalSurface, visibleSurface                      *sdl.Surface
	cursorOn                                             bool
	renderCommandLineCh, renderConsoleCh, renderCursorCh chan *Console
	enableCursorCh                                       chan bool
	updatedRects                                         []sdl.Rect
	updatedRectsCh                                       chan []sdl.Rect
}

func NewSDLRenderer(surface *sdl.Surface, font *ttf.Font) *SDLRenderer {
	rect := new(sdl.Rect)
	surface.GetClipRect(rect)

	renderer := &SDLRenderer{
		FPS:                 DEFAULT_SDL_RENDERER_FPS,
		Color:               sdl.Color{255, 255, 255, 0},
		internalSurface:     sdl.CreateRGBSurface(sdl.SWSURFACE, int(surface.W), int(surface.H), 32, 0, 0, 0, 0),
		visibleSurface:      surface,
		Font:                font,
		renderCommandLineCh: make(chan *Console),
		renderConsoleCh:     make(chan *Console),
		renderCursorCh:      make(chan *Console),
		enableCursorCh:      make(chan bool),
		updatedRects:        make([]sdl.Rect, 0),
		updatedRectsCh:      make(chan []sdl.Rect),
	}

	fontWidth, fontHeight, _ := font.SizeText("A")

	renderer.layout = metrics{

		width:      rect.W,
		height:     rect.H,
		fontWidth:  fontWidth,
		fontHeight: fontHeight,
	}

	renderer.layout.calcMetrics()

	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, renderer.layout.width, renderer.layout.height})

	go renderer.loop()

	return renderer
}

func (renderer *SDLRenderer) RenderCommandLineCh() chan<- *Console {
	return renderer.renderCommandLineCh
}

func (renderer *SDLRenderer) RenderConsoleCh() chan<- *Console {
	return renderer.renderConsoleCh
}

func (renderer *SDLRenderer) EnableCursorCh() chan<- bool {
	return renderer.enableCursorCh
}

func (renderer *SDLRenderer) RenderCursorCh() chan<- *Console {
	return renderer.renderCursorCh
}

func (renderer *SDLRenderer) UpdatedRectsCh() <-chan []sdl.Rect {
	return renderer.updatedRectsCh
}

func (renderer *SDLRenderer) renderCommandLine(console *Console) {
	renderer.clearPrompt()
	renderer.renderLine(0, console.CommandLine.String())
	renderer.renderCursor(console)
}

func (renderer *SDLRenderer) renderConsole(console *Console) {
	renderer.clear()
	for i := console.lines.Len(); i > 0; i-- {
		if i < renderer.layout.lastVisibleLine {
			renderer.renderLine(i, console.lines.At(console.lines.Len()-i))
		} else {
			continue
		}
	}
	renderer.renderLine(0, console.CommandLine.String())
}

func (renderer *SDLRenderer) enableCursor(enable bool) {
	renderer.cursorOn = enable
}

// Return the internal SDL surface
func (renderer *SDLRenderer) GetSurface() *sdl.Surface {
	return renderer.visibleSurface
}

func (renderer *SDLRenderer) clear() {
	renderer.internalSurface.FillRect(&sdl.Rect{0, 0, renderer.layout.width, renderer.layout.height}, 0)
}

func (renderer *SDLRenderer) clearPrompt() {
	renderer.internalSurface.FillRect(renderer.layout.commandLineRect, 0)
}

func (renderer *SDLRenderer) renderLine(pos int, line string) {
	var textSurface *sdl.Surface

	if renderer.Blended {
		textSurface = ttf.RenderUTF8_Blended(renderer.Font, line, renderer.Color)
	} else {
		textSurface = ttf.RenderUTF8_Solid(renderer.Font, line, renderer.Color)
	}

	x := renderer.layout.commandLineRect.X
	y := int16(renderer.layout.commandLineRect.Y) - int16(renderer.layout.commandLineRect.H*uint16(pos))
	w := renderer.layout.commandLineRect.W
	h := renderer.layout.commandLineRect.H

	if textSurface != nil {
		renderer.internalSurface.Blit(&sdl.Rect{x, y, w, h}, textSurface, nil)
		textSurface.Free()
	}

	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{x, y, w, h})
}

// Return the address of pixel at (x,y)
func (renderer *SDLRenderer) getSurfaceAddr(x, y uint) uintptr {
	pixels := uintptr(unsafe.Pointer(renderer.internalSurface.Pixels))
	offset := uintptr(y*uint(renderer.internalSurface.Pitch) + x*uint(renderer.internalSurface.Format.BytesPerPixel))
	return uintptr(unsafe.Pointer(pixels + offset))
}

func (renderer *SDLRenderer) renderXORRect(x0, y0 int16, w, h uint16, color uint32) {
	x1 := x0 + int16(w)
	y1 := y0 + int16(h)
	for y := y0; y < y1; y++ {
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
	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{x, renderer.layout.cursorY, renderer.layout.cursorWidth, renderer.layout.cursorHeight})
}

func (renderer *SDLRenderer) cursorX(commandLine *CommandLine) int16 {
	var (
		cursorX  int = int(renderer.layout.commandLineRect.X)
		finalPos = commandLine.cursorPosition + len(commandLine.Prompt)
	)

	for pos, c := range commandLine.String() {
		_, _, _, _, advance, _ := renderer.Font.GlyphMetrics(uint16(c))
		if pos < finalPos {
			cursorX += advance
		} else {
			break
		}
	}
	return int16(cursorX)
}

func (renderer *SDLRenderer) renderCursor(console *Console) {
	renderer.renderCursorRect(renderer.cursorX(console.CommandLine))
}

func (renderer *SDLRenderer) render() {
	for _, r := range renderer.updatedRects {
		renderer.visibleSurface.Blit(&r, renderer.internalSurface, &r)
	}
	renderer.updatedRectsCh <- renderer.updatedRects
	renderer.updatedRects = make([]sdl.Rect, 0)
}

func (renderer *SDLRenderer) loop() {
	var ticker *time.Ticker
	if renderer.FPS > 0 {
		ticker = time.NewTicker(1e9 / int64(renderer.FPS))
	} else {
		ticker = time.NewTicker(1)
	}
	for {
		select {
		case console := <-renderer.renderCommandLineCh:
			renderer.renderCommandLine(console)
		case console := <-renderer.renderConsoleCh:
			renderer.renderConsole(console)
		case enableCursor := <-renderer.enableCursorCh:
			renderer.enableCursor(enableCursor)
		case console := <-renderer.renderCursorCh:
			renderer.renderCursor(console)
		case <-ticker.C:
			renderer.render()
		}
	}

}
