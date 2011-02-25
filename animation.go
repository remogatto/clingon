package clingon

import (
	"time"
	"math"
)

const DEFAULT_ANIMATION_FPS = 30

type Animation struct {
	f           func(t float64) float64
	length      float64 // Length, in seconds
	t           float64 // Time, in seconds
	dt          float64 // Time step, in seconds
	valueCh     chan float64
	finishedCh  chan bool
	terminateCh chan bool // Sending a value to this channel terminates the animation

	// This is used to ensure correctness
	running bool
}

// Creates a new animation.
// The values of 't' and 'length' are in seconds.
// Call 'Animation.Start' to start the animation.
func NewAnimation(f func(t float64) float64, length float64) *Animation {
	if length <= 0 {
		panic("non-positive length")
	}

	animation := &Animation{
		f:           f,
		length:      length,
		t:           0,
		dt:          length / DEFAULT_ANIMATION_FPS,
		valueCh:     make(chan float64, 1),
		finishedCh:  make(chan bool, 1),
		terminateCh: make(chan bool, 1),
		running:     false,
	}

	return animation
}

// Starts the animation.
// This method can be called at most once per an Animation object.
func (animation *Animation) Start() {
	if !animation.running {
		// Note: the "+1" ensures that the argument to NewTicker is >= 1
		ticker := time.NewTicker(int64(1e9*animation.dt) + 1)
		go animation.animate(ticker)

		animation.running = true
	} else {
		panic("already started")
	}
}

// Tells the animation that it should terminate.
// This method returns immediately.
// It is safe to call this method even if the animation has ended on its own.
func (animation *Animation) Terminate() {
	if !animation.running {
		panic("not started")
	}

	// The channel has a buffer with size 1.
	// Only the 1st value ever sent through the channel is significant,
	// subsequent attempts to send another value can be ignored.
	select {
	case animation.terminateCh <- true: // Non-blocking send
	default:
	}
}

// The channel through which the animation delivers its consequetive values.
// This channel is never closed.
func (animation *Animation) ValueCh() <-chan float64 {
	return animation.valueCh
}

// This channel will receive a single value [when the animation ends]
// or [after the animation has been terminated]
// This channel is never closed.
func (animation *Animation) FinishedCh() <-chan bool {
	return animation.finishedCh
}

func (animation *Animation) animate(ticker *time.Ticker) {
	// Send the 1st value
	animation.valueCh <- animation.f(animation.t)

loop:
	for {
		select {
		case <-ticker.C:
			animation.t += animation.dt

			t := animation.t
			if t > animation.length {
				t = animation.length
			}

			// Non-blocking send, the channel has a buffer with size 1
			select {
			case animation.valueCh <- animation.f(t):
				// Sent successfully
			default:
				// The value-receiver has not processed the previous value,
				// the current value is silently discarded
			}

			// End of animation
			if animation.t >= animation.length {
				break loop
			}

		case <-animation.terminateCh:
			break loop
		}
	}

	ticker.Stop()
	animation.finishedCh <- true
}

// Convenience function that creates a new animation going from 0.0 to 'distance'.
// The length is specified in seconds.
func NewSliderAnimation(length float64, distance float64) *Animation {
	f := func(t float64) float64 {
		arc := t / length * (math.Pi / 2)
		return distance * math.Pow(math.Sin(arc), 0.8)
	}
	return NewAnimation(f, length)
}
