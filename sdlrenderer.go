package clingon

import (
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/ttf"
	"unsafe"
)

const (
	MAX_INTERNAL_SIZE_FACTOR = 3
)

const (
	SCROLL_UP = iota
	SCROLL_DOWN
	SCROLL_UP_ANIMATION
	SCROLL_DOWN_ANIMATION
)

// Tell the renderer to scroll the console text up/down
type Cmd_Scroll struct {
	// SCROLL_UP or SCROLL_DOWN
	Direction int
}

// Tell the renderer to stop scrolling the console text.
// It is safe to send this event even if no scrolling is running.
type Cmd_StopScroll struct{}

type Cmd_Terminate struct {
	Done chan bool
}

type SDLRenderer struct {
	// Activate/deactivate blended text rendering. By default use
	// solid text rendering
	Blended bool
	// Set the foreground color of the font (white by default)
	Color sdl.Color
	// Set the font family
	Font *ttf.Font

	internalSurface          *sdl.Surface
	visibleSurface           *sdl.Surface
	cursorOn                 bool
	eventCh                  chan interface{}
	updatedRectsCh           chan []sdl.Rect
	updatedRects             []sdl.Rect
	viewportY                int
	commandLineRect          *sdl.Rect
	cursorY                  int
	lastVisibleLine          uint
	internalSurfaceHeight    uint
	internalSurfaceMaxHeight uint
	width, height            uint // visible surface width, height
	fontWidth, fontHeight    uint
	terminated               bool
}

func NewSDLRenderer(surface *sdl.Surface, font *ttf.Font) *SDLRenderer {
	renderer := &SDLRenderer{
		Color:          sdl.Color{255, 255, 255, 0},
		visibleSurface: surface,
		Font:           font,
		eventCh:        make(chan interface{}),
		updatedRectsCh: make(chan []sdl.Rect),
		updatedRects:   make([]sdl.Rect, 0),
		width:          uint(surface.W),
		height:         uint(surface.H),
	}

	fontWidth, fontHeight, err := font.SizeText("A")
	if err != 0 {
		panic("failed to determine font dimensions: " + sdl.GetError())
	}

	renderer.fontWidth = uint(fontWidth)
	renderer.fontHeight = uint(fontHeight)

	renderer.lastVisibleLine = (renderer.height / renderer.fontHeight) * MAX_INTERNAL_SIZE_FACTOR

	renderer.internalSurfaceHeight = uint(renderer.fontHeight)
	renderer.internalSurfaceMaxHeight = renderer.height * MAX_INTERNAL_SIZE_FACTOR
	renderer.internalSurface = sdl.CreateRGBSurface(sdl.SWSURFACE,
		int(renderer.width), int(renderer.internalSurfaceHeight), 32, 0, 0, 0, 0)

	renderer.calcCommandLineRect()

	renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, uint16(renderer.width), uint16(renderer.height)})

	go renderer.loop()

	return renderer
}

// Receive the events triggered by the console.
func (renderer *SDLRenderer) EventCh() chan<- interface{} {
	return renderer.eventCh
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
	h := uint(len(console.lines)+1) * renderer.fontHeight
	if h > renderer.internalSurfaceMaxHeight {
		h = renderer.internalSurfaceMaxHeight
	}

	if renderer.internalSurfaceHeight != h {
		renderer.internalSurface.Free()

		renderer.internalSurface = sdl.CreateRGBSurface(sdl.SWSURFACE, int(renderer.width), int(h), 32, 0, 0, 0, 0)
		renderer.calcCommandLineRect()
		renderer.cursorY = int(renderer.commandLineRect.Y)
		renderer.viewportY = int(renderer.internalSurface.H - renderer.visibleSurface.H)
		renderer.internalSurfaceHeight = h
		renderer.internalSurface.SetClipRect(&sdl.Rect{0, 0, uint16(renderer.width), uint16(h)})
	}
}

func (renderer *SDLRenderer) renderCommandLine(commandLine *commandLine) {
	renderer.clearPrompt()
	renderer.renderLine(0, commandLine.toString())
	renderer.addUpdatedRect(renderer.commandLineRect)
	renderer.renderCursor(commandLine)
}

func (renderer *SDLRenderer) renderConsole(console *Console) {
	renderer.resizeInternalSurface(console)
	for i := len(console.lines); i > 0; i-- {
		if uint(i) < renderer.lastVisibleLine {
			renderer.renderLine(i, console.lines[len(console.lines)-i])
		} else {
			continue
		}
	}
	renderer.renderLine(0, console.commandLine.toString())
	renderer.addUpdatedRect(&sdl.Rect{0, 0, uint16(renderer.internalSurface.W), uint16(renderer.internalSurface.H)})
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
	x := renderer.commandLineRect.X
	y := int16(renderer.commandLineRect.Y) - int16(renderer.fontHeight*uint(pos))
	w := renderer.commandLineRect.W
	h := renderer.commandLineRect.H

	renderer.internalSurface.FillRect(&sdl.Rect{x, y, w, h}, 0)

	if len(line) == 0 {
		return
	}

	var textSurface *sdl.Surface
	if renderer.Blended {
		textSurface = ttf.RenderUTF8_Blended(renderer.Font, line, renderer.Color)
	} else {
		textSurface = ttf.RenderUTF8_Solid(renderer.Font, line, renderer.Color)
	}

	if textSurface != nil {
		renderer.internalSurface.Blit(&sdl.Rect{x, y, w, h}, textSurface, nil)
		textSurface.Free()
	}
}

