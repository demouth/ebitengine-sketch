package minigui

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	textFaceSource *text.GoTextFaceSource
)

type HorizontalAlign int

const (
	HorizontalAlignLeft HorizontalAlign = iota
	HorizontalAlignRight
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	textFaceSource = s

}

type GUI struct {
	components []Component
	whiteImage *ebiten.Image

	X               float32
	Y               float32
	Width           float32
	ComponentHeight float32
	Scale           float32

	HorizontalAlign HorizontalAlign
}

type Component interface {
	Label() string
	Update(x, y, width, height, scale float32)
	Draw(image *ebiten.Image, whiteImage *ebiten.Image, top, left, width, height, scale float32)
}

func NewGUI() *GUI {

	whiteImage := ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)
	gui := &GUI{
		whiteImage:      whiteImage,
		Width:           200,
		ComponentHeight: 24,
		X:               0,
		Y:               0,
		Scale:           1,
		HorizontalAlign: HorizontalAlignLeft,
	}
	return gui
}
func (g *GUI) Update() {
	cx, cy := ebiten.CursorPosition()
	x := g.X
	y := g.Y
	if g.HorizontalAlign == HorizontalAlignRight {
		x = g.X - g.Width*g.Scale
	}
	for _, c := range g.components {
		switch c.(type) {
		case *sliderFloat64:
			s := c.(*sliderFloat64)
			s.Update(float32(cx)-x, float32(cy)-y, g.Width, g.ComponentHeight, g.Scale)
		case *sliderFloat32:
			s := c.(*sliderFloat32)
			s.Update(float32(cx)-x, float32(cy)-y, g.Width, g.ComponentHeight, g.Scale)
		case *sliderInt:
			s := c.(*sliderInt)
			s.Update(float32(cx)-x, float32(cy)-y, g.Width, g.ComponentHeight, g.Scale)
		case *button:
			s := c.(*button)
			s.Update(float32(cx)-x, float32(cy)-y, g.Width, g.ComponentHeight, g.Scale)
		}
		y += g.ComponentHeight * g.Scale
	}
}
func (g *GUI) Draw(image *ebiten.Image) {
	var x, y float32

	// Drawing shapes
	x = g.X
	y = g.Y
	if g.HorizontalAlign == HorizontalAlignRight {
		x = g.X - g.Width*g.Scale
	}
	drawRect(
		image,
		g.whiteImage,
		x, y,
		g.Width*g.Scale,
		g.ComponentHeight*g.Scale*float32(len(g.components)),
		color.NRGBA{0x31, 0x31, 0x31, 0xff},
	)
	for _, c := range g.components {
		// Draw label
		textPadding := 5.0 * g.Scale
		fontSize := g.ComponentHeight*g.Scale - textPadding*2
		drawText(image, c.Label(), x+textPadding, y+textPadding, fontSize, color.NRGBA{R: 0xeb, G: 0xeb, B: 0xeb, A: 0xff})

		// Draw component
		c.Draw(image, g.whiteImage, x, y, g.Width, g.ComponentHeight, g.Scale)
		y += g.ComponentHeight * g.Scale
	}
}

func drawLine(screen *ebiten.Image, whiteImage *ebiten.Image, x1, y1, x2, y2, width float32, c color.NRGBA) {
	path := vector.Path{}
	path.MoveTo(x1, y1)
	path.LineTo(x2, y2)
	path.Close()
	sop := &vector.StrokeOptions{}
	sop.Width = width
	sop.LineJoin = vector.LineJoinMiter
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
	op.FillRule = ebiten.FillRuleFillAll
	// AntiAlias is not used to reduce the number of draw-triangles issued.
	// op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}
func drawRect(screen *ebiten.Image, whiteImage *ebiten.Image, x, y, width, height float32, c color.NRGBA) {
	path := vector.Path{}
	path.MoveTo(x, y)
	path.LineTo(x, y+height)
	path.LineTo(x+width, y+height)
	path.LineTo(x+width, y)
	path.Close()
	drawFill(screen, whiteImage, path, c)
}
func drawFill(screen *ebiten.Image, whiteImage *ebiten.Image, path vector.Path, c color.NRGBA) {
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
	op.FillRule = ebiten.FillRuleFillAll
	// AntiAlias is not used to reduce the number of draw-triangles issued.
	// op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}

func drawText(image *ebiten.Image, str string, x, y, fontSize float32, c color.NRGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	op.LineSpacing = 0
	op.PrimaryAlign = text.AlignStart
	fontFace := &text.GoTextFace{
		Source: textFaceSource,
		Size:   float64(fontSize),
	}
	text.Draw(
		image,
		str,
		fontFace,
		op,
	)
}
