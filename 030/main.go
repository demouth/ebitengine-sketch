package main

import (
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"math"

	"github.com/demouth/ebitengine-sketch/030/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 500
	screenHeight = 500
)

var (
	whiteImage = ebiten.NewImage(2, 2)
)

func init() {
	whiteImage.Fill(color.White)
}

type Game struct {
	countter int
}

func (g *Game) Update() error {
	g.countter++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	var path vector.Path
	path.MoveTo(250, 100)

	drawer.DrawCircle(screen, 250, 100, 10, color.RGBA{0xff, 0x00, 0x00, 0xff})
	drawer.DrawLine(screen, 250, 100, 350, 100, 10, color.RGBA{0xff, 0x00, 0x00, 0xff})

	const cx, cy, r = 350, 100, 70
	theta1 := math.Pi * float64(g.countter) / 180
	x := cx + r*math.Cos(theta1)
	y := cy + r*math.Sin(theta1)
	path.ArcTo(350, 100, float32(x), float32(y), 30)
	path.LineTo(float32(x), float32(y))
	drawer.DrawCircle(screen, 350, 100, 10, color.RGBA{0xff, 0x00, 0x00, 0xff})
	drawer.DrawCircle(screen, float32(x), float32(y), 10, color.RGBA{0xff, 0x00, 0x00, 0xff})
	drawer.DrawLine(screen, 350, 100, float32(x), float32(y), 10, color.RGBA{0xff, 0x00, 0x00, 0xff})

	var vs []ebiten.Vertex
	var is []uint16
	op := &vector.StrokeOptions{}
	op.Width = 5
	op.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, op)

	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0x33 / float32(0xff)
		vs[i].ColorG = 0xcc / float32(0xff)
		vs[i].ColorB = 0x66 / float32(0xff)
		vs[i].ColorA = 1
	}

	op2 := &ebiten.DrawTrianglesOptions{}
	op2.AntiAlias = true
	op2.FillRule = ebiten.FillRuleFillAll
	screen.DrawTriangles(vs, is, whiteImage, op2)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("arc to")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
