package clingon

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"unsafe"
)

const (
	MAX_INTERNAL_SIZE_FACTOR     = 3
)

const (
	SCROLL_UP = iota
	SCROLL_DOWN
	SCROLL_UP_ANIMATION
	SCROLL_DOWN_ANIMATION
)

type SDLRenderer struct {
	// Activate/deactivate blended text rendering. By default use
	// solid text rendering
	Blended bool
	// Set the foreground color of the font (white by default)
	Color sdl.Color
	// Set the font family
	Font *ttf.Font

	// Map of built-in animations
	Animations map[int]*Animation

	internalSurface           *sdl.Surface
	visibleSurface            *sdl.Surface
	cursorOn                  bool
	eventCh                   chan interface{}
	scrollCh                  chan int
	updatedRectsCh            chan []sdl.Rect
	pauseCh                   chan bool
	updatedRects              []sdl.Rect
	viewportY                 int16
	commandLineRect           *sdl.Rect
	cursorY                   int16
	lastVisibleLine           int
	internalSurfaceMaxHeight  uint16
	width, height             uint16 // visible surface width, height
	cursorWidth, cursorHeight uint16 // cursor width, height
	fontWidth, fontHeight     int    // font width, height
	paused                    bool
}

func NewSDLRenderer(surface *sdl.Surface, font *ttf.Font) *SDLRenderer {
	rect := new(sdl.Rect)
	surface.GetClipRect(rect)

	renderer := &SDLRenderer{
		Color:          sdl.Color{255, 255, 255, 0},
		Animations:     make(map[int]*Animation),
		visibleSurface: surface,
		Font:           font,
		eventCh:        make(chan interface{}),
		scrollCh:       make(chan int),
		pauseCh:        make(chan bool),
		updatedRectsCh: make(chan []sdl.Rect),
		updatedRects:   make([]sdl.Rect, 0),
		width:          rect.W,
		height:         rect.H,
	}

	fontWidth, fontHeight, _ := font.SizeText("A")

	renderer.fontWidth = fontWidth
	renderer.fontHeight = fontHeight

	renderer.cursorWidth = uint16(fontWidth)
	renderer.cursorHeight = uint16(fontHeight)

	renderer.lastVisibleLine = (int(renderer.height)/renderer.fontHeight) * MAX_INTERNAL_SIZE_FACTOR
	renderer.internalSurfaceMaxHeight = renderer.height * MAX_INTERNAL_SIZE_FACTOR

	renderer.internalSurface = sdl.CreateRGBSurface(sdl.SWSURFACE, int(renderer.width), int(renderer.fontHeight), 32, 0, 0, 0, 0)
	renderer.calcCommandLineRect()

	renderer.Animations[SCROLL_UP_ANIMATION] = NewSlideUpAnimation(1e9, 10.0)
	renderer.Animations[SCROLL_DOWN_ANIMATION] = NewSlideUpAnimation(1e9, 10.0)

	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, renderer.width, renderer.height})

	go renderer.loop()

	return renderer
}

// Receive the events triggered by the console.
func (renderer *SDLRenderer) EventCh() chan<- interface{} {
	return renderer.eventCh
}

// Tell the renderer to scroll up/down the console.
func (renderer *SDLRenderer) ScrollCh() chan<- int {
	return renderer.scrollCh
}

// Tell the renderer to pause its operations.
func (renderer *SDLRenderer) PauseCh() chan<- bool {
	return renderer.pauseCh
}

// From this channel, the client receives the rects representing the
// updated surface regions.
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
		int16(renderer.internalSurface.H) - int16(renderer.fontHeight),
		uint16(renderer.internalSurface.W),
		uint16(renderer.fontHeight),
	}
}

func (renderer *SDLRenderer) resizeInternalSurface(console *Console) {
	if renderer.internalSurface != nil {
		renderer.internalSurface.Free()
	}

	h := uint16((console.lines.Len() + 1) * renderer.fontHeight)

	if h > renderer.internalSurfaceMaxHeight {
		h = renderer.internalSurfaceMaxHeight
	}

	renderer.internalSurface = sdl.CreateRGBSurface(sdl.SWSURFACE, int(renderer.width), int(h), 32, 0, 0, 0, 0)
	renderer.calcCommandLineRect()
	renderer.cursorY = renderer.commandLineRect.Y
	renderer.viewportY = int16(renderer.internalSurface.H - renderer.visibleSurface.H)
	renderer.internalSurface.SetClipRect(&sdl.Rect{0, 0, renderer.width, h})
}

