package main

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"fmt"
	"strings"
	"clingon"
)

var (
	console *clingon.Console
	sdlrenderer *clingon.SDLRenderer
	appSurface, gopher *sdl.Surface
	running, toUpper bool
)

func render() {
	appSurface.Blit(&sdl.Rect{0, 0, 0, 0}, gopher, nil)
	appSurface.Blit(&sdl.Rect{40, 40, 0, 0}, sdlrenderer.GetSurface(), &sdl.Rect{0,0, 560, 400})
	appSurface.UpdateRect(40, 40, 560, 400)
}

func sdlinit() {
	if sdl.Init(sdl.INIT_VIDEO) != 0 {
		panic(sdl.GetError())
	}

	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	font := ttf.OpenFont("../testdata/VeraMono.ttf", 12)

	if font == nil {
		panic(sdl.GetError())
	}

	sdl.EnableUNICODE(1)
	appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	gopher = sdl.Load("../testdata/gopher.jpg")

	sdlrenderer = clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 560, 400, 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	console = clingon.NewConsole(sdlrenderer, &ShellEvaluator{})
	console.SetPrompt("shell:$ ")
}

func main() {
	sdlinit()
	render()
	appSurface.Flip()
	running = true
	go func() {
		for running {

			select {
			case event := <-sdl.Events:
				switch e := event.(type) {
				case sdl.QuitEvent:
					running = false
				case sdl.KeyboardEvent:
					keyName := sdl.GetKeyName(sdl.Key(e.Keysym.Sym))

					fmt.Printf("\n")
					fmt.Printf("%v: %v", e.Keysym.Sym, ": ", keyName)

					fmt.Printf("%04x ", e.Type)

					for i := 0; i < len(e.Pad0); i++ {
						fmt.Printf("%02x ", e.Pad0[i])
					}
					fmt.Printf("\n")

					fmt.Printf("Type: %02x Which: %02x State: %02x Pad: %02x\n", e.Type, e.Which, e.State, e.Pad0[0])
					fmt.Printf("Scancode: %02x Sym: %08x Mod: %04x Unicode: %04x\n", e.Keysym.Scancode, e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode)

					if (keyName == "escape") && (e.Type == sdl.KEYDOWN) {
						fmt.Printf("escape key -> request[exit the application]\n")
						running = false
					} else if (keyName == "left shift") && (e.Type == sdl.KEYDOWN) {
						toUpper = true
					} else if (keyName == "up") && (e.Type == sdl.KEYDOWN) {
						console.HistoryCh() <- clingon.HISTORY_PREV
					} else if (keyName == "down") && (e.Type == sdl.KEYDOWN) {
						console.HistoryCh() <- clingon.HISTORY_NEXT
					} else if (keyName == "left") && (e.Type == sdl.KEYDOWN) {
						console.CursorCh() <- clingon.CURSOR_LEFT
					} else if (keyName == "right") && (e.Type == sdl.KEYDOWN) {
						console.CursorCh() <- clingon.CURSOR_RIGHT
					} else {
						unicode := e.Keysym.Unicode
						if unicode > 0 {
							if toUpper {
								console.CharCh() <- uint16([]int(strings.ToUpper(string(unicode)))[0])
							} else {
								console.CharCh() <- unicode
							}
						}
					}

				}
			}

		}
	}()

	for running { render() }

	ttf.Quit()
	sdl.Quit()
}
