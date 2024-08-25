package main

import (
	"image/color"
	_ "image/png"

	"github.com/demouth/ebitengine-sketch/014/minigui"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

type Game struct {
	gui *minigui.GUI
}

func (g *Game) Update() error {
	g.gui.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0x66, 0x66, 0x66, 0xff})
	g.gui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}

	gui := minigui.NewGUI()
	gui.X = screenWidth
	gui.HorizontalAlign = minigui.HorizontalAlignRight

	var default64, min64, max64 float64 = 0.05, -0.1, 0.1
	gui.AddSliderFloat64("float64", default64, min64, max64, func(v float64) {
		// v is the value of the slider
	})

	var default32, min32, max32 float32 = 10.0, -5.0, 20.0
	gui.AddSliderFloat32("float32", default32, min32, max32, func(v float32) {
		// v is the value of the slider
	})

	var defaultInt, minInt, maxInt int = 0, -500, 500
	gui.AddSliderInt("int", defaultInt, minInt, maxInt, func(v int) {
		// v is the value of the slider
	})

	var defaultBool bool = true
	gui.AddButton("button", defaultBool, func(v bool) {
		// v is the value of the button
	})
	game.gui = gui

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ebiten-lil-gui")
	ebiten.RunGame(game)
}