func (renderer *SDLRenderer) renderCommandLine(commandLine *commandLine) {
	renderer.clearPrompt()
	renderer.renderLine(0, commandLine.toString())
	renderer.addUpdatedRect(renderer.commandLineRect)
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
	renderer.addUpdatedRect(&sdl.Rect{0, 0, uint16(renderer.internalSurface.W), uint16(renderer.internalSurface.H)})
	renderer.renderLine(0, console.commandLine.toString())
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
}

func (renderer *SDLRenderer) addUpdatedRect(rect *sdl.Rect) {
	visibleX, visibleY := renderer.transformToVisibleXY(rect.X, rect.Y)
	visibleRect := sdl.Rect{visibleX, visibleY, rect.W, rect.H}

	if visibleRect.Y > int16(renderer.height) {
		visibleRect.Y = int16(renderer.height)
	}
	if visibleRect.Y < 0 {
		visibleRect.Y = 0
	}
	if visibleRect.H > uint16(renderer.height) {
		visibleRect.H = uint16(renderer.height)
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
	renderer.renderXORRect(x, renderer.cursorY, renderer.cursorWidth, renderer.cursorHeight, cursorColor)
	renderer.addUpdatedRect(&sdl.Rect{x, renderer.cursorY, renderer.cursorWidth, renderer.cursorHeight})
}

func (renderer *SDLRenderer) cursorX(commandLine *commandLine) int16 {
	var (
		cursorX  int = int(renderer.commandLineRect.X)
		finalPos = commandLine.cursorPosition + len(commandLine.prompt)
	)

	for pos, c := range commandLine.toString() {
		_, _, _, _, advance, _ := renderer.Font.GlyphMetrics(uint16(c))
		if pos < finalPos {
			cursorX += advance
		} else {
			break
		}
	}
	return int16(cursorX)
}

func (renderer *SDLRenderer) renderCursor(commandLine *commandLine) {
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
	if renderer.viewportY < 0 {
		renderer.viewportY = 0
	}
	if renderer.viewportY > int16(renderer.internalSurface.H-renderer.visibleSurface.H) {
		renderer.viewportY = int16(renderer.internalSurface.H - renderer.visibleSurface.H)
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
		renderer.visibleSurface.Blit(nil, renderer.internalSurface, &sdl.Rect{0, renderer.viewportY, renderer.width, h})
		renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, renderer.width, renderer.height})
	}

	renderer.updatedRectsCh <- renderer.updatedRects
	renderer.updatedRects = make([]sdl.Rect, 0)
}

func (renderer *SDLRenderer) loop() {
	// Control goroutine
	go func() {
		for {
			renderer.paused = <-renderer.pauseCh
		}
	}()
	for {
		if !renderer.paused {
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
					renderer.renderConsole(event.console)
				case UpdateCursorEvent:
					renderer.enableCursor(event.enabled)
					renderer.renderCursor(event.commandLine)
				case PauseEvent:
					renderer.paused = event.paused
				}
				renderer.render(renderer.updatedRects)
			case dir := <-renderer.scrollCh:
				if dir == SCROLL_DOWN {
					renderer.Animations[SCROLL_DOWN_ANIMATION].Start()
				} else {
					renderer.Animations[SCROLL_UP_ANIMATION].Start()
				}
			case value := <-renderer.Animations[SCROLL_UP_ANIMATION].ValueCh():
				renderer.scroll(-value)
				renderer.render(nil)
			case <-renderer.Animations[SCROLL_UP_ANIMATION].FinishedCh():
			case value := <-renderer.Animations[SCROLL_DOWN_ANIMATION].ValueCh():
				renderer.scroll(value)
				renderer.render(nil)
			case <-renderer.Animations[SCROLL_DOWN_ANIMATION].FinishedCh():
			}
		} else {
			select {
			case untyped_event := <-renderer.eventCh:
				switch event := untyped_event.(type) {
				case PauseEvent:
					renderer.paused = event.paused
					if !event.paused { // if unpaused re-render the console
						renderer.renderConsole(event.console)
					}
				}
			case <-renderer.scrollCh:
			case <-renderer.Animations[SCROLL_UP_ANIMATION].ValueCh():
			case <-renderer.Animations[SCROLL_UP_ANIMATION].FinishedCh():
			case <-renderer.Animations[SCROLL_DOWN_ANIMATION].ValueCh():
			case <-renderer.Animations[SCROLL_DOWN_ANIMATION].FinishedCh():
			}
		}
	}

}
