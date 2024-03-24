package main

import (
	"image/color"
	"log"
	"runtime"

	"github.com/hajimehoshi/bitmapfont/v3"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var fontFace = bitmapfont.FaceEA

const (
	screenWidth  = 640
	screenHeight = 480
)

type TextField struct {
	X    int
	Y    int
	Text string

	ch    chan textinput.State
	end   func()
	state textinput.State
}

func NewTextField() *TextField {
	return &TextField{}
}
func (t *TextField) Update() {
	if t.X == 0 && t.Y == 0 {
		return
	}
	var processed bool

	// Text inputting can happen multiple times in one tick (1/60[s] by default).
	// Handle all of them.
	for {
		if t.ch == nil {
			x, y := t.X, t.Y
			t.ch, t.end = textinput.Start(x, y)
			// Start returns nil for non-supported envrionments.
			if t.ch == nil {
				return
			}
		}

	readchar:
		for {
			select {
			case state, ok := <-t.ch:
				processed = true
				if !ok {
					t.ch = nil
					t.end = nil
					t.state = textinput.State{}
					break readchar
				}
				if state.Committed {
					t.Text = state.Text
					t.state = textinput.State{}
					continue
				}
				t.state = state
			default:
				break readchar
			}
		}

		if t.ch == nil {
			continue
		}

		break
	}

	if processed {
		return
	}

}

func (t *TextField) Draw(screen *ebiten.Image) {
	shownText := t.Text
	c := color.Black
	if t.state.Text != "" {
		shownText += t.state.Text
		c = color.Gray16{0x9999}
	}
	x := t.X
	y := t.Y

	vector.DrawFilledCircle(screen, float32(x), float32(y), 6, color.RGBA{0xff, 0, 0, 0xff}, true)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 2)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.DrawWithOptions(screen, shownText, fontFace, op)
}

type Game struct {
	textField *TextField
}

func (g *Game) Update() error {
	if g.textField == nil {
		g.textField = NewTextField()
		g.textField.Text = ""
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.textField.X = x
		g.textField.Y = y
		g.textField.Text = ""
	}

	g.textField.Update()

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
