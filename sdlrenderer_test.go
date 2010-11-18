package console

import (
	"testing"
	pt "spectrum/prettytest"
	"⚛sdl"
	"⚛sdl/ttf"
)

func testInitSDLConsole(t *pt.T) {
	font := ttf.OpenFont("testdata/fontin_sans.otf", 10)

	sdlrenderer := NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 640, 480, 32, 0, 0, 0, 0))
	sdlrenderer.SetFont(font)

	t.NotNil(sdlrenderer)
}

func testSDLConsolePrint(t *pt.T) {
	console.CharCh() <- "h"

	<-console.renderingDone

	appSurface.Blit(&sdl.Rect{0, 0, 0, 0}, sdlrenderer.surface, nil)
	appSurface.Flip()
//	for {}
}

func TestSDLConsole(t *testing.T) {
	pt.Run(
		t,
		testInitSDLConsole,
		testSDLConsolePrint,

		beforeSDLConsole,
	)
}


