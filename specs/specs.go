package specs

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"time"
	"clingon"
	"os"
)

const (
	KEY_RETURN = iota
	KEY_UP
	KEY_DOWN
)

var (
	console                                *clingon.Console
	echoer                                 Echoer
	sdlrenderer                            *clingon.SDLRenderer
	appSurface, gopher                     *sdl.Surface
	appSurfaceW, appSurfaceH               = 640, 480
	consoleX, consoleY, consoleW, consoleH = int16(40), int16(40), uint16(560), uint16(400)
)

type Echoer struct{}

func (e *Echoer) Run(console *clingon.Console, command string) os.Error {
	console.Print(command)
	return nil
}

type Interactor interface {
	Interact(done chan bool)
}

type EnterCommand struct {
	console *clingon.Console
	command string
	time    int64
}

func NewEnterCommand(console *clingon.Console, command string, time int64) *EnterCommand {
	return &EnterCommand{console, command + "\u000d", time}
}

func (i *EnterCommand) Interact(done chan bool) {
	for _, c := range i.command {
		console.CharCh() <- uint16(c)
		if i.time > 0 {
			time.Sleep(i.time)
		}
	}
	done <- true
}

type BrowseHistory struct {
	console *clingon.Console
	dirs    []int
	time    int64
}

func NewBrowseHistory(console *clingon.Console, dirs []int, time int64) *BrowseHistory {
	return &BrowseHistory{console, dirs, time}
}
func (i *BrowseHistory) Interact(done chan bool) {
	for _, dir := range i.dirs {
		console.ReadlineCh() <- dir
		if i.time > 0 {
			time.Sleep(i.time)
		}
	}
	done <- true
}

type MoveCursor struct {
	console *clingon.Console
	dirs    []int
	time    int64
}

func NewMoveCursor(console *clingon.Console, dirs []int, time int64) *MoveCursor {
	return &MoveCursor{console, dirs, time}
}
func (i *MoveCursor) Interact(done chan bool) {
	for _, dir := range i.dirs {
		console.ReadlineCh() <- dir
		if i.time > 0 {
			time.Sleep(i.time)
		}
	}
	done <- true
}

func Interact(interactions []Interactor) (done bool) {
	doneCh := make(chan bool)
	for _, i := range interactions {
		go i.Interact(doneCh)
	L:
		for {
			select {
			case <-doneCh:
				break L
			}
		}
	}
	return true
}

func initTest() {
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

	appSurface = sdl.SetVideoMode(appSurfaceW, appSurfaceH, 32, 0)
	gopher = sdl.Load("../testdata/gopher.jpg")

	sdlrenderer = clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, int(consoleW), int(consoleH), 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	console = clingon.NewConsole(sdlrenderer, &Echoer{})
	console.GreetingText = "Welcome to the CLIngon shell!\n\n"

	render(nil)

	go func() {
		for {
			select {
			case rects := <-sdlrenderer.UpdatedRectsCh():
				render(rects)
			}

		}
	}()
}

func render(updatedRects []sdl.Rect) {
	if updatedRects == nil {
		appSurface.Blit(nil, gopher, nil)
		appSurface.Flip()
	} else {
		for _, r := range updatedRects {
			appSurface.Blit(&sdl.Rect{r.X + consoleX, r.Y + consoleY, 0, 0}, gopher, &sdl.Rect{r.X + consoleX, r.Y + consoleY, r.W, r.H})
			appSurface.Blit(&sdl.Rect{consoleX, consoleY, 0, 0}, sdlrenderer.GetSurface(), &sdl.Rect{0, 0, consoleW, consoleH})
			appSurface.UpdateRect(int32(r.X+consoleX), int32(r.Y+consoleY), uint32(r.W), uint32(r.H))
		}
	}
}
