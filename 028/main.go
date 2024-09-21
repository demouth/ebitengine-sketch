package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/028/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomonobold"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 780
	screenHeight = 520
)

var (
	//go:embed radialblur.kage
	radialblur_kage []byte
	//go:embed and.kage
	and_kage []byte

	faceSource *text.GoTextFaceSource
)

type Game struct {
	count      int
	canvas1    *ebiten.Image
	canvas2    *ebiten.Image
	canvas3    *ebiten.Image
	space      *cp.Space
	shader1    *ebiten.Shader
	shader2    *ebiten.Shader
	ecp        *ebitencp.Drawer
	softbodies []*SoftbodyCircle
}

func (g *Game) Update() error {
	g.count++
	if g.count%30 == 0 {
		x, y := (rand.Float64()-0.5)*5, (rand.Float64()-0.5)*5
		g.space.SetGravity(cp.Vector{X: x, Y: y})
	}
	for i := 0; i < len(g.softbodies); i++ {
		if rand.Float64() > 0.0002 {
			continue
		}
		softbody := g.softbodies[i]
		x, y := (rand.Float64()-0.5)*40, (rand.Float64()-0.5)*40
		for i := 0; i < len(softbody.parts); i++ {
			if rand.Float64() < 0.2 {
				continue
			}
			softbody.parts[i].SetVelocity(x, y)
		}
	}
	g.space.Step(1.0 / 10.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.canvas1.Fill(color.Black)
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)
			vec := circle.TransformC()
			drawer.DrawCircle(
				g.canvas1,
				float32(vec.X),
				float32(vec.Y),
				float32(circle.Radius()),
				color.NRGBA{0xff, 0xff, 0xff, 0xff},
			)
		}
	})
	for _, softbody := range g.softbodies {
		path := vector.Path{}
		for i, part := range softbody.parts {
			part.Position()
			if i == 0 {
				path.MoveTo(float32(part.Position().X), float32(part.Position().Y))
			} else {
				path.LineTo(float32(part.Position().X), float32(part.Position().Y))
			}
		}
		drawer.DrawFill(g.canvas1, path, color.NRGBA{0xff, 0xff, 0xff, 0xff})
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = g.canvas1
	g.canvas2.DrawRectShader(screenWidth, screenHeight, g.shader1, op)

	g.space.StaticBody.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Segment:
			seg := shape.Class.(*cp.Segment)
			drawer.DrawLine(
				screen,
				float32(seg.A().X),
				float32(seg.A().Y),
				float32(seg.B().X),
				float32(seg.B().Y),
				float32(seg.Radius()*2),
				color.RGBA{0x99, 0x99, 0x99, 0xff},
			)
		}
	})

	op = &ebiten.DrawRectShaderOptions{}
	ff := float32(0xFF)
	op.Uniforms = map[string]interface{}{
		"Color1": []float32{0x40 / ff, 0x15 / ff, 0x20 / ff, 1.0}, // #401520
		"Color2": []float32{0xF2 / ff, 0x83 / ff, 0x22 / ff, 1.0}, // #F28322
		"Color3": []float32{0xD9 / ff, 0x25 / ff, 0x25 / ff, 1.0}, // #D92525
		"Color4": []float32{0xF2 / ff, 0xEB / ff, 0xD5 / ff, 1.0}, // #F2EBD5
	}
	op.Images[0] = g.canvas2
	op.Images[1] = g.canvas3
	screen.DrawRectShader(screenWidth, screenHeight, g.shader2, op)

	// cp.DrawSpace(g.space, g.ecp.WithScreen(screen))

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{X: 0, Y: 0})

	s1, err := ebiten.NewShader([]byte(radialblur_kage))
	if err != nil {
		log.Fatal(err)
	}
	s2, err := ebiten.NewShader([]byte(and_kage))
	if err != nil {
		log.Fatal(err)
	}
	s, err := text.NewGoTextFaceSource(bytes.NewReader(gomonobold.TTF))
	if err != nil {
		log.Fatal(err)
	}
	faceSource = s

	game := &Game{}
	game.space = space
	game.canvas1 = ebiten.NewImage(screenWidth, screenHeight)
	game.canvas2 = ebiten.NewImage(screenWidth, screenHeight)
	game.canvas3 = ebiten.NewImage(screenWidth, screenHeight)

	normalFontSize := 300.0
	op := &text.DrawOptions{}
	op.GeoM.Translate(30, -35)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = normalFontSize * 0.8
	text.Draw(game.canvas3, "META\nBALL", &text.GoTextFace{
		Source: faceSource,
		Size:   normalFontSize,
	}, op)

	game.ecp = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.ecp.FlipYAxis = true
	game.ecp.Camera.Offset.X = screenWidth / 2
	game.ecp.Camera.Offset.Y = screenHeight / 2
	game.ecp.AntiAlias = false
	game.shader1 = s1
	game.shader2 = s2

	softbody := game.newSoftbody(160, 32, screenWidth/2, screenHeight/2, 6, 0.000001, 0.8)
	game.softbodies = append(game.softbodies, softbody)
	softbody = game.newSoftbody(100, 20, screenWidth/2-220, screenHeight/2-160, 6, 0.000001, 0.8)
	game.softbodies = append(game.softbodies, softbody)
	softbody = game.newSoftbody(50, 16, screenWidth/2+150, screenHeight/2+200, 6, 0.000001, 0.8)
	game.softbodies = append(game.softbodies, softbody)
	softbody = game.newSoftbody(80, 16, screenWidth/2-200, screenHeight/2+150, 6, 0.000001, 0.8)
	game.softbodies = append(game.softbodies, softbody)

	addWall(space, 0, 0, 0, screenHeight, 1, 1)
	addWall(space, screenWidth, 0, screenWidth, screenHeight, 1, 1)
	addWall(space, 0, 0, screenWidth, 0, 1, 1)
	addWall(space, 0, screenHeight, screenWidth, screenHeight, 1, 1)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("metaball")
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
	shape.SetFriction(0)
	shape.UserData = "wall"
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

func (g *Game) newSoftbody(restLength float64, numParts int, x, y, stiffness, damping, elasticity float64) *SoftbodyCircle {
	return addSoftbodyCircle(g.space, restLength, numParts, x, y, stiffness, damping, elasticity)
}

// soft body circle

type SoftbodyCircle struct {
	numParts   int
	restLength float64
	parts      []*cp.Body
	center     *cp.Body
}

func addSoftbodyCircle(space *cp.Space, restLength float64, numParts int, x, y, stiffness, damping, elasticity float64) *SoftbodyCircle {
	softBody := newSoftbodyCircle(space, restLength, numParts, x, y, stiffness, damping, elasticity)
	return softBody
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
func newSoftbodyCircle(space *cp.Space, restLength float64, numParts int, x, y, stiffness, damping, elasticity float64) *SoftbodyCircle {
	angleStep := math.Pi * 2.0 / float64(numParts)
	partRadius := partRadius(restLength, numParts)
	fmt.Println(restLength, numParts, partRadius)
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
				partRadius, stiffness, damping),
		)
		c.UserData = "part"
		c.SetCollideBodies(false)

		c = space.AddConstraint(
			cp.NewDampedSpring(center, parts[i],
				cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
				restLength, stiffness, damping),
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
