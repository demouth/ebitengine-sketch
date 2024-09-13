package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/demouth/ebitengine-sketch/024/drawer"
	"github.com/ebitengine/microui"
	"github.com/hajimehoshi/ebiten/v2"
	resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/shader"
)

var (
	//go:embed shader.kage
	shader_kage   []byte
	gopherBgImage *ebiten.Image
)

const (
	screenWidth  = 480
	screenHeight = 480
)

type Game struct {
	ctx          *microui.Context
	shader       *ebiten.Shader
	numTriangles float64
}

func (g *Game) Update() error {
	g.ctx.Update(func() {
		g.ctx.Window("triangles", image.Rect(250, 40, 400, 120), func(res microui.Res) {
			if g.ctx.HeaderEx("Num Triangles", microui.OptExpanded) != 0 {
				g.ctx.SliderEx(&g.numTriangles, 1, 10, 1, "%.2f", microui.OptAlignCenter)
			}
		})
	})
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
	screen.Fill(color.RGBA{0x66, 0x66, 0x66, 0xff})
	s := g.shader

	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Uniforms = map[string]any{}
	op.Images[0] = gopherBgImage

	vertices, indices := makeVerticesAndindices(int(g.numTriangles))
	screen.DrawTrianglesShader(vertices, indices, s, op)

	for i := 0; i < len(indices)-1; i++ {
		drawer.DrawLine(
			screen,
			float32(vertices[indices[i]].DstX),
			float32(vertices[indices[i]].DstY),
			float32(vertices[indices[(i+1)]].DstX),
			float32(vertices[indices[(i+1)]].DstY),
			2,
			color.RGBA{0xff, 0xff, 0xff, 0xff},
		)
	}
	g.ctx.Draw(screen)
}

func makeVerticesAndindices(l int) ([]ebiten.Vertex, []uint16) {
	var vertices []ebiten.Vertex
	var indices []uint16
	const length = float32(100)
	for i := 0; i < l; i++ {
		offset := float32(math.Floor(float64(i)/2)) * length
		if i%2 == 0 {
			vertices = append(vertices, []ebiten.Vertex{
				{
					DstX: 0, DstY: 0 + offset,
					SrcX: 0, SrcY: 0,
				},
				{
					DstX: length, DstY: 0 + offset,
					SrcX: 640, SrcY: 0,
				},
				{
					DstX: 0, DstY: length + offset,
					SrcX: 0, SrcY: 480,
				},
			}...)
			indices = append(indices, uint16(i)*2, uint16(i)*2+1, uint16(i)*2+2)
		} else {
			vertices = append(vertices, []ebiten.Vertex{
				{
					DstX: length, DstY: length + offset,
					SrcX: 640, SrcY: 480,
				},
			}...)
			num := math.Floor(float64(i) / 2)
			indices = append(indices, uint16(num*4)+1, uint16(num*4)+3, uint16(num*4)+2)
		}
	}
	return vertices, indices
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	img, _, err := image.Decode(bytes.NewReader(resources.GopherBg_png))
	if err != nil {
		log.Fatal(err)
	}
	gopherBgImage = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Shader (Kage)")
	g := &Game{
		ctx:          microui.NewContext(),
		numTriangles: 3,
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
