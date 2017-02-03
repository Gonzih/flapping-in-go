package main

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	img "github.com/veandco/go-sdl2/sdl_image"
)

const (
	numberOfPipes = 50
	pipesSpeed    = 10
	birdInitX     = 20
)

// ================== SCENE ================== //

type scene struct {
	renderer *sdl.Renderer
	bg       *sdl.Texture
	bgx      int32
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

	s.bird = bird{x: birdInitX, y: windowHeight / 2, w: 50, h: 50, gravity: 5}

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

	s.resetPipes()
	return s, nil
}

func (s *scene) resetPipes() {
	s.pipes.pipes = nil
	var w int32 = 52
	gapW := w * 5

	for i := 0; i < numberOfPipes; i++ {
		gapH := int32(rand.Intn(windowHeight/2) + 150)
		bottomPipeHeight := int32(rand.Intn(windowHeight - int(gapH)))
		topPipeHeight := windowHeight - gapH - bottomPipeHeight
		pos := gapW*int32(i+2) + rand.Int31n(gapW)

		fmt.Printf("gapH %d, bottomHeight %d, topHeight %d, sum %d \n", gapH, bottomPipeHeight, topPipeHeight, gapH+bottomPipeHeight+topPipeHeight)

		s.pipes.pipes = append(s.pipes.pipes,
			&pipe{
				pos: pos,
				w:   w,
				h:   bottomPipeHeight,
				up:  false,
			},
			&pipe{
				pos: pos,
				w:   w,
				h:   topPipeHeight,
				up:  true,
			})
	}

}

func (s *scene) restart() {
	s.score()
	s.bird.dead = false
	s.bird.x = birdInitX
	s.bird.y = windowHeight / 2
	s.bgx = 0
	s.resetPipes()
	s.draw()
	sdl.Delay(500)
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
	s.bgx = (s.bgx + 1) % 2000

	if s.pipes.hits(&s.bird) {
		s.bird.dead = true
	}

	if s.bird.dead {
		s.restart()
	}
}

func (s *scene) draw() {
	s.renderer.Clear()
	s.renderer.Copy(s.bg, &sdl.Rect{X: s.bgx, Y: 0, W: windowWidth, H: windowHeight}, nil)

	s.bird.draw(s.renderer)
	s.pipes.draw(s.renderer)

	s.renderer.Present()
}

func (s *scene) score() {
	fmt.Printf("Score was: %d\n", s.pipes.score())
}

// ================== BIRD ================== //

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
	b.speed = -20
}

// ================== PIPES ================== //

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

func (pp *pipes) score() int {
	score := 0

	for _, p := range pp.pipes {
		if p.pos < birdInitX {
			score++
		}
	}

	return score
}

// ================== PIPE ================== //

type pipe struct {
	w, h, pos int32
	up        bool
}

func (p *pipe) update() {
	p.pos -= pipesSpeed
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
