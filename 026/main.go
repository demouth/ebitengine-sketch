package main

import (
	"bytes"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/026/drawer"
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

	if g.time%200 == 0 {
		g.glyphs = append(g.glyphs, addGlyph(g))
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
		if s.UserData != "wall" {
			return // skip
		}

		switch s.Class.(type) {
		case *cp.Segment:
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
	space.Iterations = 20
	space.SetGravity(cp.Vector{X: 0, Y: 100})

	g := &Game{
		space:  space,
		drawer: ebitencp.NewDrawer(screenWidth, screenHeight),
	}
	g.drawer.FlipYAxis = true
	g.drawer.Camera.Offset.X = screenWidth / 2
	g.drawer.Camera.Offset.Y = screenHeight / 2

	g.glyphs = append(g.glyphs, addGlyph(g))

	addWall(space, 260, 500, 340, 500, 30, 0.1)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func addGlyph(g *Game) *Glyph {
	vertices := []ebiten.Vertex{}
	indices := []uint16{}

	path := &vector.Path{}
	{
		face := &text.GoTextFace{Source: textFaceSource, Size: 200}
		op := &text.LayoutOptions{}
		op.LineSpacing = 200
		text.AppendVectorPath(path, "お", face, op)
	}
	vertices, indices = path.AppendVerticesAndIndicesForFilling(vertices, indices)
	s := g.space
	circles := []*cp.Body{}
	x := rand.Float64()*180 + 120
	y := -300.0
	for i := range vertices {
		circle, _ := addCircle(
			s,
			1,
			float64(vertices[i].DstX)+x,
			float64(vertices[i].DstY)+y,
			0.1,
		)
		circles = append(circles, circle)
	}
	for i := range vertices {
		for j := range vertices {
			if i == j {
				continue
			}
			if circles[i].Position().Distance(circles[j].Position()) < 100 {
				c := s.AddConstraint(
					// cp.NewPinJoint(
					// 	circles[i], circles[j],
					// 	cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
					// ),
					cp.NewDampedSpring(
						circles[i], circles[j],
						cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
						circles[i].Position().Distance(circles[j].Position()),
						10.0, 1.2,
					),
				)
				c.SetCollideBodies(false)
			}
		}
	}
	glyph := &Glyph{
		vertices: vertices,
		indices:  indices,
		bodies:   circles,
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

	if y > screenHeight+100 {
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
		g.vertices[i].DstX = float32(g.bodies[i].Position().X)
		g.vertices[i].DstY = float32(g.bodies[i].Position().Y)
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleNonZero
	screen.DrawTriangles(g.vertices, g.indices, whiteImage, op)

	for i := range g.bodies {
		drawer.DrawCircle(
			screen,
			float32(g.bodies[i].Position().X),
			float32(g.bodies[i].Position().Y),
			2,
			color.RGBA{0x00, 0x00, 0x00, 0xff},
		)
	}
}
