package specs

import (
	"⚛sdl"
	"testing"
	pt "prettytest"
	"fmt"
	"clingon"
	"time"
)

type specsSuite struct {
	pt.Suite
}

func (s *specsSuite) beforeAll() {
	initTest()
}

func (s *specsSuite) afterAll() {
	// Request termination of the rendering loop,
	// and wait until it actually terminates
	terminateRendering <- 0
	<-renderingLoopTerminated

	sdl.Quit()
}

func (s *specsSuite) should_receive_commands() {
	s.True(Interact([]Interactor{NewEnterCommand(console, "", 2e8)}))
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
		},
		2e8)}))
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
	},
		2e8)}))
}

func (s *specsSuite) should_slidedown_slideup() {
	var anim *clingon.Animation

	for i := 1; i <= 3; i++ {
		anim = clingon.NewSliderAnimation(1.0, distance)
		animationCh <- NewAnimation{anim, clingon.SCROLL_DOWN}
		anim.Start()
		s.True(<-anim.FinishedCh())

		anim = clingon.NewSliderAnimation(1.0, distance)
		animationCh <- NewAnimation{anim, clingon.SCROLL_UP}
		anim.Start()
		s.True(<-anim.FinishedCh())
	}
}

func (s *specsSuite) should_scroll_up_down() {
	for i := 0; i < 40; i++ {
		s.True(Interact([]Interactor{NewEnterCommand(console, fmt.Sprintf("Line %d", i), 0)}))
	}
	renderer := console.RendererOrNil().(*clingon.SDLRenderer)
	renderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_UP}
	time.Sleep(1e9)
	renderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_UP}
	time.Sleep(1e9)
	renderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_DOWN}
	time.Sleep(1e9)
	renderer.EventCh() <- clingon.Cmd_Scroll{clingon.SCROLL_DOWN}
	time.Sleep(1e9)
	s.True(true)
}

func (s *specsSuite) should_pause_unpause() {
	renderer := console.RendererOrNil().(*clingon.SDLRenderer)

	console.Print("Pause the console")
	console.SetRenderer(nil)
	time.Sleep(2e9)
	console.Print("Unpause the console")
	console.SetRenderer(renderer)
	time.Sleep(2e9)
	s.True(true)
}

func (s *specsSuite) should_set_a_new_renderer() {
	console.Print("Cursor should blink before switching to the new renderer")
	time.Sleep(2e9)

	// Request termination of the current rendering loop,
	// and wait until it actually terminates
	terminateRendering <- 0
	<-renderingLoopTerminated

	newRenderer := clingon.NewSDLRenderer(sdl.CreateRGBSurface(sdl.SRCALPHA, int(consoleW), int(consoleH), 32, 0, 0, 0, 0), font)
	newRenderer.GetSurface().SetAlpha(sdl.SRCALPHA, 0xdd)
	console.SetRenderer(newRenderer)
	newRenderingLoop(newRenderer)

	console.Print("Cursor should blink after switching to the new renderer")
	time.Sleep(2e9)
	s.True(true)
}

func TestConsoleSpecs(t *testing.T) {
	pt.RunWithFormatter(
		t,
		&pt.BDDFormatter{"Console"},
		new(specsSuite),
	)
}
