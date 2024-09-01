package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/demouth/ebitencp"
	"github.com/ebitengine/microui"
	"github.com/hajimehoshi/ebiten/v2"
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

type Game struct {
	count       int
	debugDrawer *ebitencp.Drawer
	debugMode   bool
	cameraX     float64
	cameraY     float64
	space       *cp.Space

	gx   float64
	gy   float64
	step float64

	softbody *SoftbodyCircle

	RestLength float64
	Stiffness  float64
	Damping    float64

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

		sb := g.softbody
		path := vector.Path{}
		for i, p := range sb.parts {
			pos := p.Position()
			x := float32(pos.X) + screenWidth/2
			y := float32(pos.Y) + screenHeight/2
			x -= float32(g.cameraX)
			y -= float32(g.cameraY)
			if i == 0 {
				path.MoveTo(x, y)
			} else {
				path.LineTo(x, y)
			}
			p.Position()
			p.SetPosition(p.Position())
		}
		path.Close()
		g.drawFill(screen, path, color.NRGBA{0x00, 0x00, 0x00, 0xff})
		partRadius := partRadius(sb.restLength, sb.numParts)
		g.drawLine(screen, path, color.NRGBA{0x00, 0x00, 0x00, 0xff}, float32(partRadius*2))

		g.space.EachShape(func(shape *cp.Shape) {
			switch shape.Class.(type) {
			case *cp.Segment:
				segment := shape.Class.(*cp.Segment)
				ta := segment.TransformA()
				tb := segment.TransformB()
				cx := float32(g.cameraX)
				cy := float32(g.cameraY)
				var path vector.Path
				path.MoveTo(
					float32(ta.X)-cx+float32(screenWidth)/2,
					float32(ta.Y)-cy+float32(screenHeight)/2)
				path.LineTo(
					float32(tb.X)-cx+float32(screenWidth)/2,
					float32(tb.Y)-cy+float32(screenHeight)/2)
				path.Close()
				g.drawLine(screen, path, color.NRGBA{0x00, 0x00, 0x00, 0xff}, float32(segment.Radius()*2))
			}
		})

		{
			// eyes
			p := g.softbody.center.Position()
			v := g.softbody.center.Velocity()
			rad := math.Atan2(v.Y, v.X)
			eyeR := float32(10)
			eyeD := float32(13)
			eyeX := float32(math.Cos(rad)) * 15
			eyeY := float32(math.Sin(rad)) * 15
			eyeX += float32(p.X) - float32(g.cameraX)
			eyeY += float32(p.Y) - float32(g.cameraY)
			corneaR := float32(5)
			corneaX := float32(math.Cos(rad)) * eyeR * 0.5
			corneaY := float32(math.Sin(rad)) * eyeR * 0.5
			g.drawCircle(screen, screenWidth/2-eyeD+eyeX, screenHeight/2+eyeY, eyeR, color.NRGBA{0xff, 0xff, 0xff, 0xff})
			g.drawCircle(screen, screenWidth/2+eyeD+eyeX, screenHeight/2+eyeY, eyeR, color.NRGBA{0xff, 0xff, 0xff, 0xff})
			g.drawCircle(screen, screenWidth/2-eyeD+eyeX+corneaX, screenHeight/2+eyeY+corneaY, corneaR, color.NRGBA{0x00, 0x00, 0x00, 0xff})
			g.drawCircle(screen, screenWidth/2+eyeD+eyeX+corneaX, screenHeight/2+eyeY+corneaY, corneaR, color.NRGBA{0x00, 0x00, 0x00, 0xff})
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func init() {
	whiteSubImage.Fill(color.White)
}
func (g *Game) newSoftbody() *SoftbodyCircle {
	return addSoftbodyCircle(g.space, float64(g.RestLength), 0, -400, 0.9)
}
func main() {
	game := &Game{}
	game.step = 1.0 / 60.0
	game.gy = 100
	game.debugDrawer = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.debugDrawer.FlipYAxis = true
	game.debugMode = false
	game.RestLength = 50
	game.Stiffness = 200
	game.Damping = 13
	game.ctx = microui.NewContext()

	space := cp.NewSpace()
	space.Iterations = 20
	gravity := cp.Vector{X: float64(game.gx), Y: float64(game.gy)}
	space.SetGravity(gravity)
	game.space = space
	game.softbody = game.newSoftbody()
	game.cameraX = game.softbody.center.Position().X
	game.cameraY = game.softbody.center.Position().Y

	addWall(space, -400, 0, -65, 110, 30, 0.99)
	addWall(space, 400, 0, 65, 110, 30, 0.99)
	addWall(space, -60, 110, -60, 200, 30, 0.99)
	addWall(space, 60, 110, 60, 200, 30, 0.99)
	addWall(space, 300, 400, -40, 450, 10, 0.99)
	addWall(space, -300, 600, 20, 700, 10, 0.99)
	addWall(space, 130, 600, 130, 1900, 10, 0.99)
	addWall(space, 20, 700, 20, 2000, 10, 0.99)
	addWall(space, 20, 2000, 40, 2040, 10, 0.99)
	addWall(space, 40, 2040, 60, 2070, 10, 0.99)
	addWall(space, 60, 2070, 80, 2090, 10, 0.99)
	addWall(space, 80, 2090, 100, 2100, 10, 0.99)
	addWall(space, 100, 2100, 140, 2110, 10, 0.99)
	addWall(space, 140, 2110, 200, 2100, 10, 0.99)
	addWall(space, 200, 2100, 300, 2030, 10, 0.99)
	addWall(space, 900, 2500, 1100, 2500, 20, 0.99)
	addWall(space, 900, 2400, 900, 2500, 20, 0.99)
	addWall(space, 1100, 2400, 1100, 2500, 20, 0.99)

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
	shape.SetFriction(0.2)
	shape.UserData = "circle"
	return body, shape
}
func addWall(space *cp.Space, x1, y1, x2, y2, radius, elasticity float64) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	shape := space.AddShape(cp.NewSegment(space.StaticBody, pos1, pos2, radius))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.2)
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
	g.ctx.Update(func() {
		ctx := g.ctx
		ctx.Window("ebitengine/microui", image.Rect(20, 20, 300, 300), func(res microui.Res) {
			win := ctx.CurrentContainer()
			win.Rect.Max.X = win.Rect.Min.X + max(win.Rect.Dx(), 240)
			win.Rect.Max.Y = win.Rect.Min.Y + max(win.Rect.Dy(), 100)

			if ctx.HeaderEx("Space", microui.OptExpanded) != 0 {
				ctx.LayoutRow(2, []int{100, -1}, 0)
				ctx.Label("Gravity:")
				ctx.Slider(&g.gy, 00, 400)
				g.ctx.Label("ActualFPS:")
				g.ctx.Label(fmt.Sprintf("%.1f", ebiten.ActualFPS()))
				g.ctx.Label("Camera:")
				g.ctx.Label(fmt.Sprintf("X:%.1f, Y:%.1f", g.cameraX, g.cameraY))
				g.ctx.Label("Game Count:")
				g.ctx.Label(fmt.Sprintf("%v", g.count))

			}
			if ctx.HeaderEx("Soft Body", microui.OptExpanded) != 0 {
				ctx.LayoutRow(2, []int{100, -1}, 0)
				ctx.Label("RestLength:")
				ctx.Slider(&g.RestLength, 30, 55)
				ctx.Label("Stiffness:")
				ctx.Slider(&g.Stiffness, 0, 1000)
				ctx.Label("Damping:")
				ctx.Slider(&g.Damping, 0, 50)
				ctx.Checkbox("Debug Mode", &g.debugMode)
			}
		})
	})

	g.count++
	if g.count > 1900 {
		g.count = 0
		g.softbody.remove(g.space)
		g.softbody = g.newSoftbody()
		g.cameraX = g.softbody.center.Position().X
		g.cameraY = g.softbody.center.Position().Y
	}
	sb := g.softbody
	sb.setParams(float64(g.RestLength), float64(g.Stiffness), float64(g.Damping))

	dffX := sb.center.Position().X - g.cameraX
	dffY := sb.center.Position().Y - g.cameraY
	g.cameraX += dffX * 0.05
	g.cameraY += dffY * 0.05

	g.space.SetGravity(cp.Vector{X: float64(g.gx), Y: float64(g.gy)})
	g.space.Step(g.step)
	g.debugDrawer.HandleMouseEvent(g.space)
	g.debugDrawer.Camera.Offset.X = g.cameraX
	g.debugDrawer.Camera.Offset.Y = g.cameraY
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
	return math.Pi * restLength * 2.0 / float64(numParts+1) * 0.7
}
func (sc *SoftbodyCircle) setParams(restLength, stiffness, damping float64) {
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
	partRadius := partRadius(restLength, sc.numParts)
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
		numParts:   numParts,
		restLength: restLength,
		parts:      parts,
		center:     center,
	}
	return sc
}
