package specs

import (
	"⚛sdl"
	"⚛sdl/ttf"
	"testing"
	pt "prettytest"
	"clingon"
)

type specsSuite struct { pt.Suite }

func (s *specsSuite) beforeAll() {
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

func (s *specsSuite) afterAll() {
	sdl.Quit()
}

func (s *specsSuite) should_receive_commands() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "Hello gopher!", 2e8)}))
}

func (s *specsSuite) should_browse_history() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "foo", 2e8)}))
	s.True(Interact([]Interactor{NewEnterCommand(console, "bar", 2e8)}))
	s.True(Interact([]Interactor{NewEnterCommand(console, "biz", 2e8)}))
	s.True(Interact([]Interactor{NewBrowseHistory(console, 
			[]int{
			clingon.HISTORY_PREV, 
			clingon.HISTORY_PREV, 
			clingon.HISTORY_PREV, 
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
			clingon.HISTORY_NEXT,
		}, 2e8)}))
}

func (s *specsSuite) should_move_cursor() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "fooooooooooooooooo!!!!", 2e8)}))
	s.True(Interact([]Interactor{NewBrowseHistory(console, []int{clingon.HISTORY_PREV}, 2e8)}))
	s.True(Interact([]Interactor{NewMoveCursor(console, []int{
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_LEFT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
			clingon.CURSOR_RIGHT,
		}, 2e8)}))
}

func TestConsoleSpecs(t *testing.T) {
	pt.RunWithFormatter(
		t,
		&pt.BDDFormatter{"Console"},
		new(specsSuite),
	)
}
