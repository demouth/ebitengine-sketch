package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"

	"github.com/ebitengine/microui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 1200
)

type Game struct {
	count int

	shapeWidth      float32
	shapeHeight     float32
	shapeLineWidth  float32
	cache           *Cache
	useCache        bool
	showSprite      bool
	numParticles    float32
	maxNumParticles float32
	ctx             *microui.Context

	gy float32

	particles []*Particle

	colors *Colors
}

func (g *Game) Update() error {
	g.count++

	g.ctx.Begin()
	if g.ctx.BeginWindow("ebitengine/microui", image.Rect(20, 20, 300, 240)) {
		win := g.ctx.GetCurrentContainer()
		win.Rect.Max.X = win.Rect.Min.X + max(win.Rect.Dx(), 240)
		win.Rect.Max.Y = win.Rect.Min.Y + max(win.Rect.Dy(), 100)
		if g.ctx.HeaderEx("Ebitengine Info", microui.OptExpanded) != 0 {
			g.ctx.LayoutRow(2, []int{84, -1}, 0)
			g.ctx.Label("ActualFPS:")
			g.ctx.Label(fmt.Sprintf("%.1f", ebiten.ActualFPS()))
		}
		if g.ctx.HeaderEx("Cache", microui.OptExpanded) != 0 {
			g.ctx.Checkbox("Show Sprite", &g.showSprite)
			calcMax := func(useCache bool) float32 {
				if useCache {
					return 20000
				}
				return 1000
			}
			if g.ctx.Checkbox("Use Cache", &g.useCache) > 0 {
				g.particles = initParticles(int(g.numParticles))
			}
			g.ctx.LayoutBeginColumn()
			g.ctx.LayoutRow(2, []int{46, -1}, 0)
			g.ctx.Label("Num:")
			if g.ctx.SliderEx(&g.numParticles, 100, calcMax(g.useCache), 100, "%.0f", microui.OptAlignCenter) > 0 {
				g.particles = initParticles(int(g.numParticles))
			}
			g.ctx.LayoutEndColumn()
		}
		g.ctx.EndWindow()
	}
	g.ctx.End()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0xff, 0xff, 0xff, 0xff})

	if g.showSprite {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.3, 0.3)
		op.GeoM.Translate(0, 0)
		screen.DrawImage(g.cache.cacheImage, op)
	} else {
		for _, p := range g.particles {
			if g.useCache {
				p.DrawFromCache(screen, g.cache)
			} else {
				p.Draw(screen, g.shapeWidth, g.shapeLineWidth, g.colors)
			}
			p.Move()
		}
	}

	g.ctx.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.colors = NewColors()
	game.shapeWidth = 30
	game.shapeHeight = 30
	game.shapeLineWidth = 1
	game.numParticles = 10000
	game.maxNumParticles = 20000
	game.useCache = true
	game.cache = newCache(game.shapeWidth, game.shapeHeight, game.shapeLineWidth, game.colors)
	game.particles = initParticles(int(game.numParticles))

	game.ctx = microui.NewContext()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Cache drawing results")
	ebiten.RunGame(game)
}

func initParticles(num int) []*Particle {
	particles := []*Particle{}
	for i := 0; i < num; i++ {
		p := &Particle{
			X:       rand.Float32() * screenWidth,
			Y:       rand.Float32() * screenHeight,
			frameNo: uint8(rand.Intn(0xff)),
			colorNo: uint8(rand.Intn(8)),
		}
		particles = append(particles, p)
	}
	return particles
}

///////////////////// colors //////////////////////

type Colors struct {
	colors []color.NRGBA
}

func NewColors() *Colors {
	// https://openprocessing.org/sketch/1845890
	colors := []color.NRGBA{
		{0xDE, 0x18, 0x3C, 0xFF},
		{0xF2, 0xB5, 0x41, 0xFF},
		{0x0C, 0x79, 0xBB, 0xFF},
		{0x2D, 0xAC, 0xB2, 0xFF},
		{0xE4, 0x64, 0x24, 0xFF},
		{0xEC, 0xAC, 0xBE, 0xFF},
		{0x00, 0x00, 0x00, 0xFF},
		{0x19, 0x44, 0x6B, 0xFF},
	}
	return &Colors{colors: colors}
}
func (c *Colors) Random() color.NRGBA {
	i := rand.Intn(len(c.colors))
	return c.colors[i]
}
func (c *Colors) Color(colorNo uint8) color.NRGBA {
	return c.colors[colorNo%uint8(len(c.colors))]
}
func (c *Colors) Len() int {
	return len(c.colors)
}

