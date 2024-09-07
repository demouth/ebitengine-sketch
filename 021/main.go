package main

import (
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	pmath "changkun.de/x/polyred/math"
	"github.com/KEINOS/go-noise"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 800
	maxLife      = 80
)

var (
	step   = 40
	smooth = float64(10)
)

type Perticle struct {
	X, Y       float32
	OldX, OldY float32
	SX, SY     float32
	life       int
	color      color.RGBA
	dead       bool
}

func (p *Perticle) Move(forceMap [][]Noise) {
	if p.dead {
		p.X = rand.Float32() * screenWidth
		p.Y = rand.Float32() * screenHeight
		p.OldX = p.X
		p.OldY = p.Y
		p.SX = 0
		p.SY = 0
		p.dead = false
		p.life = maxLife
		return
	}
	p.life--

	force := forceMap[int(p.Y)/step][int(p.X)/step]
	p.OldX = p.X
	p.OldY = p.Y
	p.SX += float32(force.R) * 0.6
	p.SY += float32(force.G) * 0.6
	p.SX *= 0.94
	p.SY *= 0.94
	// p.SX += (rand.Float32() - 0.5) * 0.01
	// p.SY += (rand.Float32() - 0.5) * 0.01
	p.X += p.SX
	p.Y += p.SY

	ix := int(p.Y) / step
	if len(forceMap) <= ix {
		p.dead = true
		return
	}
	if ix < 0 {
		p.dead = true
		return
	}
	iy := int(p.X) / step
	if len(forceMap[ix]) <= iy {
		p.dead = true
		return
	}
	if iy < 0 {
		p.dead = true
		return
	}
	if math.Abs(float64(p.SX)) < 0.01 && math.Abs(float64(p.SY)) < 0.01 {
		p.dead = true
		return
	}
	if p.life < 0 {
		p.dead = true
	}
}

type Game struct {
	reference [][]Noise
	time      int
	shift     float64
	seed      int64
	// grad      GradientTable
	perticles []Perticle
	canvas    *ebiten.Image
}
type Noise struct {
	R, G float64
}

func (g *Game) Update() error {
	g.time++
	for i := 0; i < len(g.perticles); i++ {
		g.perticles[i].Move(g.reference)
	}
	if g.time%60 == 0 {
		g.shift += .02
		g.reference = genDot(screenWidth/step+1, screenHeight/step+1, g.seed, g.shift)
	}
	// draw to canvas
	drawRect(
		g.canvas,
		0, 0, float32(screenWidth), float32(screenHeight),
		color.RGBA{0x00, 0x00, 0x00, 0x06})
	for i := 0; i < len(g.perticles); i++ {
		p := g.perticles[i]
		if p.OldX <= 0 || p.OldY <= 0 {
			continue
		}
		if p.dead {
			continue
		}

		col := p.color
		var lw float32 = 2.0
		var elderly float32 = 0.5
		if float32(p.life) < float32(maxLife)*elderly {
			r := float32(p.life) / float32(maxLife) / elderly
			col = pmath.LerpC(p.color, color.RGBA{0x00, 0x00, 0x00, 0x00}, float64(1-r))
			lw = lw * r
		}
		speed := math.Abs(float64(p.SX)) + math.Abs(float64(p.SY))
		lw = lw * float32(speed)
		lw = float32(math.Min(5, float64(lw)))
		lw = float32(math.Max(1, float64(lw)))

		drawLine(g.canvas, p.X, p.Y, p.OldX, p.OldY, lw, col)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.canvas, nil)
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf(
			"FPS:%.2f\nShift:%.3f",
			ebiten.ActualFPS(),
			g.shift,
		),
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("perlin noise")
	colors := NewColors(2)
	g := &Game{}
	g.reference = genDot(screenWidth/step+1, screenHeight/step+1, g.seed, g.shift)
	for i := 0; i < 3000; i++ {
		p := Perticle{
			X:     rand.Float32() * screenWidth,
			Y:     rand.Float32() * screenHeight,
			color: colors.Random(),
			life:  maxLife,
		}
		g.perticles = append(g.perticles, p)
	}
	g.canvas = ebiten.NewImage(screenWidth, screenHeight)
	g.canvas.Fill(color.RGBA{0x00, 0x00, 0x00, 0xff}) // #000000
	ebiten.RunGame(g)
}

