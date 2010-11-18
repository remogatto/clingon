package console

import (
	pt "spectrum/prettytest"
	"⚛sdl"
	"⚛sdl/ttf"
)

var (
	console *Console
	sdlrenderer *SDLRenderer
	appSurface *sdl.Surface
)

func beforeSDLConsole(t *pt.T) {
	if sdl.Init(sdl.INIT_VIDEO) != 0 {
		panic(sdl.GetError())
	}

	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	font := ttf.OpenFont("testdata/fontin_sans.otf", 12)

	if font == nil {
		panic(sdl.GetError())
	}

	appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	sdlrenderer = NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 640, 480, 32, 0, 0, 0, 0))
	sdlrenderer.SetFont(font)
	console = NewConsole(sdlrenderer)
	console.SetPrompt("console>")
}

func afterSDLConsole(t *pt.T) {
	sdl.Quit()
}
