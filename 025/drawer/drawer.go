package drawer

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteSubImage *ebiten.Image
)

func init() {
	whiteSubImage = ebiten.NewImage(3, 3)
	whiteSubImage.Fill(color.White)
}

func DrawLine(screen *ebiten.Image, x1, y1, x2, y2, width float32, c color.RGBA) {
	path := vector.Path{}
	path.MoveTo(x1, y1)
	path.LineTo(x2, y2)
	path.Close()
	sop := &vector.StrokeOptions{}
	sop.Width = width
	sop.LineJoin = vector.LineJoinRound
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(c.R) / float32(0xff)
		vs[i].ColorG = float32(c.G) / float32(0xff)
		vs[i].ColorB = float32(c.B) / float32(0xff)
		vs[i].ColorA = float32(c.A) / float32(0xff)
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillAll
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func DrawFill(screen *ebiten.Image, path vector.Path, c color.NRGBA) {
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(c.R) / float32(0xff)
		vs[i].ColorG = float32(c.G) / float32(0xff)
		vs[i].ColorB = float32(c.B) / float32(0xff)
		vs[i].ColorA = float32(c.A) / float32(0xff)
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillAll
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}

func DrawCircle(screen *ebiten.Image, x, y, radius float32, c color.NRGBA) {
	path := vector.Path{}
	path.Arc(x, y, radius, 0, 2*math.Pi, vector.Clockwise)
	DrawFill(screen, path, c)
}
