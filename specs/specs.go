package specs

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"time"
	pt "prettytest"
	"clingon"
)

const (
	KEY_RETURN = iota
	KEY_UP
	KEY_DOWN
)

var (
	console *clingon.Console
	echoer Echoer
	sdlrenderer *clingon.SDLRenderer
	appSurface, gopher *sdl.Surface
)

type Echoer struct {}

func (*Echoer) Run(command string) string {
	return command
}

type Interactor interface {
	Interact(done chan bool)
}

type EnterCommand struct {
	console *clingon.Console
	command string
	time int64
}

func NewEnterCommand(console *clingon.Console, command string, time int64) *EnterCommand {
	return &EnterCommand{console, command + "\u000d", time}
}

func ExecuteEnterCommand(console *clingon.Console, command string, time int64) (done bool) {
	doneCh := make(chan bool)
	go (&EnterCommand{console, command + "\u000d", time}).Interact(doneCh)
	for {
		select {
		case done = <-doneCh: return
		default: render()
		}
	}
	return
}
func (i *EnterCommand) Interact(done chan bool) {
	for _, c := range i.command {
		console.CharCh() <- uint16(c)
		time.Sleep(i.time)
	}
	done <- true
}

type BrowseHistory struct {
	console *clingon.Console
	dirs []int
	time int64
}

func NewBrowseHistory(console *clingon.Console, dirs []int, time int64) *BrowseHistory {
	return &BrowseHistory{console, dirs, time}
}
func (i *BrowseHistory) Interact(done chan bool) {
	for _, dir := range i.dirs {
		console.HistoryCh() <- dir
		time.Sleep(i.time)
	}
	done <- true
}

type MoveCursor struct {
	console *clingon.Console
	dirs []int
	time int64
}
func NewMoveCursor(console *clingon.Console, dirs []int, time int64) *MoveCursor {
	return &MoveCursor{console, dirs, time}
}
func (i *MoveCursor) Interact(done chan bool) {
	for _, dir := range i.dirs {
		console.CursorCh() <- dir
		time.Sleep(i.time)
	}
	done <- true
}

func Interact(interactions []Interactor) (done bool) {
	doneCh := make(chan bool)
	for _, i := range interactions {
		go i.Interact(doneCh)
	L: for {
			select {
			case done = <-doneCh: break L
			default: render()
			}
		}
	}
	return done
}

func render() {
	appSurface.Blit(&sdl.Rect{0, 0, 0, 0}, gopher, nil)
	appSurface.Blit(&sdl.Rect{40, 40, 0, 0}, sdlrenderer.GetSurface(), &sdl.Rect{0,0, 560, 400})
	appSurface.UpdateRect(40, 40, 560, 400)
}

func beforeAll(t *pt.T) {
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

	appSurface = sdl.SetVideoMode(640, 480, 32, 0)
	gopher = sdl.Load("../testdata/gopher.jpg")

	sdlrenderer = clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 560, 400, 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	console = clingon.NewConsole(sdlrenderer, &Echoer{})

	render()
	appSurface.Flip()
}

func afterAll(t *pt.T) {
	sdl.Quit()
}

