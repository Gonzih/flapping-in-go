package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowHeight = 600
	windowWidth  = 800
)

func run() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)

	if err != nil {
		return fmt.Errorf("Error while initing sdl: %v", err)
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
		case *sdl.MouseButtonEvent:
			scene.bird.jump()
		}
	}

	return nil
}

func main() {
	err := run()

	if err != nil {
		log.Fatal(err)
	}
}
