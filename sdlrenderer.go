package clingon

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"time"
	"unsafe"
)

const DEFAULT_CONSOLE_RENDERER_FPS = 10.0

const (
	SCROLL_UP = iota
	SCROLL_DOWN
	SCROLL_UP_ANIMATION
	SCROLL_DOWN_ANIMATION
)

type metrics struct {
	// Surface width, height
	width, height uint16

	cursorWidth, cursorHeight uint16
	fontWidth, fontHeight     int
}

func (m *metrics) calcMetrics() {
	m.cursorWidth = uint16(m.fontWidth)
	m.cursorHeight = uint16(m.fontHeight)
}

type SDLRenderer struct {
	// Activate/deactivate blended text rendering. By default use
	// solid text rendering
	Blended bool
	// Set the foreground color of the font (white by default)
	Color sdl.Color
	// Set the font family
	Font *ttf.Font

	Animations map[int]*Animation

	// Frame per-second, greater than 0
	fps float

	layout                   metrics
	internalSurface          *sdl.Surface
	visibleSurface           *sdl.Surface
	cursorOn                 bool
	eventCh                  chan interface{}
	scrollCh                 chan int
	fpsCh                    chan float
	updatedRectsCh           chan []sdl.Rect
	updatedRects             []sdl.Rect
	viewportY                int16
	commandLineRect          *sdl.Rect
	cursorY                  int16
	lastVisibleLine          int
	internalSurfaceMaxHeight uint16
}

func NewSDLRenderer(surface *sdl.Surface, font *ttf.Font) *SDLRenderer {
	rect := new(sdl.Rect)
	surface.GetClipRect(rect)

	renderer := &SDLRenderer{
		fps:            DEFAULT_CONSOLE_RENDERER_FPS,
		Color:          sdl.Color{255, 255, 255, 0},
		Animations:     make(map[int]*Animation),
		visibleSurface: surface,
		Font:           font,
		eventCh:        make(chan interface{}),
		scrollCh:       make(chan int),
		fpsCh:          make(chan float),
		updatedRectsCh: make(chan []sdl.Rect),
		updatedRects:   make([]sdl.Rect, 0),
	}

	fontWidth, fontHeight, _ := font.SizeText("A")

	renderer.layout = metrics{

		width:      rect.W,
		height:     rect.H,
		fontWidth:  fontWidth,
		fontHeight: fontHeight,
	}

	renderer.layout.calcMetrics()

	renderer.lastVisibleLine = int(float(renderer.layout.height)/float(renderer.layout.fontHeight)) * 2
	renderer.internalSurfaceMaxHeight = renderer.layout.height * 2

	renderer.Animations[SCROLL_UP_ANIMATION] = NewSlideUpAnimation(1e9, 1.0)
	renderer.Animations[SCROLL_DOWN_ANIMATION] = NewSlideUpAnimation(1e9, 1.0)

	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, renderer.layout.width, renderer.layout.height})

	go renderer.loop()

	return renderer
}

func (renderer *SDLRenderer) EventCh() chan<- interface{} {
	return renderer.eventCh
}

// Tell the renderer to change its FPS (frames per second)
func (renderer *SDLRenderer) ScrollCh() chan<- int {
	return renderer.scrollCh
}

// Tell the renderer to change its FPS (frames per second)
func (renderer *SDLRenderer) FPSCh() chan<- float {
	return renderer.fpsCh
}

func (renderer *SDLRenderer) UpdatedRectsCh() <-chan []sdl.Rect {
	return renderer.updatedRectsCh
}

// Return the visible SDL surface
func (renderer *SDLRenderer) GetSurface() *sdl.Surface {
	return renderer.visibleSurface
}

func (renderer *SDLRenderer) calcCommandLineRect() {
	renderer.commandLineRect = &sdl.Rect{
		0,
		int16(renderer.internalSurface.H) - int16(renderer.layout.fontHeight),
		uint16(renderer.internalSurface.W),
		uint16(renderer.layout.fontHeight),
	}
}

