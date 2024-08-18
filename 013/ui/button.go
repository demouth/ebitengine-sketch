package ui

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	arcadeFaceSource *text.GoTextFaceSource

	whiteImage = ebiten.NewImage(3, 3)

	cursor   = ebiten.CursorShapeDefault
	touchIDs []ebiten.TouchID
	strokes  map[StrokeSource]struct{}
)

type StrokeSource interface {
	Position() (int, int)
	IsJustReleased() bool
}
type Stroke struct {
	source StrokeSource
}
type TouchStrokeSource struct {
	ID ebiten.TouchID
}

func (t *TouchStrokeSource) Position() (int, int) {
	return ebiten.TouchPosition(t.ID)
}
func (t *TouchStrokeSource) IsJustReleased() bool {
	return inpututil.IsTouchJustReleased(t.ID)
}

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	arcadeFaceSource = s

	whiteImage.Fill(color.White)

	strokes = make(map[StrokeSource]struct{})
}

type Button struct {
	Width           float32
	Height          float32
	X               float32
	Y               float32
	Text            string
	FontSize        float64
	OnMouseDown     func()
	OnMouseDownHold func()

	hovered bool
}

type Component interface {
	Update()
	Draw(*ebiten.Image)
}

type Components []Component

func (c Components) Update() {
	for _, component := range c {
		component.Update()
	}
}

func (c Components) Draw(screen *ebiten.Image) {
	for _, component := range c {
		component.Draw(screen)
	}
}

func (b *Button) Update() {
	x, y := ebiten.CursorPosition()

	justTouched := false
	touchIDs = inpututil.AppendJustPressedTouchIDs(touchIDs[:0])
	for _, id := range touchIDs {
		s := &TouchStrokeSource{id}
		strokes[s] = struct{}{}
		justTouched = true
	}
	for s := range strokes {
		if s.IsJustReleased() {
			delete(strokes, s)
			continue
		}
		x, y = s.Position()
	}

	var newCursor ebiten.CursorShapeType
	var newHovered bool
	if x >= int(b.X) && x <= int(b.X+b.Width) && y >= int(b.Y) && y <= int(b.Y+b.Height) {
		newHovered = true
		newCursor = ebiten.CursorShapePointer
	} else {
		newHovered = false
		newCursor = ebiten.CursorShapeDefault
	}

	if b.hovered != newHovered {
		ebiten.SetCursorShape(newCursor)
		cursor = newCursor
		b.hovered = newHovered
	}

	if (inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || justTouched) && b.hovered {
		if b.OnMouseDown != nil {
			b.OnMouseDown()
		}
	}

	if (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(strokes) > 0) && b.hovered {
		if b.OnMouseDownHold != nil {
			b.OnMouseDownHold()
		}
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	var paddingTop float64 = (float64(b.Height) - b.FontSize) / 2
	outline := color.NRGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xff}
	background := color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff}
	if b.hovered {
		background = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	}
	b.drawBackground(screen, outline, background)

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(b.X+b.Width/2), float64(b.Y)+paddingTop)
	op.ColorScale.ScaleWithColor(color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff})
	op.LineSpacing = 0
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, b.Text, &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   b.FontSize,
	}, op)
}

func (b *Button) drawBackground(screen *ebiten.Image, outline color.NRGBA, background color.NRGBA) {
	path := b.makePath()
	var vs []ebiten.Vertex
	var is []uint16
	vs, is = path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(background.R) / float32(0xff)
		vs[i].ColorG = float32(background.G) / float32(0xff)
		vs[i].ColorB = float32(background.B) / float32(0xff)
		vs[i].ColorA = float32(background.A) / float32(0xff)
	}
	screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.FillRuleFillAll,
	})

	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{
		Width:    3,
		LineJoin: vector.LineJoinRound,
	})
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(outline.R) / float32(0xff)
		vs[i].ColorG = float32(outline.G) / float32(0xff)
		vs[i].ColorB = float32(outline.B) / float32(0xff)
		vs[i].ColorA = float32(outline.A) / float32(0xff)
	}
	screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.FillRuleFillAll,
	})
}

func (b *Button) makePath() vector.Path {
	w := b.Width
	var path vector.Path
	path.MoveTo(b.X, b.Y)
	path.LineTo(b.X+w, b.Y)
	path.LineTo(b.X+w, b.Y+b.Height)
	path.LineTo(b.X, b.Y+b.Height)
	path.Close()
	return path
}
