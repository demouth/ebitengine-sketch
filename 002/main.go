package main

import (
	"fmt"
	"log"
	"math/rand"

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
	mainCharacter = NewApple(200, 200)

	fruits = []*Fruit{mainCharacter}
	world  = World{X: 10, Y: 10, Width: screenWidth - 20, Height: screenHeight - 20}

	calc = &Calc{World: world}
	draw = &Draw{}

	isKeyPressed = false
)

func init() {
	for _ = range 30 {
		fruits = append(
			fruits,
			NewApple(
				rand.Float64()*screenWidth,
				rand.Float64()*screenHeight,
			),
		)
	}
}

func (g *Game) Update() error {
	fruits = calc.Fruits(fruits)

	ac := 0.1
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		ac = 0.5
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		mainCharacter.VX -= ac
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		mainCharacter.VX += ac
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		mainCharacter.VY -= ac
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		mainCharacter.VY += ac
	}

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
