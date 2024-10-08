package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 480
	screenHeight = 480
)

type Game struct {
	touchIDs []ebiten.TouchID
}

type Drawer interface {
	World(screen *ebiten.Image, world World)
	Fruit(screen *ebiten.Image, world World, f *Fruit)
	Fruits(screen *ebiten.Image, world World, fruits []*Fruit)
}

var (
	mainCharacter = NewApple(screenWidth/2, screenHeight/2)

	fruits = []*Fruit{mainCharacter}
	world  = World{X: 10, Y: 10, Width: screenWidth - 20, Height: screenHeight - 20}

	calc = &Calc{World: world}

	drawer Drawer

	isKeyPressed = false
)

func init() {
	for _ = range 40 {
		fruits = append(
			fruits,
			NewApple(
				rand.Float64()*screenWidth,
				rand.Float64()*screenHeight,
			),
		)
	}

	drawer = &SpriteDrawer{}
}

func (g *Game) Update() error {
	// touch devices
	{
		g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
		hh := float64(screenHeight) / 2
		hw := float64(screenWidth) / 2
		for _, id := range g.touchIDs {
			if id > 0 {
				break
			}
			x, y := ebiten.TouchPosition(id)
			mainCharacter.VX += (float64(x) - hw) / hw * 1
			mainCharacter.VY += (float64(y) - hh) / hh * 1
		}
		if len(g.touchIDs) > 1 {
			drawer = &Draw{}
		} else {
			drawer = &SpriteDrawer{}
		}
	}
	// mouse

	if inpututil.MouseButtonPressDuration(ebiten.MouseButtonLeft) > 0 {
		hh := float64(screenHeight) / 2
		hw := float64(screenWidth) / 2
		x, y := ebiten.CursorPosition()
		mainCharacter.VX += (float64(x) - hw) / hw * 1
		mainCharacter.VY += (float64(y) - hh) / hh * 1
	}

	fruits = calc.Fruits(fruits)

	ac := 0.2
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		ac = 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		drawer = &SpriteDrawer{}
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		drawer = &Draw{}
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
	drawer.World(screen, world)
	drawer.Fruits(screen, world, fruits)
	msg := fmt.Sprintf(
		"Arrow keys: move character\n"+
			"Space keys: move fast\n"+
			"A key: Draw a character\n"+
			"S key: Draw an apple\n"+
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
