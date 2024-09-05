package main

import (
	_ "embed"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed shader.kage
	shader_kage []byte
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	shader *ebiten.Shader
	idx    int
	time   int
}

func (g *Game) Update() error {
	g.time++
	if g.shader == nil {
		s, err := ebiten.NewShader([]byte(shader_kage))
		if err != nil {
			return err
		}
		g.shader = s
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	s := g.shader

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	cx, cy := ebiten.CursorPosition()

	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Uniforms = map[string]any{
		"Time":   float32(g.time) / 60,
		"Cursor": []float32{float32(cx), float32(cy)},
	}

	vertices := []ebiten.Vertex{
		{
			DstX: float32(w / 2), DstY: float32(h/2) - 100,
			ColorR: 1, ColorG: 0, ColorB: 0, ColorA: 1,
		},
		{
			DstX: float32(w/2) - 150, DstY: float32(h/2) + 100,
			ColorR: 0, ColorG: 1, ColorB: 0, ColorA: 1,
		},
		{
			DstX: float32(w/2) + 150, DstY: float32(h/2) + 100,
			ColorR: 0, ColorG: 0, ColorB: 1, ColorA: 1,
		},
	}
	indices := []uint16{0, 1, 2}
	screen.DrawTrianglesShader(vertices, indices, s, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Shader (Kage)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