func (renderer *SDLRenderer) addUpdatedRect(rect *sdl.Rect) {
	visibleX, visibleY := renderer.transformToVisibleXY(int(rect.X), int(rect.Y))
	visibleRect := sdl.Rect{int16(visibleX), int16(visibleY), rect.W, rect.H}

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

func (renderer *SDLRenderer) renderXORRect(x0, y0 int, w, h uint, color uint32) {
	x1 := x0 + int(w)
	y1 := y0 + int(h)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			addr := renderer.getSurfaceAddr(uint(x), uint(y))
			currColor := *(*uint32)(unsafe.Pointer(addr))
			*(*uint32)(unsafe.Pointer(addr)) = ^(currColor ^ color)
		}
	}
}

func (renderer *SDLRenderer) renderCursorRect(x int) {
	var cursorColor uint32
	if renderer.cursorOn {
		cursorColor = 0
	} else {
		cursorColor = 0xffffffff
	}
	renderer.renderXORRect(x, renderer.cursorY, renderer.fontWidth, renderer.fontHeight, cursorColor)
	renderer.addUpdatedRect(&sdl.Rect{int16(x), int16(renderer.cursorY), uint16(renderer.fontWidth), uint16(renderer.fontHeight)})
}

func (renderer *SDLRenderer) cursorX(commandLine *commandLine) int {
	var (
		cursorX  int = int(renderer.commandLineRect.X)
		finalPos     = commandLine.cursorPosition + len(commandLine.prompt)
	)

	for pos, c := range commandLine.toString() {
		_, _, _, _, advance, _ := renderer.Font.GlyphMetrics(uint16(c))
		if pos < finalPos {
			cursorX += advance
		} else {
			break
		}
	}
	return cursorX
}

func (renderer *SDLRenderer) renderCursor(commandLine *commandLine) {
	renderer.renderCursorRect(renderer.cursorX(commandLine))
}

func (renderer *SDLRenderer) hasScrolled() bool {
	return (renderer.viewportY != int(renderer.internalSurface.H-renderer.visibleSurface.H))
}

func (renderer *SDLRenderer) setViewporY(y int) {
	renderer.viewportY = y

	if renderer.viewportY < 0 {
		renderer.viewportY = 0
	}
	if renderer.viewportY > int(renderer.internalSurface.H-renderer.visibleSurface.H) {
		renderer.viewportY = int(renderer.internalSurface.H - renderer.visibleSurface.H)
	}
}

func (renderer *SDLRenderer) transformToVisibleXY(internalX, internalY int) (int, int) {
	visibleX := internalX
	visibleY := internalY - renderer.viewportY
	return visibleX, visibleY
}

func (renderer *SDLRenderer) transformToInternalXY(visibleX, visibleY int) (int, int) {
	internalX := visibleX
	internalY := visibleY + renderer.viewportY
	return internalX, internalY
}

func (renderer *SDLRenderer) render(rects []sdl.Rect) {
	if rects != nil {
		for _, r := range rects {
			internalX, internalY := renderer.transformToInternalXY(int(r.X), int(r.Y))
			renderer.visibleSurface.Blit(&r, renderer.internalSurface, &sdl.Rect{int16(internalX), int16(internalY), r.W, r.H})
		}
	} else {
		h := uint(int(renderer.internalSurface.H) - renderer.viewportY)
		renderer.visibleSurface.Blit(nil, renderer.internalSurface,
			&sdl.Rect{0, int16(renderer.viewportY), uint16(renderer.width), uint16(h)})
		renderer.updatedRects = append(renderer.updatedRects, sdl.Rect{0, 0, uint16(renderer.width), uint16(renderer.height)})
	}
	renderer.updatedRectsCh <- renderer.updatedRects
	renderer.updatedRects = make([]sdl.Rect, 0)
}

func (renderer *SDLRenderer) loop() {
	var anim_orNil *Animation = nil
	var animValueCh_orNil <-chan float64 = nil
	var baseViewportY int

	stopScrolling := func() {
		// Stop the running scroll animation (if any)
		if anim_orNil != nil {
			anim_orNil.Terminate()
			<-anim_orNil.FinishedCh()
			anim_orNil = nil
		}
	}

	for {
		select {
		case untyped_event := <-renderer.eventCh:
			switch event := untyped_event.(type) {
			case Cmd_UpdateCommandLine:
				stopScrolling()
				if renderer.hasScrolled() {
					renderer.renderConsole(event.console)
				}
				renderer.enableCursor(true)
				renderer.renderCommandLine(event.commandLine)

			case Cmd_UpdateConsole:
				renderer.renderConsole(event.console)

			case Cmd_UpdateCursor:
				renderer.enableCursor(event.enabled)
				renderer.renderCursor(event.commandLine)

			case Cmd_Scroll:
				stopScrolling()

				var newAnim *Animation
				if event.Direction == SCROLL_DOWN {
					newAnim = NewSliderAnimation(0.750, +0.8*float64(renderer.height))
				} else {
					newAnim = NewSliderAnimation(0.750, -0.8*float64(renderer.height))
				}

				anim_orNil = newAnim
				animValueCh_orNil = newAnim.ValueCh()
				baseViewportY = renderer.viewportY

				newAnim.Start()

			case Cmd_StopScroll:
				stopScrolling()

			case Cmd_Terminate:
				stopScrolling()
				renderer.updatedRectsCh <- nil
				renderer.internalSurface.Free()
				renderer.internalSurface = nil
				event.Done <- true
				return
			}

			renderer.render(renderer.updatedRects)

		case value := <-animValueCh_orNil:
			renderer.setViewporY(baseViewportY + int(value))
			renderer.render(nil)
		}
	}
}