// noise

func genDot(width, height int, seed int64, shift float64) [][]Noise {
	perlin := make([][]Noise, height, height)
	nr, err := noise.New(noise.Perlin, seed)
	ng, err := noise.New(noise.Perlin, seed+1)
	if err != nil {
		log.Fatal(err)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x == 0 {
				perlin[y] = make([]Noise, width, width)
			}
			r := nr.Eval64(float64(x)/smooth, float64(y)/smooth, shift) // v is between -1.0 and 1.0 of float64

			g := ng.Eval64(float64(x)/smooth, float64(y)/smooth, shift) // v is between -1.0 and 1.0 of float64
			perlin[y][x] = Noise{R: r, G: g}
		}
	}
	return perlin
}

// graphic utils

var (
	whiteSubImage = ebiten.NewImage(3, 3)
)

func init() {
	whiteSubImage.Fill(color.White)
}

func drawRect(screen *ebiten.Image, x1, y1, x2, y2 float32, c color.RGBA) {
	path := vector.Path{}
	path.MoveTo(x1, y1)
	path.LineTo(x2, y1)
	path.LineTo(x2, y2)
	path.LineTo(x1, y2)
	path.Close()
	drawFill(screen, path, c)
}
func drawFill(screen *ebiten.Image, path vector.Path, c color.RGBA) {
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
func drawCircle(screen *ebiten.Image, x, y, radius float32, c color.RGBA) {
	path := vector.Path{}
	path.Arc(x, y, radius, 0, 2*math.Pi, vector.Clockwise)
	drawFill(screen, path, c)
}
func drawLine(screen *ebiten.Image, x1, y1, x2, y2, width float32, c color.RGBA) {
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
	// AntiAlias is not used to reduce the number of draw-triangles issued.
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}

///////////////////// colors //////////////////////

type Colors struct {
	colors []color.RGBA
}

func NewColors(palette int) *Colors {
	var colors []color.RGBA
	if palette == 0 {
		// https://openprocessing.org/sketch/1845890
		colors = []color.RGBA{
			{0xDE, 0x18, 0x3C, 0xFF},
			{0xF2, 0xB5, 0x41, 0xFF},
			{0x0C, 0x79, 0xBB, 0xFF},
			{0x2D, 0xAC, 0xB2, 0xFF},
			{0xE4, 0x64, 0x24, 0xFF},
			{0xEC, 0xAC, 0xBE, 0xFF},
			// {0x00, 0x00, 0x00, 0xFF},
			{0x19, 0x44, 0x6B, 0xFF},
		}
	} else if palette == 1 {

		// Vincent Willem van Gogh
		// https://goworkship.com/magazine/artist-inspired-color-palettes/
		colors = []color.RGBA{
			{0x00, 0x39, 0x55, 0xff}, // 003955
			{0x39, 0x7e, 0xc0, 0xff}, // 397ec0
			{0x73, 0x38, 0x37, 0xff}, // 733837
			{0xeb, 0xc7, 0x4b, 0xff}, // ebc74b
			{0x60, 0x7a, 0x4d, 0xff}, // 607a4d
		}
	} else {

		colors = []color.RGBA{
			{0x68, 0x8C, 0x89, 0xFF}, //#688C89
			{0xF2, 0xC1, 0x85, 0xFF}, //#F2C185
			{0x73, 0x02, 0x02, 0xFF}, //#730202
			{0xA6, 0x17, 0x17, 0xFF}, //#A61717
			{0xF2, 0x38, 0x38, 0xFF}, //#F23838
		}
	}
	return &Colors{colors: colors}
}
func (c *Colors) Random() color.RGBA {
	i := rand.Intn(len(c.colors))
	return c.colors[i]
}
func (c *Colors) Color(colorNo uint8) color.RGBA {
	return c.colors[colorNo%uint8(len(c.colors))]
}
func (c *Colors) Len() int {
	return len(c.colors)
}
