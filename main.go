package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
)

const (
	windowHeight = 600
	windowWidth  = 800
)

func run() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)

	if err != nil {
		return fmt.Errorf("Error while initializing sdl: %v", err)
	}

	err = ttf.Init()

	if err != nil {
		return fmt.Errorf("Error while initializing ttf: %v", err)
	}

	defer sdl.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer(windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	defer window.Destroy()

	if err != nil {
		return fmt.Errorf("Error while creating window: %v", err)
	}

	scene, err := NewScene(renderer)

	if err != nil {
		return fmt.Errorf("Error while creating scene: %v", err)
	}

	go scene.run(20)

loop:
	for {
		switch sdl.WaitEvent().(type) {
		case *sdl.QuitEvent:
			break loop
		case *sdl.MouseButtonEvent, *sdl.KeyUpEvent:
			scene.bird.jump()
		}
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	err := run()

	if err != nil {
		log.Fatal(err)
	}
}
