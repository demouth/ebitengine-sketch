package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 480
	screenHeight = 640
)

type Game struct {
}

var (
	fruits = []*Fruit{
		NewApple(100, 100),
		NewApple(130, 120),
		NewApple(150, 150),
	}
	world = World{X: 10, Y: 10, Width: screenWidth - 20, Height: screenHeight - 20}

	calc = &Calc{World: world}
	draw = &Draw{}

	isKeyPressed = false
)

func (g *Game) Update() error {
	fruits = calc.Fruits(fruits)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	draw.World(screen, world)
	draw.Fruits(screen, world, fruits)
	msg := fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("002")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
