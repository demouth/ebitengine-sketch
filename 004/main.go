package main

import (
	"fmt"
	"image/color"
	"log"
	"runtime"

	"github.com/hajimehoshi/bitmapfont/v3"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var fontFace = text.NewGoXFace(bitmapfont.FaceEA)

const (
	screenWidth  = 640
	screenHeight = 480
)

type TextField struct {
	X int
	Y int

	field textinput.Field
}

func NewTextField() *TextField {
	return &TextField{}
}
func (t *TextField) Update() error {
	if t.X == 0 && t.Y == 0 {
		return nil
	}

	x, y := t.X, t.Y

	println("x:", x, "y:", y)

	handled, err := t.field.HandleInput(x, y)
	fmt.Println("handled:", handled, "err:", err)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	return nil
}
func (t *TextField) Reset() {
	t.field.SetTextAndSelection("", 0, 0)
}
func (t *TextField) Focus() {
	t.field.Focus()
}

func (t *TextField) Draw(screen *ebiten.Image) {
	c := color.Black
	x := t.X
	y := t.Y
	h := int(fontFace.Metrics().HLineGap + fontFace.Metrics().HAscent + fontFace.Metrics().HDescent)
	var scale float64 = 2

	vector.DrawFilledCircle(screen, float32(x), float32(y), 6, color.RGBA{0xff, 0, 0, 0xff}, true)

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.GeoM.Translate(0, -float64(h)*scale)
	op.ColorScale.ScaleWithColor(c)
	op.LineSpacing = fontFace.Metrics().HLineGap + fontFace.Metrics().HAscent + fontFace.Metrics().HDescent
	text.Draw(screen, t.field.TextForRendering(), fontFace, op)
}

type Game struct {
	textField *TextField
	touchIDs  []ebiten.TouchID
}

func (g *Game) Update() error {
	if g.textField == nil {
		g.textField = NewTextField()
		g.textField.Focus()
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.textField.X = x
		g.textField.Y = y
		g.textField.Reset()
		g.textField.Focus()
	}

	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	for _, id := range g.touchIDs {
		x, y := ebiten.TouchPosition(id)
		g.textField.X = x
		g.textField.Y = y
		g.textField.Reset()
		g.textField.Focus()
	}

	if err := g.textField.Update(); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xff, 0xcc, 0xff})
	g.textField.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	if runtime.GOOS != "darwin" && runtime.GOOS != "js" {
		log.Printf("github.com/hajimehoshi/ebiten/v2/exp/textinput is not supported in this environment (GOOS=%s) yet", runtime.GOOS)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("003 exp/textinput")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
