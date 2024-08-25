package minigui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type button struct {
	label    string
	value    bool
	hovered  bool
	callback func(v bool)
}

func (g *GUI) AddButton(label string, value bool, callback func(v bool)) {
	b := &button{
		label:    label,
		value:    value,
		callback: callback,
	}
	g.components = append(g.components, b)
}

func (s *button) Label() string {
	return s.label
}
func (s *button) Update(x, y, width, height, scale float32) {
	paddingLeft := 2.0 * scale
	paddingTop := 4.0 * scale
	buttonSize := height*scale - paddingTop*2
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if x >= width/2*scale+paddingLeft && x <= width/2*scale+paddingLeft+buttonSize && y > paddingTop && y <= paddingTop+buttonSize {
		s.hovered = true
	} else {
		s.hovered = false
	}

	if s.hovered && clicked {
		s.value = !s.value
		s.callback(s.value)
	}
}
func (s *button) Draw(image *ebiten.Image, whiteImage *ebiten.Image, left, top, width, height, scale float32) {
	paddingLeft := 2.0 * scale
	paddingTop := 3.0 * scale
	buttonSize := height*scale - paddingTop*2

	x := left + width/2*scale + paddingLeft
	y := top + paddingTop
	w := buttonSize
	h := buttonSize

	c := color.NRGBA{0x66, 0x66, 0x66, 0xff}
	if s.hovered {
		c = color.NRGBA{0x79, 0x79, 0x79, 0xff}
	}
	drawRect(image, whiteImage, x, y, w, h, c)

	if s.value {
		paddingIcon := 4.0 * scale
		drawRect(image, whiteImage, x+paddingIcon, y+paddingIcon, w-paddingIcon*2, h-paddingIcon*2, color.NRGBA{0x2c, 0xc9, 0xff, 0xff})
	}
}
