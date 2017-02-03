package main

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	img "github.com/veandco/go-sdl2/sdl_image"
)

type scene struct {
	renderer *sdl.Renderer
	bg       *sdl.Texture
	bird     bird
	pipes    pipes
}

// NewScene creates new scene
func NewScene(r *sdl.Renderer) (*scene, error) {
	s := &scene{renderer: r}
	var err error

	s.bg, err = img.LoadTexture(s.renderer, "resources/bg.png")

	if err != nil {
		return s, fmt.Errorf("Error while loading bg: %v", err)
	}

	s.bird = bird{x: 50, y: windowHeight / 2, w: 50, h: 50, gravity: 1}

	for i := 1; i <= 4; i++ {
		t, err := img.LoadTexture(s.renderer, fmt.Sprintf("resources/bird/frame-%d.png", i))

		if err != nil {
			return s, fmt.Errorf("Error while loading bird texturne #%d: %v", i, err)
		}

		s.bird.frames = append(s.bird.frames, t)
	}

	s.pipes = pipes{}

	s.pipes.tex, err = img.LoadTexture(s.renderer, "resources/pipe.png")

	if err != nil {
		return s, fmt.Errorf("Error while loading pipe texture %v", err)
	}

	for i := 0; i < 10; i++ {
		s.pipes.pipes = append(s.pipes.pipes, &pipe{
			pos: windowWidth/2 + int32(rand.Intn(2*windowWidth)),
			w:   52,
			h:   int32(rand.Intn(windowHeight / 2)),
			up:  rand.Intn(10) > 4,
		})
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
	s.pipes.update()

	if s.pipes.hits(&s.bird) {
		s.bird.dead = true
	}

	if s.bird.dead {
		s.restart()
	}
}

func (s *scene) draw() {
	s.renderer.Clear()
	s.renderer.Copy(s.bg, &sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}, nil)

	s.bird.draw(s.renderer)
	s.pipes.draw(s.renderer)

	s.renderer.Present()
}

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

type pipes struct {
	pipes []*pipe
	tex   *sdl.Texture
}

func (pp *pipes) update() {
	for _, p := range pp.pipes {
		p.update()
	}
}

func (pp *pipes) draw(r *sdl.Renderer) {
	for _, p := range pp.pipes {
		p.draw(r, pp.tex)
	}
}

func (pp *pipes) hits(b *bird) bool {
	for _, p := range pp.pipes {
		hits := p.hits(b)
		if hits {
			return true
		}
	}

	return false
}

type pipe struct {
	w, h, pos int32
	up        bool
}

func (p *pipe) update() {
	p.pos--
}

func (p *pipe) draw(r *sdl.Renderer, tex *sdl.Texture) {
	rect := &sdl.Rect{X: p.pos, Y: windowHeight - p.h, W: p.w, H: p.h}

	flip := sdl.FLIP_NONE
	if !p.up {
		rect.H = p.h
		rect.Y = 0
		flip = sdl.FLIP_VERTICAL
	}

	r.CopyEx(tex, nil, rect, 0, nil, flip)
}

func (p *pipe) hits(b *bird) bool {
	if b.x+b.w <= p.pos || b.x >= p.pos+p.w {
		return false
	}

	if !p.up {
		return b.y <= p.h
	}

	return b.y+b.h >= windowHeight-p.h
}
