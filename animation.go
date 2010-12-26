package clingon

import (
	"time"
	"math"
)

const DEFAULT_ANIMATION_FPS = 30.0

const (
	ANIMATION_START = iota
	ANIMATION_STOP
	ANIMATION_PAUSE
)

type Animation struct {
	f          func(t int64) float64
	duration   int64
	valueCh    chan float64
	finishedCh chan bool
	paused     bool
	t          int64
}

func NewAnimation(f func(t int64) float64, duration int64) *Animation {
	animation := &Animation{
		f:          f,
		duration:   duration,
		valueCh:    make(chan float64),
		finishedCh: make(chan bool),
	}

	go animation.animate()

	return animation
}

func (animation *Animation) Start() {
	animation.paused = false
}

func (animation *Animation) Pause() int64 {
	animation.paused = true
	return animation.t
}

func (animation *Animation) Resume(t int64) {
	animation.paused = false
	animation.t = t
}

func (animation *Animation) ValueCh() <-chan float64 {
	return animation.valueCh
}

func (animation *Animation) FinishedCh() <-chan bool {
	return animation.finishedCh
}

func (animation *Animation) animate() {
	var (
		time_step = int64(1e9 / int64(DEFAULT_ANIMATION_FPS))
		ticker    = time.NewTicker(time_step)
	)

	animation.paused = true

	for {
		if !animation.paused {
			select {
			case <-ticker.C:
				if animation.t <= animation.duration {
					next_t := animation.t + time_step
					if next_t > animation.duration {
						animation.t = animation.duration
					}
					animation.valueCh <- animation.f(animation.t)
					animation.t = next_t
				} else {
					animation.t = 0
					animation.paused = true
					animation.finishedCh <- true
				}
			}
		} else {
			select {
			case <-ticker.C:
			}
		}
	}
}

func NewSlidingAnimation(duration int64, distance float64) *Animation {
	return NewAnimation(func(t int64) float64 {
		arc := (1 - float64(t)/float64(duration)) * (math.Pi / 2)
		return float64(distance) * (math.Cos(arc))
	},
		duration)
}
