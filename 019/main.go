package main

import (
	_ "embed"
	_ "image/png"
	"log"
	"math"

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

	cx, cy := ebiten.CursorPosition()

	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Uniforms = map[string]any{
		"Time":   float32(g.time) / 60,
		"Cursor": []float32{float32(cx), float32(cy)},
	}

	vertices := []ebiten.Vertex{
		{
			DstX: 0, DstY: -100,
			ColorR: 1, ColorG: 0, ColorB: 0, ColorA: 1,
			Custom0: 0,
		},
		{
			DstX: -150, DstY: 100,
			ColorR: 0, ColorG: 1, ColorB: 0, ColorA: 1,
			Custom0: 0,
		},
		{
			DstX: 150, DstY: 100,
			ColorR: 0, ColorG: 0, ColorB: 1, ColorA: 1,
			Custom0: 0,
		},
	}
	indices := []uint16{0, 1, 2}

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	cx, cy = cx-screenWidth/2, cy-screenHeight/2
	thetaY := float64(cx) / 1000
	thetaX := float64(cy) / 1000
	for i := 0; i < len(vertices); i++ {
		x, y, z := vertices[i].DstX, vertices[i].DstY, vertices[i].Custom0

		// rotate Y
		sin64, cos64 := math.Sincos(thetaY)
		sin, cos := float32(sin64), float32(cos64)
		x = x*cos - z*sin
		z = z*cos + x*sin

		// rotate X
		sin64, cos64 = math.Sincos(thetaX)
		sin, cos = float32(sin64), float32(cos64)
		y = y*cos - z*sin
		z = z*cos + y*sin

		// perspective
		fl := float32(550)
		scale := fl / (fl + z)
		x *= scale
		y *= scale
		vertices[i].DstX = x + float32(w/2)
		vertices[i].DstY = y + float32(h/2)
		vertices[i].Custom0 = z
	}

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