///////////////////// particle //////////////////////

type Particle struct {
	X, Y    float32
	frameNo uint8
	colorNo uint8
}

func (p *Particle) Draw(screen *ebiten.Image, shapeWidth, lineWidth float32, colors *Colors) {
	drawShape(screen, p.X, p.Y, shapeWidth/2, lineWidth, p.frameNo, p.colorNo, colors)
}
func (p *Particle) DrawFromCache(screen *ebiten.Image, cache *Cache) {
	cache.drawShapeFromeCache(screen, p.X, p.Y, p.frameNo, p.colorNo)
}
func (p *Particle) Move() {
	// p.X += rand.Float32()*2 - 1
	// p.Y += rand.Float32()*2 - 1
	p.frameNo++
}

///////////////////// cached draw //////////////////////

type Cache struct {
	cacheImage     *ebiten.Image
	shapeWidth     float32
	shapeHeight    float32
	shapeLineWidth float32
}

func newCache(shapeWidth, shapeHeight, shapeLineWidth float32, colors *Colors) *Cache {
	cacheImage := ebiten.NewImage(int(shapeWidth*0xff), int(shapeHeight)*colors.Len())

	for i := 0; i < colors.Len(); i++ {
		for j := uint8(0); j < 0xff; j++ {
			dist := shapeWidth
			drawShape(
				cacheImage,
				dist*float32(j)+dist/2,
				dist*float32(i)+dist/2,
				shapeHeight/2,
				shapeLineWidth,
				j,
				uint8(i),
				colors,
			)
		}
	}
	return &Cache{
		cacheImage:     cacheImage,
		shapeWidth:     shapeWidth,
		shapeHeight:    shapeHeight,
		shapeLineWidth: shapeLineWidth,
	}
}

func (c *Cache) drawShapeFromeCache(screen *ebiten.Image, x, y float32, frameNo uint8, colorNo uint8) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(c.shapeWidth/2), -float64(c.shapeHeight/2))
	op.GeoM.Translate(float64(x), float64(y))
	sx := int(frameNo) * int(c.shapeWidth)
	sy := int(colorNo) * int(c.shapeHeight)
	frameWidth := int(c.shapeWidth)
	frameHeight := int(c.shapeHeight)
	screen.DrawImage(
		c.cacheImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image),
		op,
	)
}

///////////////////// draw //////////////////////

var (
	whiteSubImage *ebiten.Image
)

func init() {
	whiteSubImage = ebiten.NewImage(3, 3)
	whiteSubImage.Fill(color.White)
}
func drawLine(screen *ebiten.Image, path vector.Path, c color.NRGBA, width float32, antiAlias bool) {
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
	op.AntiAlias = antiAlias
	screen.DrawTriangles(vs, is, whiteSubImage, op)
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
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func drawShape(screen *ebiten.Image, x, y, length, lineWidth float32, frameNo uint8, colorNo uint8, colors *Colors) {
	r := float32(frameNo) / 255.0
	sr := r * r
	er := 1.0 - (1-r)*(1-r)
	path := vector.Path{}
	const num = 20.0
	for i := 0; i < num; i++ {
		rad := float64(i) * 2.0 * math.Pi / num
		sx := float32(math.Cos(rad)) * sr * length
		sy := float32(math.Sin(rad)) * sr * length
		path.MoveTo(x+sx, y+sy)
		ex := float32(math.Cos(rad)) * er * length
		ey := float32(math.Sin(rad)) * er * length
		path.LineTo(x+ex, y+ey)
		path.Close()
	}
	drawLine(screen, path, colors.Color(colorNo), lineWidth, true)
}
