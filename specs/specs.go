package specs

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"time"
	pt "spectrum/prettytest"
	c "console"
)

const (
	KEY_RETURN = iota
	KEY_UP
	KEY_DOWN
)

var (
	console *c.Console
	echoer Echoer
	sdlrenderer *c.SDLRenderer
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
	console *c.Console
	command string
	time int64
}

func NewEnterCommand(console *c.Console, command string, time int64) *EnterCommand {
	return &EnterCommand{console, command + "\u000d", time}
}

func ExecuteEnterCommand(console *c.Console, command string, time int64) (done bool) {
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
	console *c.Console
	dirs []int
	time int64
}

func NewBrowseHistory(console *c.Console, dirs []int, time int64) *BrowseHistory {
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
	console *c.Console
	dirs []int
	time int64
}
func NewMoveCursor(console *c.Console, dirs []int, time int64) *MoveCursor {
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
	appSurface.FillRect(&sdl.Rect{0, 0, 640, 480}, 0)
	appSurface.Blit(&sdl.Rect{195, 115, 0, 0}, gopher, nil)
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

	sdlrenderer = c.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, 640, 480, 32, 0, 0, 0, 0), font)
	sdlrenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xaa)

	console = c.NewConsole(sdlrenderer, &Echoer{})

	render()
}

func after(t *pt.T) {
	sdl.Quit()
}

