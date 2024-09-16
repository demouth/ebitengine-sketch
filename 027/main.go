package main

import (
	"bytes"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/027/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	textFaceSource *text.GoTextFaceSource
	whiteImage     *ebiten.Image
)

func init() {
	whiteImage = ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)

	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	textFaceSource = s
}

type Game struct {
	time   int
	space  *cp.Space
	drawer *ebitencp.Drawer
	glyphs []*Glyph
}

func (g *Game) Update() error {
	g.time++

	if g.time%140 == 0 {
		runes := []rune{'V', 'A', 'I', 'Y', 'T', 'N'}
		r := runes[rand.Intn(len(runes))]

		g.glyphs = append(g.glyphs, addGlyph(g, string(r)))
	}

	newGlyphs := []*Glyph{}
	for _, gl := range g.glyphs {
		if gl.OffScreen(g.space) {
			gl.Remove(g.space)
		} else {
			newGlyphs = append(newGlyphs, gl)
		}
	}
	g.glyphs = newGlyphs

	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0x66, 0x66, 0x66, 0xff})
	// cp.DrawSpace(g.space, g.drawer.WithScreen(screen))
	for _, gl := range g.glyphs {
		gl.Draw(screen)
	}
	g.space.EachShape(func(s *cp.Shape) {
		// if s.UserData != "wall" {
		// return // skip
		// }

		switch s.Class.(type) {
		case *cp.Segment:
			if s.UserData == "segment" {
				return
			}
			segment := s.Class.(*cp.Segment)
			ta := segment.TransformA()
			tb := segment.TransformB()
			drawer.DrawLine(
				screen,
				float32(ta.X), float32(ta.Y),
				float32(tb.X), float32(tb.Y),
				float32(segment.Radius()*2),
				color.RGBA{0xff, 0xff, 0xff, 0xff},
			)
		}
	})

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f\nNUM: %d",
		ebiten.ActualFPS(),
		len(g.glyphs),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	space := cp.NewSpace()
	space.Iterations = 50
	space.SetGravity(cp.Vector{X: 0, Y: 100})

	g := &Game{
		space:  space,
		drawer: ebitencp.NewDrawer(screenWidth, screenHeight),
	}
	g.drawer.FlipYAxis = true
	g.drawer.Camera.Offset.X = screenWidth / 2
	g.drawer.Camera.Offset.Y = screenHeight / 2

	g.glyphs = append(g.glyphs, addGlyph(g, "A"))

	addWall(space, 220, 500, 340, 500, 30, 0.1)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
