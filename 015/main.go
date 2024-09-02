package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/ebitengine/microui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 600
	screenHeight = 600
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

var (
	whiteSubImage = ebiten.NewImage(3, 3)
)
var (
	softbodies []*SoftbodyCircle = make([]*SoftbodyCircle, 0)
)

type Game struct {
	count       int
	debugDrawer *ebitencp.Drawer
	debugMode   bool

	space *cp.Space

	gx   float32
	gy   float32
	step float64

	RestLength float32
	Stiffness  float32
	Damping    float32

	ctx *microui.Context
}

func (g *Game) Update() error {
	g.ProcessFrame()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0xff, 0xff, 0xff, 0xff})
	g.ctx.Draw(screen)
	if g.debugMode {
		cp.DrawSpace(g.space, g.debugDrawer.WithScreen(screen))
	} else {

		for _, sb := range softbodies {
			path := vector.Path{}
			for i, p := range sb.parts {
				pos := p.Position()
				x := float32(pos.X) + screenWidth/2
				y := float32(pos.Y) + screenHeight/2
				if i == 0 {
					path.MoveTo(x, y)
				} else {
					path.LineTo(x, y)
				}
				p.Position()
				p.SetPosition(p.Position())
			}
			path.Close()
			g.drawFill(screen, path, sb.color)
			partRadius := partRadius(sb.restLength, sb.numParts)
			g.drawLine(screen, path, sb.color, float32(partRadius*2))

			{
				// eyes
				p := sb.center.Position()
				v := sb.center.Velocity()
				rad := math.Atan2(v.Y, v.X)
				eyeR := float32(8)
				eyeD := float32(7)
				eyeX := float32(math.Cos(rad)) * 5
				eyeY := float32(math.Sin(rad)) * 5
				eyeX += float32(p.X)
				eyeY += float32(p.Y)
				corneaR := float32(4)
				corneaX := float32(math.Cos(rad)) * eyeR * 0.5
				corneaY := float32(math.Sin(rad)) * eyeR * 0.5
				g.drawCircle(screen, screenWidth/2-eyeD+eyeX, screenHeight/2+eyeY, eyeR, color.NRGBA{0xff, 0xff, 0xff, 0xff})
				g.drawCircle(screen, screenWidth/2+eyeD+eyeX, screenHeight/2+eyeY, eyeR, color.NRGBA{0xff, 0xff, 0xff, 0xff})
				g.drawCircle(screen, screenWidth/2-eyeD+eyeX+corneaX, screenHeight/2+eyeY+corneaY, corneaR, sb.color)
				g.drawCircle(screen, screenWidth/2+eyeD+eyeX+corneaX, screenHeight/2+eyeY+corneaY, corneaR, sb.color)
			}
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func init() {
	whiteSubImage.Fill(color.White)
}
func main() {
	game := &Game{}
	game.step = 1.0 / 60.0
	game.gy = 200
	game.debugDrawer = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.debugDrawer.FlipYAxis = true
	game.debugMode = false
	game.RestLength = 50
	game.Stiffness = 600
	game.Damping = 32
	game.ctx = microui.NewContext()

	space := cp.NewSpace()
	space.Iterations = 20
	gravity := cp.Vector{X: float64(game.gx), Y: float64(game.gy)}
	space.SetGravity(gravity)
	game.space = space

	addWall(space, -screenWidth/2, screenHeight/2, screenWidth/2, screenHeight/2, 1, 0.9)
	// addWall(space, -screenWidth/2, -screenHeight/2, screenWidth/2, -screenHeight/2, 1, 0.9)
	addWall(space, -screenWidth/2, -screenHeight/2, -screenWidth/2, screenHeight/2, 1, 0.9)
	addWall(space, screenWidth/2, -screenHeight/2, screenWidth/2, screenHeight/2, 1, 0.9)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y, elasticity float64) (*cp.Body, *cp.Shape) {
	mass := 1.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.7)
	shape.UserData = "circle"
	return body, shape
}
func addWall(space *cp.Space, x1, y1, x2, y2, radius, elasticity float64) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	shape := space.AddShape(cp.NewSegment(space.StaticBody, pos1, pos2, radius))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.9)
	shape.UserData = "wall"
}

