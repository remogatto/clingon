package console

import (
	"⚛sdl"
	"⚛sdl/ttf"
)

type SDLRenderer struct {
	surface *sdl.Surface
	width, height uint16
	promptLineRect *sdl.Rect
	font *ttf.Font
	updateCh chan *Console
}

func NewSDLRenderer(surface *sdl.Surface) *SDLRenderer {
	rect := new(sdl.Rect)
	surface.GetClipRect(rect)
	
	renderer := &SDLRenderer{

	width: rect.W,
	height: rect.H,
	surface: surface,
	updateCh: make(chan *Console),

	}

	renderer.promptLineRect = &sdl.Rect{0, int16(rect.H - 20), rect.W, rect.H}
	
	go renderer.loop()

	return renderer
}

func (renderer *SDLRenderer) SetFont(font *ttf.Font) {
	renderer.font = font
}

func (renderer *SDLRenderer) UpdateCh() chan *Console {
	return renderer.updateCh
}

func (renderer *SDLRenderer) update(console *Console) {
	white := sdl.Color{255, 255, 255, 0}
	textSurface := ttf.RenderText_Blended(renderer.font, console.GetCurrentLine(), white)

	renderer.surface.FillRect(renderer.promptLineRect, 0)
	renderer.surface.Blit(renderer.promptLineRect, textSurface, nil)
}

// Return the internal SDL surface
func (renderer *SDLRenderer) GetSurface() *sdl.Surface {
	return renderer.surface
}

func (renderer *SDLRenderer) loop() {
	for {
		select {
		case console := <-renderer.updateCh:
			renderer.update(console)
			console.renderingDone <- true
		}
	}
}