func (renderer *SDLRenderer) resizeInternalSurface(console *Console) {
	if renderer.internalSurface != nil {
		renderer.internalSurface.Free()
	}

	h := uint16((console.lines.Len() + 1) * renderer.layout.fontHeight)

	if h > renderer.internalSurfaceMaxHeight {
		h = renderer.internalSurfaceMaxHeight
	}

	renderer.internalSurface = sdl.CreateRGBSurface(sdl.SWSURFACE, int(renderer.layout.width), int(h), 32, 0, 0, 0, 0)
	renderer.calcCommandLineRect()
	renderer.cursorY = renderer.commandLineRect.Y
	renderer.viewportY = int16(renderer.internalSurface.H - renderer.visibleSurface.H)
}

func (renderer *SDLRenderer) renderCommandLine(commandLine *CommandLine) {
	renderer.clearPrompt()
	renderer.renderLine(0, commandLine.String())
	renderer.renderCursor(commandLine)
}

func (renderer *SDLRenderer) renderConsole(console *Console) {
	renderer.resizeInternalSurface(console)
	for i := console.lines.Len(); i > 0; i-- {
		if i < renderer.lastVisibleLine {
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

func (renderer *SDLRenderer) clear() {
	renderer.internalSurface.FillRect(nil, 0)
}

func (renderer *SDLRenderer) clearPrompt() {
	renderer.internalSurface.FillRect(renderer.commandLineRect, 0)
}

func (renderer *SDLRenderer) renderLine(pos int, line string) {
	var textSurface *sdl.Surface

	if renderer.Blended {
		textSurface = ttf.RenderUTF8_Blended(renderer.Font, line, renderer.Color)
	} else {
		textSurface = ttf.RenderUTF8_Solid(renderer.Font, line, renderer.Color)
	}

	x := renderer.commandLineRect.X
	y := int16(renderer.commandLineRect.Y) - int16(renderer.commandLineRect.H*uint16(pos))
	w := renderer.commandLineRect.W
	h := renderer.commandLineRect.H

	if textSurface != nil {
		renderer.internalSurface.Blit(&sdl.Rect{x, y, w, h}, textSurface, nil)
		textSurface.Free()
	}

	renderer.addUpdatedRect(&sdl.Rect{x, y, w, h})
}

func (renderer *SDLRenderer) addUpdatedRect(rect *sdl.Rect) {
	visibleX, visibleY := renderer.transformToVisibleXY(rect.X, rect.Y)
	visibleRect := sdl.Rect{visibleX, visibleY, rect.W, rect.H}

	if visibleRect.Y > int16(renderer.layout.height) {
		visibleRect.Y = int16(renderer.layout.height)
	}
	if visibleRect.Y < 0 {
		visibleRect.Y = 0
	}
	if visibleRect.H > uint16(renderer.layout.height) {
		visibleRect.H = uint16(renderer.layout.height)
	}

	renderer.updatedRects = append(renderer.updatedRects, visibleRect)
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
	renderer.renderXORRect(x, renderer.cursorY, renderer.layout.cursorWidth, renderer.layout.cursorHeight, cursorColor)
	renderer.addUpdatedRect(&sdl.Rect{x, renderer.cursorY, renderer.layout.cursorWidth, renderer.layout.cursorHeight})
}

func (renderer *SDLRenderer) cursorX(commandLine *CommandLine) int16 {
	var (
		cursorX  int = int(renderer.commandLineRect.X)
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

func (renderer *SDLRenderer) renderCursor(commandLine *CommandLine) {
	renderer.renderCursorRect(renderer.cursorX(commandLine))
}

func (renderer *SDLRenderer) hasScrolled() bool {
	if renderer.viewportY != int16(renderer.internalSurface.H-renderer.visibleSurface.H) {
		return true
	}
	return false
}

func (renderer *SDLRenderer) scroll(direction float64) {
	if direction > 0 {
		renderer.viewportY += int16(direction)
	} else {
		renderer.viewportY += int16(direction)
	}
	if renderer.viewportY > int16(renderer.internalSurface.H-renderer.visibleSurface.H) {
		renderer.viewportY = int16(renderer.internalSurface.H-renderer.visibleSurface.H)
	}
	if renderer.viewportY < 0 {
			renderer.viewportY = 0
	}
}

func (renderer *SDLRenderer) transformToVisibleXY(internalX, internalY int16) (int16, int16) {
	visibleX := internalX
	visibleY := internalY - renderer.viewportY
	return visibleX, visibleY
}

func (renderer *SDLRenderer) transformToInternalXY(visibleX, visibleY int16) (int16, int16) {
	internalX := visibleX
	internalY := visibleY + renderer.viewportY
	return internalX, internalY
}

func (renderer *SDLRenderer) render(rects []sdl.Rect) {
	if rects != nil {
		for _, r := range rects {
			internalX, internalY := renderer.transformToInternalXY(r.X, r.Y)
			renderer.visibleSurface.Blit(&r, renderer.internalSurface, &sdl.Rect{internalX, internalY, r.W, r.H})
		}
	} else {
		h := uint16(int16(renderer.internalSurface.H) - renderer.viewportY)
		renderer.visibleSurface.Blit(nil, renderer.internalSurface, &sdl.Rect{0, renderer.viewportY, renderer.layout.width, h})
		renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, renderer.layout.width, renderer.layout.height})
	}

	renderer.updatedRectsCh <- renderer.updatedRects
	renderer.updatedRects = make([]sdl.Rect, 0)
}

func (renderer *SDLRenderer) newTicker(fps float) *time.Ticker {
	if fps <= 0 {
		fps = DEFAULT_CONSOLE_RENDERER_FPS
	}
	renderer.fps = fps
	return time.NewTicker(int64(1e9 / fps))
}

func (renderer *SDLRenderer) loop() {
	var ticker *time.Ticker = time.NewTicker(int64(1e9 / renderer.fps))

	for {
		select {
		case untyped_event := <-renderer.eventCh:
			switch event := untyped_event.(type) {
			case UpdateCommandLineEvent:
				if renderer.hasScrolled() {
					renderer.renderConsole(event.console)
				}
				renderer.enableCursor(true)
				renderer.renderCommandLine(event.commandLine)
			case UpdateConsoleEvent:
				renderer.resizeInternalSurface(event.console)
				renderer.renderConsole(event.console)
			case UpdateCursorEvent:
				renderer.enableCursor(event.enabled)
				renderer.renderCursor(event.commandLine)
			case NewlineEvent:
			}
		case dir := <-renderer.scrollCh:
			if dir == SCROLL_DOWN {
				renderer.Animations[SCROLL_DOWN_ANIMATION].Start()
			} else {
				renderer.Animations[SCROLL_UP_ANIMATION].Start()
			}
			ticker.Stop()
		case value := <-renderer.Animations[SCROLL_UP_ANIMATION].ValueCh():
			renderer.scroll(-value * 10)
			renderer.render(nil)
		case <-renderer.Animations[SCROLL_UP_ANIMATION].FinishedCh():
			ticker = renderer.newTicker(-1)
		case value := <-renderer.Animations[SCROLL_DOWN_ANIMATION].ValueCh():
			renderer.scroll(value * 10)
			renderer.render(nil)
		case <-renderer.Animations[SCROLL_DOWN_ANIMATION].FinishedCh():
			ticker = renderer.newTicker(-1)
		case fps := <-renderer.fpsCh:
			ticker.Stop()
			ticker = renderer.newTicker(fps)
		case <-ticker.C:
			renderer.render(renderer.updatedRects)

		}
	}

}
