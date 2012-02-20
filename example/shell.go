package main

import (
	"flag"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/ttf"
	"github.com/remogatto/clingon"
	"os"
)

var (
	sdlrenderer  *clingon.SDLRenderer
	greetingText = `
Welcome to the CLIngon shell!
=============================

Available keys:

* F10 toggle/untoggle
* PageUp/PageDown for scrolling
* ESC exit
`
)

type configuration struct {
	consoleX, consoleY  int16
	consoleW, consoleH  uint16
	fullscreen, verbose bool
	bgImage             string
	animationLength     float64 // In seconds
}

type renderer struct {
	config                     *configuration
	sdlRenderer                clingon.Renderer
	appSurface, bgImageSurface *sdl.Surface
}

// Animation length, in seconds
const ANIMATION_LENGTH = 0.750

func (r *renderer) render(updatedRects []sdl.Rect, y int16) {
	if updatedRects == nil { // Initially and during animations we blit the entire surface
		if r.bgImageSurface != nil {
			r.appSurface.Blit(nil, r.bgImageSurface, nil)
		}
		r.appSurface.Blit(&sdl.Rect{r.config.consoleX, y, 0, 0}, sdlrenderer.GetSurface(), nil)
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
func initialize(config *configuration) *renderer {
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
		flags := sdl.FULLSCREEN
		appSurface = sdl.SetVideoMode(640, 480, 32, uint32(flags))
		sdl.ShowCursor(sdl.DISABLE)
	} else {
		appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	}
	if config.bgImage != "" {
		bgImage = sdl.Load(config.bgImage)
	}

	sdlrenderer = clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, int(config.consoleW), int(config.consoleH), 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	return &renderer{
		config:         config,
		appSurface:     appSurface,
		sdlRenderer:    sdlrenderer,
		bgImageSurface: bgImage,
	}
}

func main() {
	// Handle options
	help := flag.Bool("help", false, "Show usage")
	verbose := flag.Bool("verbose", false, "Verbose output")
	fullscreen := flag.Bool("fullscreen", false, "Go fullscreen!")
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

	config := configuration{
		verbose:         *verbose,
		fullscreen:      *fullscreen,
		bgImage:         *bgImage,
		consoleX:        40,
		consoleY:        40,
		consoleW:        560,
		consoleH:        400,
		animationLength: ANIMATION_LENGTH,
	}

	r := initialize(&config)

	y_channel := make(chan int16)
	running := true

	// Have to start this *before* printing to the console,
	// because the printing will cause some messages to be sent to 'sdlrenderer.UpdatedRectsCh()'
	go func() {
		var y int16 = config.consoleY
		for running {
			select {
			case newY := <-y_channel:
				y = newY
				r.render(nil, y)
			case rects := <-sdlrenderer.UpdatedRectsCh():
				r.render(rects, y)
			}
		}
	}()

	var console *clingon.Console = clingon.NewConsole(&ShellEvaluator{})
	console.SetRenderer(r.sdlRenderer)
	y_channel <- config.consoleY
	console.Print(greetingText)
	console.SetPrompt("shell:$ ")

	var anim *clingon.Animation = nil
	var animValueCh <-chan float64 = nil
	toggleAnimation := false

	for running {

		select {
		case event := <-sdl.Events:
			switch e := event.(type) {
			case sdl.QuitEvent:
				running = false
			case sdl.MouseMotionEvent:
				var y int
				state := sdl.GetRelativeMouseState(nil, &y)
				if state == 1 {
					if y < 0 {
						sdlrenderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_UP}
					} else if y > 0 {
						sdlrenderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_DOWN}

					}
				}
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
					if anim != nil {
						anim.Terminate()
						<-anim.FinishedCh()
						anim = nil
					}

					toggleAnimation = !toggleAnimation

					slidingDistance := int16(r.appSurface.H) - config.consoleY
					anim = clingon.NewSliderAnimation(config.animationLength, float64(slidingDistance))

					animValueCh = anim.ValueCh()
					anim.Start()

				} else if (keyName == "page up") && (e.Type == sdl.KEYDOWN) {
					sdlrenderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_UP}
				} else if (keyName == "page down") && (e.Type == sdl.KEYDOWN) {
					sdlrenderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_DOWN}
				} else if (keyName == "up") && (e.Type == sdl.KEYDOWN) {
					console.PutReadline(clingon.HISTORY_PREV)
				} else if (keyName == "down") && (e.Type == sdl.KEYDOWN) {
					console.PutReadline(clingon.HISTORY_NEXT)
				} else if (keyName == "left") && (e.Type == sdl.KEYDOWN) {
					console.PutReadline(clingon.CURSOR_LEFT)
				} else if (keyName == "right") && (e.Type == sdl.KEYDOWN) {
					console.PutReadline(clingon.CURSOR_RIGHT)
				} else {
					unicode := e.Keysym.Unicode
					if unicode > 0 {
						console.PutUnicode(unicode)
					}
				}
			}

		case value := <-animValueCh:
			var y float64
			if toggleAnimation {
				y = 40 + value
			} else {
				y = float64(r.appSurface.H) - value
			}

			y_channel <- int16(y)
		}

	} // End of "for running"

	sdl.Quit()
}
