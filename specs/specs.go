package specs

import (
	"⚛sdl"
	"⚛sdl/ttf"
	pt "spectrum/prettytest"
	c "console"
)

var (
	console *c.Console
	sdlrenderer *c.SDLRenderer
	appSurface, gopher *sdl.Surface
)

func render() {
	appSurface.FillRect(&sdl.Rect{0, 0, 640, 480}, 0)
	appSurface.Blit(&sdl.Rect{0, 0, 0, 0}, gopher, nil)
	appSurface.Blit(&sdl.Rect{0, 0, 0, 0}, sdlrenderer.GetSurface(), nil)
	appSurface.Flip()
}

func before(t *pt.T) {
	if sdl.Init(sdl.INIT_VIDEO) != 0 {
		panic(sdl.GetError())
	}

	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	font := ttf.OpenFont("../testdata/fontin_sans.otf", 12)

	if font == nil {
		panic(sdl.GetError())
	}

	appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	gopher = sdl.Load("../testdata/gopher.png")

	sdlrenderer = c.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 640, 480, 32, 0, 0, 0, 0))
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	sdlrenderer.SetFont(font)
	console = c.NewConsole(sdlrenderer)
	console.SetPrompt("console>")
	sdlrenderer.Init(console)

	render()
}

func after(t *pt.T) {
	sdl.Quit()
}

