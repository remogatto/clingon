package main

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"fmt"
	"flag"
	"os"
	"clingon"
)

var (
	config                   configuration
	console                  *clingon.Console
	sdlrenderer              *clingon.SDLRenderer
	running, toggleAnimation bool
	r                        *renderer
	slideDown, slideUp       *clingon.Animation
)

type configuration struct {
	consoleX, consoleY  int16
	consoleW, consoleH  uint16
	fullscreen, verbose bool
	fps                 float
	bgImage             string
	animationDuration   int64
}

type renderer struct {
	config                                 *configuration
	appSurface, bgImageSurface, cliSurface *sdl.Surface
}

const ANIMATION_TIME = 500 * 1e6

func (r *renderer) render(updatedRects []sdl.Rect, y int16) {
	if updatedRects == nil { // Initially and during animations we blit the entire surface
		if r.bgImageSurface != nil {
			r.appSurface.Blit(nil, r.bgImageSurface, nil)
		}
		r.appSurface.Blit(&sdl.Rect{config.consoleX, y, 0, 0}, sdlrenderer.GetSurface(), nil)
		r.appSurface.Flip()
	} else { // When idle we can keep updating modified regions only
		for _, rect := range updatedRects {
			if r.bgImageSurface != nil {
				r.appSurface.Blit(
					&sdl.Rect{rect.X + r.config.consoleX, rect.Y + y, 0, 0},
					r.bgImageSurface,
					&sdl.Rect{rect.X + r.config.consoleX, rect.Y + y, rect.W, rect.H})
			}
			r.appSurface.Blit(
				&sdl.Rect{rect.X + r.config.consoleX, rect.Y + y, 0, 0},
				sdlrenderer.GetSurface(), &rect)
			r.appSurface.UpdateRect(int32(rect.X+r.config.consoleX), int32(rect.Y+y), uint32(rect.W), uint32(rect.H))
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

	slidingDistance := int16(appSurface.H) - config.consoleY
	slideDown = clingon.NewSlideDownAnimation(config.animationDuration, float64(slidingDistance))
	slideUp = clingon.NewSlideUpAnimation(config.animationDuration, float64(slidingDistance))
}

func main() {
	// Handle options
	help := flag.Bool("help", false, "Show usage")
	verbose := flag.Bool("verbose", false, "Verbose output")
	fullscreen := flag.Bool("fullscreen", false, "Go fullscreen!")
	fps := flag.Float("fps", clingon.DEFAULT_CONSOLE_RENDERER_FPS, "Frames per second")
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
		verbose:           *verbose,
		fps:               *fps,
		fullscreen:        *fullscreen,
		bgImage:           *bgImage,
		consoleX:          40,
		consoleY:          40,
		consoleW:          560,
		consoleH:          400,
		animationDuration: 1e9,
	}

	initialize(&config)
	r.render(nil, config.consoleY)

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
						toggleAnimation = !toggleAnimation
						console.Paused = toggleAnimation
						if console.Paused {
							t := slideUp.Pause()
							slideDown.Resume(t)
						} else {
							t := slideDown.Pause()
							if t == 0 {
								slideUp.Start()
							} else {
								slideUp.Resume(1e9 - t)
							}
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

	var (
		y int16 = config.consoleY
//		animating bool
	)
	
	for running {
		select {
		case value := <-slideDown.ValueCh():
			y = 40 + int16(value)
			r.render(nil, y)
		case value := <-slideUp.ValueCh():
			y = 40 + int16(value)
			r.render(nil, y)
		case <-slideDown.FinishedCh():
		case <-slideUp.FinishedCh():
		case rects := <-sdlrenderer.UpdatedRectsCh():
			r.render(rects, y)
		}
	}

	sdl.Quit()
}
