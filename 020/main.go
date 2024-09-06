package main

import (
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/KEINOS/go-noise"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	step   = 20
	smooth = float64(3)
)

type Game struct {
	present   [][]Dot // [y][x]Dot
	reference [][]Dot // [y][x]Dot
	pos       [][]Pos
	time      int
	shift     float64
	seed      int64
	grad      GradientTable
}
type Dot struct {
	C, R float64
}
type Pos struct {
	X, Y float32
}

func (g *Game) Update() error {
	for y, ly := 0, len(g.reference); y < ly; y++ {
		if len(g.pos) <= y {
			g.pos = append(g.pos, make([]Pos, len(g.reference[y])))
		}
		for x, lx := 0, len(g.reference[y]); x < lx; x++ {
			if len(g.present) <= y {
				continue
			}
			if len(g.present[y]) <= x {
				continue
			}
			rgba1 := g.reference[y][x]
			rgba2 := g.present[y][x]
			g.present[y][x].R += (rgba1.R - rgba2.R) * 0.1
			g.present[y][x].C += (rgba1.C - rgba2.C) * 0.1

			r := g.present[y][x].R
			cos := float32(math.Cos(float64(r)*math.Pi)) * 40
			sin := float32(math.Sin(float64(r)*math.Pi))*40 - 40
			g.pos[y][x] = Pos{float32(x*step) + cos, float32(y*step) + sin}
		}
	}

	if g.time%60 == 0 {
		g.reference = genDot(screenWidth/step+1, screenHeight/step+1, g.seed, g.shift)
		g.shift += .15
	}
	g.time++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// screen.Fill(color.RGBA{0xC7, 0xCC, 0xD9, 0xff}) // #C7CCD9
	screen.Fill(color.RGBA{0x19, 0x44, 0x6B, 0xff}) // #19446B
	for y, ly := 0, len(g.reference); y < ly; y++ {
		for x, lx := 0, len(g.reference[y]); x < lx; x++ {
			if len(g.present) <= y {
				continue
			}
			if len(g.present[y]) <= x {
				continue
			}
			var col color.NRGBA
			col2 := g.grad.GetInterpolatedColorFor(g.present[y][x].C)
			col = color.NRGBA{
				R: uint8(col2.R * float64(0xff)),
				G: uint8(col2.G * float64(0xff)),
				B: uint8(col2.B * float64(0xff)),
				A: uint8(float64(0xff)),
			}
			r := g.present[y][x].R
			if x+1 >= len(g.pos[y]) {
				continue
			}
			if y+1 >= len(g.pos) {
				continue
			}
			current := g.pos[y][x]
			right := g.pos[y][x+1]
			bottom := g.pos[y+1][x]
			bottomRight := g.pos[y+1][x+1]

			drawShape(screen, current, right, bottom, bottomRight, float32(r), col)
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%.2f", ebiten.ActualFPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten perlin noise")
	g := &Game{}
	g.reference = genDot(screenWidth/step+1, screenHeight/step+1, g.seed, g.shift)
	g.present = genDot(screenWidth/step+1, screenHeight/step+1, g.seed, g.shift)
	g.grad = NewGradientTable()
	ebiten.RunGame(g)
}

// noise

func genDot(width, height int, seed int64, shift float64) [][]Dot {
	perlin := make([][]Dot, height, height)
	colorN, err := noise.New(noise.Perlin, seed)
	radiusN, err := noise.New(noise.Perlin, seed)
	if err != nil {
		log.Fatal(err)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 {
				perlin[y] = make([]Dot, width, width)
			}
			// color
			cv := colorN.Eval64(float64(x)/smooth/5, float64(y)/smooth/5, shift) // v is between -1.0 and 1.0 of float64
			col := (cv + 1.0) / 2.0
			// radius
			rv := radiusN.Eval64(float64(x)/smooth, float64(y)/smooth, shift) // v is between -1.0 and 1.0 of float64
			rad := (rv + 1.0) / 2.0

			perlin[y][x] = Dot{C: col, R: rad}
		}
	}
	return perlin
}

// gradient

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func NewGradientTable() GradientTable {
	keypoints := GradientTable{
		// {MustParseHex("#A68F72"), 0},
		// {MustParseHex("#3F7373"), 0.3},
		// {MustParseHex("#732B1A"), 0.6},
		// {MustParseHex("#BF754B"), 1.0},
		{MustParseHex("#DE183C"), 0},
		{MustParseHex("#F2B541"), 0.2},
		{MustParseHex("#0C79BB"), 0.4},
		{MustParseHex("#2DACB2"), 0.6},
		{MustParseHex("#E46424"), 0.8},
		{MustParseHex("#ECACBE"), 1.0},
	}

	return keypoints
}
func (gt GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(gt)-1; i++ {
		c1 := gt[i]
		c2 := gt[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}
	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return gt[len(gt)-1].Col
}
func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}

// graphic utils

var (
	whiteSubImage = ebiten.NewImage(3, 3)
)

func init() {
	whiteSubImage.Fill(color.White)
}
func drawFill(screen *ebiten.Image, path vector.Path, c color.NRGBA) {
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
	// op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func drawCircle(screen *ebiten.Image, x, y, radius float32, c color.NRGBA) {
	path := vector.Path{}
	path.Arc(x, y, radius, 0, 2*math.Pi, vector.Clockwise)
	drawFill(screen, path, c)
}
func drawLine(screen *ebiten.Image, x1, y1, x2, y2, width float32, c color.NRGBA) {
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
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func drawShape(screen *ebiten.Image, p0, p1, p2, p3 Pos, r float32, c color.NRGBA) {
	drawLine(screen, p0.X, p0.Y, p1.X, p1.Y, 1, c)
	drawLine(screen, p0.X, p0.Y, p2.X, p2.Y, 1, c)
	drawLine(screen, p1.X, p1.Y, p3.X, p3.Y, 1, c)
	drawLine(screen, p2.X, p2.Y, p3.X, p3.Y, 1, c)
	drawLine(screen, p0.X, p0.Y, p3.X, p3.Y, 1, c)
	drawLine(screen, p1.X, p1.Y, p2.X, p2.Y, 1, c)
}
