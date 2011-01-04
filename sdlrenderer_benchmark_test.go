package clingon

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"fmt"
	"testing"
)

var (
	appSurface  *sdl.Surface
	sdlrenderer *SDLRenderer
)

func initSDL() {
	if sdl.Init(sdl.INIT_VIDEO) != 0 {
		panic(sdl.GetError())
	}

	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	font := ttf.OpenFont("testdata/VeraMono.ttf", 20)

	if font == nil {
		panic(sdl.GetError())
	}

	appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	sdlrenderer = NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 640, 480, 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)
	console = NewConsole(nil, nil)

	go func() {
		for {
			select {
			case rects := <-sdlrenderer.UpdatedRectsCh():
				render(rects)
			}

		}
	}()

	render(nil)
}

func render(updatedRects []sdl.Rect) {
	if updatedRects == nil {
		appSurface.Flip()
	} else {
		for _, r := range updatedRects {
			appSurface.Blit(&r, sdlrenderer.GetSurface(), &r)
			appSurface.UpdateRect(int32(r.X), int32(r.Y), uint32(r.W), uint32(r.H))
		}
	}
}

func BenchmarkRenderConsoleBlended(b *testing.B) {
	b.StopTimer()
	initSDL()
	var strings []string = make([]string, 0)
	for i := 0; i < 1000; i++ {
		strings = append(strings, fmt.Sprintf("Line %d %s", i, "The quick brown fox jumps over the lazy dog"))
	}
	console.pushLines(strings)
	sdlrenderer.Blended = true
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sdlrenderer.renderConsole(console)
		sdlrenderer.render(nil)
	}
}

func BenchmarkRenderConsoleSolid(b *testing.B) {
	b.StopTimer()
	initSDL()
	var strings []string = make([]string, 0)
	for i := 0; i < 1000; i++ {
		strings = append(strings, fmt.Sprintf("Line %d %s", i, "The quick brown fox jumps over the lazy dog"))
	}
	console.pushLines(strings)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sdlrenderer.renderConsole(console)
		sdlrenderer.render(nil)
	}
}
