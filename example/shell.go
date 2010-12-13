package main

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"fmt"
	"flag"
	"math"
	"os"
	"clingon"
)

var (
	config      configuration
	console     *clingon.Console
	sdlrenderer *clingon.SDLRenderer
	running     bool
	r           *renderer
)

type configuration struct {
	consoleX, consoleY  int16
	consoleW, consoleH  uint16
	fullscreen, verbose, autoFps bool
	fps                 float
	bgImage             string
}

type renderer struct {
	config                                 *configuration
	appSurface, bgImageSurface, cliSurface *sdl.Surface
	animateCLI                             bool
	t                                      float64
}

const (
	animation_step = math.Pi / 50
)

func (r *renderer) render(updatedRects []sdl.Rect) {
	if updatedRects == nil { // Initially we must blit the entire surface
		if r.bgImageSurface != nil {
			r.appSurface.Blit(nil, r.bgImageSurface, nil)
		}
		r.appSurface.Blit(&sdl.Rect{config.consoleX, config.consoleY, 0, 0}, sdlrenderer.GetSurface(), nil)
		r.appSurface.Flip()
	} else { // Then we can keep updated only the changed regions
		if !r.animateCLI {
			for _, rect := range updatedRects {
				if r.bgImageSurface != nil {
					r.appSurface.Blit(
						&sdl.Rect{rect.X + r.config.consoleX, rect.Y + r.config.consoleY, 0, 0},
						r.bgImageSurface,
						&sdl.Rect{rect.X + r.config.consoleX, rect.Y + r.config.consoleY, rect.W, rect.H})
				}
				r.appSurface.Blit(
					&sdl.Rect{rect.X + r.config.consoleX, rect.Y + r.config.consoleY, 0, 0},
					sdlrenderer.GetSurface(), &rect)
				r.appSurface.UpdateRect(int32(rect.X+r.config.consoleX), int32(rect.Y+r.config.consoleY), uint32(rect.W), uint32(rect.H))
			}
		} else {
			if !console.Paused {
				if r.config.consoleY > 40 {
					r.config.consoleY = 40 + int16((480-40+1)*(1-math.Cos(r.t)))
					r.t -= animation_step
				}
				if r.config.consoleY <= 40 {
					r.t = 0
					r.config.consoleY = 40
					r.animateCLI = false
				}
			} else {
				if r.config.consoleY < 480 {
					r.config.consoleY = 40 + int16((480-40+1)*(1-math.Cos(r.t)))
					r.t += animation_step
				}
				if r.config.consoleY >= 480 {
					r.t = (math.Pi / 2)
					r.config.consoleY = 480
					r.animateCLI = false
				}
			}
			if r.bgImageSurface != nil {
				r.appSurface.Blit(nil, r.bgImageSurface, nil)
			}
			r.appSurface.Blit(&sdl.Rect{r.config.consoleX, r.config.consoleY, 0, 0}, sdlrenderer.GetSurface(), nil)
			r.appSurface.Flip()
		}
	}
}

// Initialization boilerplate
func initialize(config *configuration) {
	var bgImage, appSurface *sdl.Surface

	if sdl.Init(sdl.INIT_VIDEO) != 0 {
		panic(sdl.GetError())
	}

	if ttf.Init() != 0 {
		panic(sdl.GetError())
	}

	font := ttf.OpenFont(flag.Arg(0), 12)

	if font == nil {
		panic(sdl.GetError())
	}

	sdl.EnableUNICODE(1)

	if config.fullscreen {
		appSurface = sdl.SetVideoMode(640, 480, 32, sdl.FULLSCREEN)
		sdl.ShowCursor(sdl.DISABLE)
	} else {
		appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	}
	if config.bgImage != "" {
		bgImage = sdl.Load(config.bgImage)
	}

	sdlrenderer = clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, int(config.consoleW), int(config.consoleH), 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	if config.fps > 0 {
		sdlrenderer.FPSCh() <- config.fps
	}

	console = clingon.NewConsole(sdlrenderer, &ShellEvaluator{})
	console.SetPrompt("shell:$ ")
	console.GreetingText = "Welcome to the CLIngon shell!\n=============================\nPress F10 to toggle/untoggle\n\n"

	r = &renderer{
		config:         config,
		appSurface:     appSurface,
		cliSurface:     sdlrenderer.GetSurface(),
		bgImageSurface: bgImage,
	}
}

func main() {
	// Handle options
	help := flag.Bool("help", false, "Show usage")
	verbose := flag.Bool("verbose", false, "Verbose output")
	fullscreen := flag.Bool("fullscreen", false, "Go fullscreen!")
	fps := flag.Float("fps", clingon.DEFAULT_SDL_RENDERER_FPS, "Frames per second")
	autoFps := flag.Bool("auto-fps", false, "Automatically double the FPS ratio on animations")
	bgImage := flag.String("bg-image", "", "Background image file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "shell - A system shell based on CLIngon (Command Line INterface for Go Nerds\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "\tshell [options] <fontfile> \n\n")
		fmt.Fprintf(os.Stderr, "Options are:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help == true {
		flag.Usage()
		return
	}

	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	config = configuration{
		verbose:    *verbose,
		fps:        *fps,
                autoFps: *autoFps,
		fullscreen: *fullscreen,
		bgImage:    *bgImage,
		consoleX:   40,
		consoleY:   40,
		consoleW:   560,
		consoleH:   400,
	}

	initialize(&config)
	r.render(nil)

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

					if config.verbose {
						fmt.Printf("\n")
						fmt.Printf("%v: %v", e.Keysym.Sym, ": ", keyName)

						fmt.Printf("%04x ", e.Type)

						for i := 0; i < len(e.Pad0); i++ {
							fmt.Printf("%02x ", e.Pad0[i])
						}
						fmt.Printf("\n")

						fmt.Printf("Type: %02x Which: %02x State: %02x Pad: %02x\n", e.Type, e.Which, e.State, e.Pad0[0])
						fmt.Printf("Scancode: %02x Sym: %08x Mod: %04x Unicode: %04x\n", e.Keysym.Scancode, e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode)
					}
					if (keyName == "escape") && (e.Type == sdl.KEYDOWN) {
						running = false
					} else if (keyName == "f10") && (e.Type == sdl.KEYDOWN) {
						console.Paused = !console.Paused
						r.animateCLI = true
						if config.autoFps {
							sdlrenderer.FPSCh() <- config.fps * 2
						}
					} else if (keyName == "up") && (e.Type == sdl.KEYDOWN) {
						console.ReadlineCh() <- clingon.HISTORY_PREV
					} else if (keyName == "down") && (e.Type == sdl.KEYDOWN) {
						console.ReadlineCh() <- clingon.HISTORY_NEXT
					} else if (keyName == "left") && (e.Type == sdl.KEYDOWN) {
						console.ReadlineCh() <- clingon.CURSOR_LEFT
					} else if (keyName == "right") && (e.Type == sdl.KEYDOWN) {
						console.ReadlineCh() <- clingon.CURSOR_RIGHT
					} else {
						unicode := e.Keysym.Unicode
						if unicode > 0 {
							console.CharCh() <- unicode
						}
					}

				}
			}

		}
	}()

	for running {
		select {
		case rects := <-sdlrenderer.UpdatedRectsCh():
			r.render(rects)
		}
	}

	sdl.Quit()
}