func addGlyph(g *Game, str string) *Glyph {
	vertices := []ebiten.Vertex{}
	indices := []uint16{}

	path := &vector.Path{}
	{
		face := &text.GoTextFace{Source: textFaceSource, Size: 200}
		op := &text.LayoutOptions{}
		op.LineSpacing = 200
		text.AppendVectorPath(path, str, face, op)
	}
	vertices, indices = path.AppendVerticesAndIndicesForFilling(vertices, indices)
	s := g.space
	segments := []*cp.Body{}
	// x := rand.Float64()*180 + 120
	// y := -0.0

	avgX := 0.0
	avgY := 0.0
	for i := range vertices {
		avgX += float64(vertices[i].DstX)
		avgY += float64(vertices[i].DstY)
	}
	avgX /= float64(len(vertices))
	avgY /= float64(len(vertices))
	for i := range vertices {
		vertices[i].DstX -= float32(avgX)
		vertices[i].DstY -= float32(avgY)
	}

	for i := range vertices {
		// j := i + 1
		// if j == len(vertices) {
		// 	continue
		// }
		j := (i + 1) % len(vertices)
		segment, _ := addSegment(
			s,
			float64(vertices[i].DstX),
			float64(vertices[i].DstY),
			float64(vertices[j].DstX),
			float64(vertices[j].DstY),
			0.01,
			0.0,
			0.0,
		)
		segments = append(segments, segment)
	}

	m := map[*cp.Body]map[*cp.Body]bool{}
	for i := 0; i < len(segments); i++ {
		j := (i + 1) % len(segments)
		p1 := cp.Vector{X: float64(vertices[j].DstX), Y: float64(vertices[j].DstY)}
		p2 := cp.Vector{X: float64(vertices[j].DstX), Y: float64(vertices[j].DstY)}
		c := s.AddConstraint(
			cp.NewPinJoint(
				segments[i], segments[j],
				p1,
				p2,
			),
			// cp.NewDampedSpring(
			// 	segments[i], segments[j],
			// 	p1,
			// 	p2,
			// 	p1.Distance(p2),
			// 	430.0, 120.2,
			// ),
		)
		fmt.Println(i, j)
		c.SetCollideBodies(false)
		if m[segments[i]] == nil {
			m[segments[i]] = map[*cp.Body]bool{}
		}
		if m[segments[j]] == nil {
			m[segments[j]] = map[*cp.Body]bool{}
		}
		m[segments[i]][segments[j]] = true
		m[segments[j]][segments[i]] = true
	}
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if i == j {
				continue
			}
			if m[segments[i]][segments[j]] {
				continue
			}

			p1 := cp.Vector{X: float64(vertices[j].DstX), Y: float64(vertices[j].DstY)}
			p2 := cp.Vector{X: float64(vertices[j].DstX), Y: float64(vertices[j].DstY)}
			if p1.Distance(p2) < 50 {
				c := s.AddConstraint(
					cp.NewPinJoint(
						segments[i], segments[j],
						p1,
						p2,
					),
					// cp.NewDampedSpring(
					// 	segments[i], segments[j],
					// 	p1,
					// 	p2,
					// 	p1.Distance(p2),
					// 	430.0, 120.2,
					// ),
				)
				c.SetCollideBodies(false)
			}
		}
	}

	for _, segment := range segments {
		segment.SetPosition(cp.Vector{X: 300, Y: 0})
	}

	glyph := &Glyph{
		vertices: vertices,
		indices:  indices,
		bodies:   segments,
	}
	return glyph
}
func addCircle(space *cp.Space, radius float64, x, y, elasticity float64) (*cp.Body, *cp.Shape) {
	mass := 3.0
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

var (
	counter uint
)

func addSegment(space *cp.Space, x1, y1, x2, y2, radius, elasticity, friction float64) (*cp.Body, *cp.Shape) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	mass := 100.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForSegment(mass, pos1, pos2, 0)))
	shape := space.AddShape(cp.NewSegment(body, pos1, pos2, radius))
	shape.SetElasticity(elasticity)
	shape.SetFriction(friction)
	shape.UserData = "segment"
	body.UserData = "segment"

	shape.SetFilter(cp.NewShapeFilter(counter, 1, 1))
	counter++
	return body, shape
}
func removeBodyCallback(space *cp.Space, key interface{}, data interface{}) {
	var b *cp.Body
	var ok bool
	if b, ok = key.(*cp.Body); !ok {
		return
	}
	b.EachConstraint(func(c *cp.Constraint) {
		space.RemoveConstraint(c)
	})
	space.RemoveBody(b)
	b.EachShape(func(s *cp.Shape) {
		space.RemoveShape(s)
	})
}

// Glyph is a collection of cp.Bodies that make up a glyph

type Glyph struct {
	vertices []ebiten.Vertex
	indices  []uint16
	bodies   []*cp.Body
}

func (g *Glyph) OffScreen(space *cp.Space) bool {
	var y float64 = 0
	for i := range g.bodies {
		y += g.bodies[i].Position().Y
	}
	y /= float64(len(g.bodies))

	if y > screenHeight {
		return true
	}
	return false
}
func (g *Glyph) Remove(space *cp.Space) {
	for i := range g.bodies {
		space.AddPostStepCallback(removeBodyCallback, g.bodies[i], nil)
	}
}
func (g *Glyph) Draw(screen *ebiten.Image) {
	for i := range g.vertices {
		g.bodies[i].EachShape(func(s *cp.Shape) {
			switch s.Class.(type) {
			case *cp.Segment:
				segment := s.Class.(*cp.Segment)
				ta := segment.TransformA()
				g.vertices[i].DstX = float32(ta.X)
				g.vertices[i].DstY = float32(ta.Y)
			}
		})
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleNonZero
	screen.DrawTriangles(g.vertices, g.indices, whiteImage, op)
}
