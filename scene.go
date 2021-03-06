package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	img "github.com/veandco/go-sdl2/sdl_image"
	ttf "github.com/veandco/go-sdl2/sdl_ttf"
)

const (
	numberOfPipes = 100
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
	font     *ttf.Font
}

// NewScene creates new scene
func NewScene(r *sdl.Renderer) (*scene, error) {
	s := &scene{renderer: r}
	var err error

	s.font, err = ttf.OpenFont("resources/font.ttf", 32)

	if err != nil {
		return s, fmt.Errorf("Error while opening font: %v", err)
	}

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
	s.pipes.pipes = s.generatePipes()
}

func (s *scene) generatePipes() []*pipe {
	var pipes []*pipe
	var w int32 = 52
	gapW := w * 5

	for i := 0; i < numberOfPipes; i++ {
		limit := int32(windowHeight/2) - 50
		bottomPipeHeight := rand.Int31n(limit)
		topPipeHeight := rand.Int31n(limit)
		pos := gapW*int32(i+2) + rand.Int31n(w*4) + w/2

		if rand.Intn(10) > 4 {
			pos += w
		} else {
			pos -= w
		}

		pipes = append(pipes,
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

	return pipes

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

	if s.bgx >= windowWidth {
		s.bgx = 0
	}

	if s.pipes.hits(&s.bird) {
		s.bird.dead = true
	}

	if s.bird.dead {
		s.restart()
	}
}

func (s *scene) draw() {
	s.renderer.Clear()
	s.drawBg()
	s.bird.draw(s.renderer)
	s.pipes.draw(s.renderer)
	s.drawScore()

	s.renderer.Present()
}

func (s *scene) drawScore() {
	score := int32(s.score() / 2)
	scoreText := fmt.Sprintf("%d", score)
	surf, err := s.font.RenderUTF8_Solid(scoreText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	var width int32 = 30

	if score >= 10 {
		width += 30
	}

	if score >= 100 {
		width += 30
	}

	if err != nil {
		log.Fatalf("Error while rendering score: %v", err)
	}

	texture, err := s.renderer.CreateTextureFromSurface(surf)
	defer texture.Destroy()

	if err != nil {
		log.Fatalf("Could not create texture from surface: %v", err)
	}

	s.renderer.Copy(texture, nil, &sdl.Rect{X: 15, Y: 15, W: width, H: 60})
}

func (s *scene) drawBg() {
	x1 := -s.bgx
	x2 := windowWidth - s.bgx
	srcRect := sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}
	destRect1 := sdl.Rect{X: x1, Y: 0, W: windowWidth, H: windowHeight}
	destRect2 := sdl.Rect{X: x2, Y: 0, W: windowWidth, H: windowHeight}
	s.renderer.Copy(s.bg, &srcRect, &destRect1)
	s.renderer.Copy(s.bg, &srcRect, &destRect2)
}

func (s *scene) score() int {
	return s.pipes.score()
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
	dest := &sdl.Rect{X: p.pos, Y: windowHeight - p.h, W: p.w, H: p.h}
	src := &sdl.Rect{X: 0, Y: 0, W: p.w, H: p.h}

	flip := sdl.FLIP_NONE
	if !p.up {
		dest.Y = 0
		flip = sdl.FLIP_VERTICAL
	}

	r.CopyEx(tex, src, dest, 0, nil, flip)
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
