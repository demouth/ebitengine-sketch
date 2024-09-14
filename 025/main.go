package main

import (
	"bytes"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	textFaceSource *text.GoTextFaceSource
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	textFaceSource = s
}

type Game struct {
	vertices   []ebiten.Vertex
	indices    []uint16
	whiteImage *ebiten.Image
	time       int
}

func (g *Game) Update() error {
	g.time++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0x66, 0x66, 0x66, 0xff})

	path := &vector.Path{}
	{
		face := &text.GoTextFace{Source: textFaceSource, Size: 200}
		op := &text.LayoutOptions{}
		op.LineSpacing = 200
		text.AppendVectorPath(path, "寿司を\n食べた\nい", face, op)
	}

	g.vertices, g.indices = path.AppendVerticesAndIndicesForFilling(g.vertices[:0], g.indices[:0])

	for i := range g.vertices {
		g.vertices[i].DstX = g.vertices[i].DstX + float32(math.Sin(float64(float32(g.time)+g.vertices[i].DstX)/10))*4 + rand.Float32()*3
		g.vertices[i].DstY = g.vertices[i].DstY + float32(math.Cos(float64(float32(g.time)+g.vertices[i].DstY)/14))*4 + rand.Float32()*3 - 30
		g.vertices[i].SrcX = 1
		g.vertices[i].SrcY = 1
		g.vertices[i].ColorR = 0xff
		g.vertices[i].ColorG = 0xff
		g.vertices[i].ColorB = 0xff
		g.vertices[i].ColorA = 1
	}

	{
		op := &ebiten.DrawTrianglesOptions{}
		op.FillRule = ebiten.FillRuleNonZero
		op.AntiAlias = true
		screen.DrawTriangles(g.vertices, g.indices, g.whiteImage, op)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{
		whiteImage: ebiten.NewImage(3, 3),
	}
	g.whiteImage.Fill(color.White)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine - outline text")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