func addSoftbodyCircle(space *cp.Space, restLength float64, x, y, elasticity float64) *SoftbodyCircle {
	softBody := newSoftbodyCircle(space, restLength, x, y, elasticity)
	return softBody
}

func (g *Game) drawLine(screen *ebiten.Image, path vector.Path, c color.NRGBA, width float32) {
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
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func (g *Game) drawFill(screen *ebiten.Image, path vector.Path, c color.NRGBA) {
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
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func (g *Game) drawCircle(screen *ebiten.Image, x, y, radius float32, c color.NRGBA) {
	path := vector.Path{}
	path.Arc(x, y, radius, 0, 2*math.Pi, vector.Clockwise)
	g.drawFill(screen, path, c)
}
func (g *Game) ProcessFrame() {
	g.ctx.Begin()
	TestWindow(g.ctx, g)
	g.ctx.End()

	g.count++
	if g.count%80 == 0 {
		softbodies = append(softbodies, addSoftbodyCircle(g.space, float64(g.RestLength), rand.Float64()*200.0-100.0, -400, 0.1))
	}
	newSoftbodies := make([]*SoftbodyCircle, 0)
	for _, sb := range softbodies {

		sb.aging()
		if sb.isDied() {
			sb.remove(g.space)
			continue
		}

		remain := (sb.lifespan - sb.age) / sb.lifespan
		rlRemain := remain
		sfRemina := remain
		if remain < 0.1 {
			rlRemain = rlRemain / 0.1
			rlRemain = math.Pow(rlRemain, 2)
			sfRemina = math.Pow(sfRemina, 2)
		} else {
			rlRemain = 1.0
			sfRemina = 1.0
		}
		sb.setParams(float64(g.RestLength)*rlRemain, float64(g.Stiffness)*sfRemina, float64(g.Damping))
		newSoftbodies = append(newSoftbodies, sb)
	}
	softbodies = newSoftbodies

	g.space.SetGravity(cp.Vector{X: float64(g.gx), Y: float64(g.gy)})

	g.space.Step(g.step)
	g.debugDrawer.HandleMouseEvent(g.space)
}

func TestWindow(ctx *microui.Context, g *Game) {
	if ctx.BeginWindowEx("ebitengine/microui", image.Rect(20, 20, 300, 240), microui.OptClosed) != 0 {
		defer ctx.EndWindow()
		win := ctx.GetCurrentContainer()
		win.Rect.Max.X = win.Rect.Min.X + max(win.Rect.Dx(), 240)
		win.Rect.Max.Y = win.Rect.Min.Y + max(win.Rect.Dy(), 100)

		if ctx.HeaderEx("Space", microui.OptExpanded) != 0 {
			ctx.LayoutBeginColumn()
			ctx.LayoutRow(2, []int{100, -1}, 0)
			ctx.Label("Gravity:")
			ctx.Slider(&g.gy, 200, 400)
			ctx.LayoutEndColumn()
		}
		if ctx.HeaderEx("Soft Body", microui.OptExpanded) != 0 {
			ctx.LayoutBeginColumn()
			ctx.LayoutRow(2, []int{100, -1}, 0)
			ctx.Label("RestLength:")
			ctx.Slider(&g.RestLength, 30, 55)
			ctx.Label("Stiffness:")
			ctx.Slider(&g.Stiffness, 0, 1000)
			ctx.Label("Damping:")
			ctx.Slider(&g.Damping, 0, 50)
			ctx.LayoutEndColumn()
			if ctx.Button("Switch Draw") {
				g.debugMode = !g.debugMode
			}
		}
	}
}

func removeShapeCallback(space *cp.Space, key interface{}, data interface{}) {
	var s *cp.Shape
	var ok bool
	if s, ok = key.(*cp.Shape); !ok {
		return
	}
	s.Body().EachConstraint(func(c *cp.Constraint) {
		space.RemoveConstraint(c)
	})
	space.RemoveBody(s.Body())
	space.RemoveShape(s)
}
func removeBodyCallback(space *cp.Space, key interface{}, data interface{}) {
	var b *cp.Body
	var ok bool
	if b, ok = key.(*cp.Body); !ok {
		return
	}
	space.RemoveBody(b)
	b.EachShape(func(s *cp.Shape) {
		space.RemoveShape(s)
	})
}

// soft ball

type SoftbodyCircle struct {
	numParts   int
	restLength float64
	parts      []*cp.Body
	center     *cp.Body
	color      color.NRGBA

	random float64

	lifespan float64
	age      float64
	died     bool
}

func (sc *SoftbodyCircle) aging() {
	sc.age++
	if sc.age > sc.lifespan {
		sc.died = true
	}
}
func (sc *SoftbodyCircle) isDied() bool {
	return sc.died
}
func (sc *SoftbodyCircle) remove(space *cp.Space) {
	sc.center.EachShape(func(s *cp.Shape) {
		space.AddPostStepCallback(removeShapeCallback, s, nil)
	})
	for _, p := range sc.parts {
		p.EachShape(func(s *cp.Shape) {
			space.AddPostStepCallback(removeShapeCallback, s, nil)
		})
	}
}
func partRadius(restLength float64, numParts int) float64 {
	return math.Pi * restLength * 2.0 / float64(numParts+1) * 0.8
}
func (sc *SoftbodyCircle) setParams(restLength, stiffness, damping float64) {
	restLength *= sc.random
	sc.restLength = restLength
	sc.center.EachConstraint(func(c *cp.Constraint) {
		switch c.Class.(type) {
		case *cp.DampedSpring:
			if c.UserData == "center" {
				s := c.Class.(*cp.DampedSpring)
				s.RestLength = restLength
				s.Stiffness = stiffness
				s.Damping = damping
			}
		}
	})
	partRadius := partRadius(restLength, 32)
	for _, p := range sc.parts {
		p.EachConstraint(func(c *cp.Constraint) {
			switch c.Class.(type) {
			case *cp.DampedSpring:
				if c.UserData == "part" {
					s := c.Class.(*cp.DampedSpring)
					s.RestLength = partRadius
				}
			}
		})
	}
	for _, p := range sc.parts {
		p.EachShape(func(s *cp.Shape) {
			switch s.Class.(type) {
			case *cp.Circle:
				c := s.Class.(*cp.Circle)
				c.SetRadius(partRadius)
			}
		})
	}
}
func newSoftbodyCircle(space *cp.Space, restLength float64, x, y, elasticity float64) *SoftbodyCircle {
	numParts := 32
	angleStep := math.Pi * 2.0 / float64(numParts)
	partRadius := partRadius(restLength, numParts)
	parts := make([]*cp.Body, numParts)
	center, _ := addCircle(space, 1, x, y, elasticity)

	for i := 0; i < numParts; i++ {
		part, _ := addCircle(space, partRadius, x+restLength*math.Cos(angleStep*float64(i)), y+restLength*math.Sin(angleStep*float64(i)), elasticity)
		parts[i] = part
	}

	for i := 0; i < numParts; i++ {
		neighborIndex := (i + 1) % numParts
		c := space.AddConstraint(
			cp.NewDampedSpring(
				parts[i], parts[neighborIndex],
				cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
				partRadius, 1010, 0.3),
		)
		c.UserData = "part"
		c.SetCollideBodies(false)

		c = space.AddConstraint(
			cp.NewDampedSpring(center, parts[i],
				cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
				50, 1000, 0.3),
		)
		c.UserData = "center"
	}

	sc := &SoftbodyCircle{
		lifespan:   1900,
		age:        0,
		numParts:   numParts,
		restLength: restLength,
		parts:      parts,
		center:     center,
		random:     rand.Float64()*0.8 + 0.6,
		color:      colors.Random(),
	}
	return sc
}

///////////////////// colors //////////////////////

var (
	colors = NewColors()
)

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
