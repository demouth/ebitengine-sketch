package minigui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type GUI struct {
	components []Component
	whiteImage *ebiten.Image

	x               float32
	y               float32
	width           float32
	componentHeight float32
}

type Component interface {
	IsComponent()
}
type sliderFloat64 struct {
	value    float64
	min      float64
	max      float64
	hovered  bool
	callback func(v float64)
}

func (s *sliderFloat64) IsComponent() {}

func NewGUI() *GUI {
	whiteImage := ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)
	gui := &GUI{
		whiteImage:      whiteImage,
		width:           200,
		componentHeight: 10,
		x:               400,
		y:               10,
	}
	return gui
}

func (g *GUI) Update() {
	cx, cy := ebiten.CursorPosition()
	x := g.x
	y := g.y
	for _, c := range g.components {
		switch c.(type) {
		case *sliderFloat64:
			s := c.(*sliderFloat64)
			s.Update(float32(cx)-x, float32(cy)-y, g.width, g.componentHeight)
		}
		y += g.componentHeight
	}
}
func (g *GUI) Draw(image *ebiten.Image) {
	x := g.x
	y := g.y
	for _, c := range g.components {
		drawRect(image, g.whiteImage, x, y, g.width, g.componentHeight, color.NRGBA{0x31, 0x31, 0x31, 0xff})
		switch c.(type) {
		case *sliderFloat64:
			s := c.(*sliderFloat64)
			s.Draw(image, g.whiteImage, x, y, g.width, g.componentHeight)
		}
		y += g.componentHeight
	}
}

func (g *GUI) AddSliderFloat64(label string, value float64, min, max float64, callback func(v float64)) {
	s := &sliderFloat64{
		value:    value,
		min:      min,
		max:      max,
		callback: callback,
	}
	g.components = append(g.components, s)
}

func (s *sliderFloat64) Update(x, y, width, height float32) {
	clicked := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if x > width/2 && x < width && y > 0 && y < height {
		s.hovered = true
	} else {
		s.hovered = false
	}

	if s.hovered && clicked {
		ratio := float64((x - width/2) / width * 2)
		valueRange := s.max - s.min
		s.value = s.min + valueRange*ratio
		s.callback(s.value)
	}
}
func (s *sliderFloat64) Draw(image *ebiten.Image, whiteImage *ebiten.Image, x, y, width, height float32) {
	padding := float32(1)
	x = x + width/2 + padding
	y = y + padding
	w := width/2 - padding*2
	h := height - padding*2
	c := color.NRGBA{0x66, 0x66, 0x66, 0xff}
	if s.hovered {
		c = color.NRGBA{0x79, 0x79, 0x79, 0xff}
	}
	drawRect(image, whiteImage, x, y, w, h, c)

	valueRange := float32(s.max - s.min)
	drawRange := w
	ratio := float32(s.value-s.min) / valueRange
	drawWidth := drawRange * ratio

	drawLine(image, whiteImage, x+drawWidth, y, x+drawWidth, y+h, 1, color.NRGBA{0x2c, 0xc9, 0xff, 0xff})
}

func drawLine(screen *ebiten.Image, whiteImage *ebiten.Image, x1, y1, x2, y2, width float32, c color.NRGBA) {
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
	op.FillRule = ebiten.FillRuleFillAll
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
	screen.DrawTriangles(vs, is, whiteImage, op)
}
