package main

import (
	"fmt"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	img "github.com/veandco/go-sdl2/sdl_image"
)

// scene stores current game state
type scene struct {
	renderer *sdl.Renderer
	bg       *sdl.Texture
	bird     bird
}

// NewScene creates new scene
func NewScene(r *sdl.Renderer) (*scene, error) {
	s := &scene{renderer: r}
	var err error

	s.bg, err = img.LoadTexture(s.renderer, "resources/bg.png")

	if err != nil {
		return s, fmt.Errorf("Error while loading bg: %v", err)
	}

	s.bird = bird{x: 20, y: windowHeight / 2, w: 50, h: 50, gravity: 1}

	for i := 1; i <= 4; i++ {
		t, err := img.LoadTexture(s.renderer, fmt.Sprintf("resources/bird/frame-%d.png", i))

		if err != nil {
			return s, fmt.Errorf("Error while loading bird texturne #%d: %v", i, err)
		}

		s.bird.frames = append(s.bird.frames, t)
	}

	return s, nil
}

func (s *scene) restart() {
	s.bird.dead = false
	s.bird.x = 0
	s.bird.y = windowHeight / 2
}

func (s *scene) run(fps uint32) {
	for {
		s.update()
		s.draw()
		sdl.Delay(uint32(1000 / fps))
	}
}

func (s *scene) update() {
	s.bird.update()

	if s.bird.dead {
		s.restart()
	}
}

func (s *scene) draw() {
	s.renderer.Clear()
	s.renderer.Copy(s.bg, &sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}, nil)

	s.bird.draw(s.renderer)

	s.renderer.Present()
}

// bird represents state of birdy
type bird struct {
	x, y, w, h     int32
	speed, gravity int32
	frames         []*sdl.Texture
	frame          int
	mu             sync.Mutex
	dead           bool
}

func (b *bird) update() {
	f := b.frame + 1

	if f > 3 {
		b.frame = 0
	} else {
		b.frame = f
	}

	b.x++

	b.mu.Lock()
	defer b.mu.Unlock()
	b.y += b.speed
	b.speed += b.gravity

	if b.y > windowHeight || b.y < 0 {
		b.dead = true
	}
}

func (b *bird) draw(r *sdl.Renderer) {
	r.Copy(b.frames[b.frame], nil, &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
}

func (b *bird) jump() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.speed = -10
}
