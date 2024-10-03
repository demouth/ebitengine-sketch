package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 1200
	screenHeight = 1200
)

var (
	//go:embed assets/kingyo2.png
	sprite      []byte
	spriteImage *ebiten.Image
	//go:embed radialblur.kage
	radialblur_kage []byte
)

func loadImage(b []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	origImage := ebiten.NewImageFromImage(img)

	s := origImage.Bounds().Size()
	ebitenImage := ebiten.NewImage(s.X, s.Y)
	op := &ebiten.DrawImageOptions{}
	ebitenImage.DrawImage(origImage, op)
	return ebitenImage
}

type Game struct {
	space  *cp.Space
	ecp    *ebitencp.Drawer
	tofus  Tofus
	shader *ebiten.Shader
}

func (g *Game) Update() error {
	g.tofus.Update()
	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	kingyoImage := ebiten.NewImage(screenWidth, screenHeight)
	// bluredImage := ebiten.NewImage(screenWidth*0.5, screenHeight*0.5)

	for _, tofu := range g.tofus {
		tofu.Draw(kingyoImage)
	}

	// make shadow
	{
		// imageop := &ebiten.DrawImageOptions{}
		// imageop.GeoM.Scale(0.5, 0.5)
		// shadowImage := ebiten.NewImage(bluredImage.Bounds().Dx(), bluredImage.Bounds().Dy())
		// shadowImage.DrawImage(kingyoImage, imageop)
		// shaderop := &ebiten.DrawRectShaderOptions{}
		// shaderop.Images[0] = shadowImage
		// bluredImage.DrawRectShader(bluredImage.Bounds().Dx(), bluredImage.Bounds().Dy(), g.shader, shaderop)
	}

	// draw kingyo and shadow
	{
		// op := &ebiten.DrawImageOptions{}
		// op.GeoM.Scale(2, 2)
		// op.GeoM.Translate(5, 5)
		// screen.DrawImage(bluredImage, op)
		screen.DrawImage(kingyoImage, nil)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	spriteImage = loadImage(sprite)
	shader, err := ebiten.NewShader([]byte(radialblur_kage))
	if err != nil {
		panic(err)
	}

	space := cp.NewSpace()
	space.SetDamping(0.2)
	space.SetGravity(cp.Vector{X: 0, Y: 0})

	tofus := Tofus{}
	for i := 0; i < 220; i++ {
		tofu := NewTofu(space, screenWidth*rand.Float64(), screenHeight*rand.Float64(), 8)
		tofus = append(tofus, tofu)
	}
	for _, tofu := range tofus {
		tofu.Add(rand.Float64()*100-50, rand.Float64()*100-50)
	}

	g := &Game{}
	g.space = space
	g.tofus = tofus
	g.shader = shader

	g.ecp = ebitencp.NewDrawer(0, 0)
	g.ecp.FlipYAxis = true

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("goldfish")
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y, elasticity float64) (*cp.Body, *cp.Shape) {
	mass := 1.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.99)
	shape.UserData = "circle"
	return body, shape
}

// tofu

type Tofus []*Tofu

func (t Tofus) Update() {
	for i := 0; i < len(t); i++ {
		t1 := t[i]
		var (
			sepPosSum = cp.Vector{}
			sepCount  = 0
			aliVelSum = cp.Vector{}
			aliCount  = 0
			cohPosSum = cp.Vector{}
			cohCount  = 0
			avoid     = cp.Vector{}
		)
		const (
			SeparateNeighborhoodRadius = 40.0
			AlignNeighborhoodRadius    = 80.0
			CohesionNeighborhoodRadius = 80.0
			SeparateWeight             = 100
			AlignmentWeight            = 1
			CohesionWeight             = 1
			AvoidWallWeight            = 1
		)
		for j := 0; j < len(t); j++ {
			if i == j {
				continue
			}
			t2 := t[j]
			dist := t1.head.Position().Distance(t2.head.Position())
			// separation
			if dist < SeparateNeighborhoodRadius {
				sepPosSum = sepPosSum.Add(t1.head.Position().Sub(t2.head.Position()).Mult(1.0 / dist))
				sepCount++
			}
			if t1.color == t2.color {
				// alignment
				if dist < AlignNeighborhoodRadius {
					aliVelSum = aliVelSum.Add(t2.head.Velocity())
					aliCount++
				}
				// cohesion
				if dist < CohesionNeighborhoodRadius {
					cohPosSum = cohPosSum.Add(t2.head.Position())
					cohCount++
				}
			}
		}
		// separation
		if sepCount != 0 {
			sepPosSum = sepPosSum.Mult(1.0 / float64(sepCount))
		}
		// alignment
		if aliCount != 0 {
			aliVelSum = aliVelSum.Mult(1.0 / float64(aliCount))
		}
		// cohesion
		if cohCount != 0 {
			cohPosSum = cohPosSum.Mult(1.0 / float64(cohCount))
			cohPosSum = cohPosSum.Sub(t1.head.Position())
			cohPosSum = cohPosSum.Normalize()
			cohPosSum = cohPosSum.Mult(2)
		}

		// avoid wall
		{
			center := cp.Vector{X: screenWidth / 2, Y: screenHeight / 2}
			dist := t1.head.Position().Distance(center)
			edgeBoundary := float64(screenWidth) * 0.4
			if dist > edgeBoundary {
				overhang := dist - edgeBoundary
				avoid = t1.head.Position().Sub(center).Normalize().Mult(-overhang * 0.2)
			}
		}

		vec := cp.Vector{}
		vec = vec.Add(sepPosSum).Mult(SeparateWeight)
		vec = vec.Add(aliVelSum).Mult(AlignmentWeight)
		vec = vec.Add(cohPosSum).Mult(CohesionWeight)
		vec = vec.Add(avoid).Mult(AvoidWallWeight)
		vec = vec.Normalize()
		vec = vec.Mult(3)
		t[i].Add(vec.X, vec.Y)
		// When it stops, move it in a random direction
		if t[i].vec.Length() < 1 {
			vec := cp.Vector{X: (rand.Float64() - 0.5) * 100, Y: (rand.Float64() - 0.5) * 100}
			t[i].Add(vec.X, vec.Y)
		}
	}
	for i := 0; i < len(t); i++ {
		t[i].Move()
	}
}

const (
	TofuColorRed uint8 = iota
	TofuColorBlack
)

type Tofu struct {
	circles [][]*cp.Body
	vx, vy  float64
	vec     cp.Vector
	head    *cp.Body
	color   uint8
}

func (t *Tofu) Move() {
	t.head.SetVelocity(t.vec.X, t.vec.Y)
	t.vec = t.vec.Mult(0.97)
}
func (t *Tofu) Add(x, y float64) {
	t.vec = t.vec.Add(cp.Vector{X: x, Y: y})
}

func (t *Tofu) Draw(screen *ebiten.Image) {
	vertices := []ebiten.Vertex{}
	indices := []uint16{}
	for y := 0; y < len(t.circles); y++ {
		startIndex := uint16(len(vertices) - len(t.circles[y]))
		for x := 0; x < len(t.circles[y]); x++ {
			vertex := ebiten.Vertex{
				DstX:   float32(t.circles[y][x].Position().X),
				DstY:   float32(t.circles[y][x].Position().Y),
				SrcX:   float32(200) * float32(x) / float32(len(t.circles[y])-1),
				SrcY:   float32(800) * float32(y) / float32(len(t.circles)-1),
				ColorA: 1,
			}
			if t.color != TofuColorRed {
				vertex.ColorR = 1
				vertex.ColorG = 1
				vertex.ColorB = 1
			} else {
				vertex.ColorR = 0
				vertex.ColorG = 0
				vertex.ColorB = 0
			}
			vertices = append(vertices, vertex)
		}
		if y == 0 {
			continue
		}
		indices = append(indices, startIndex+0, startIndex+1, startIndex+3, startIndex+1, startIndex+4, startIndex+3)
		indices = append(indices, startIndex+1, startIndex+2, startIndex+4, startIndex+2, startIndex+5, startIndex+4)
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleFillAll
	op.AntiAlias = false
	screen.DrawTriangles(
		vertices,
		indices,
		spriteImage,
		op,
	)
}

func NewTofu(space *cp.Space, startX, startY, step float64) *Tofu {
	type circleForJoints struct {
		body *cp.Body
		x    int
		y    int
	}
	stepX := step
	stepY := step
	radius := stepX * 0.6
	circles := [][]*cp.Body{}
	circlesForJoints := []circleForJoints{}
	const numY = 8
	for y := 0; y < numY; y++ {
		circles = append(circles, []*cp.Body{})
		for x := 0; x < 3; x++ {
			circle, shape := addCircle(space, radius, startX+stepX*float64(x), startY+stepY*float64(y), 0.0)
			// shape.SetFilter(cp.SHAPE_FILTER_NONE)
			shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
			circles[y] = append(circles[y], circle)
			circlesForJoints = append(circlesForJoints, circleForJoints{
				body: circle,
				x:    x,
				y:    y,
			})
		}
	}

	l := len(circlesForJoints)
	hypot := math.Hypot(stepX, stepY)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			if circlesForJoints[i].body.Position().Distance(circlesForJoints[j].body.Position()) <= hypot*1.1 {
				stiffness, damping := 1500.0, 30.0
				if circlesForJoints[j].y >= numY/2.0 {
					r := 1 - float64(circlesForJoints[j].y-numY/2.0)/(numY/2.0)
					stiffness = 1 + 150.0*r
					damping = 0.1 + 15.5*r
				}
				c := space.AddConstraint(
					cp.NewDampedSpring(
						circlesForJoints[i].body, circlesForJoints[j].body,
						cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
						circlesForJoints[i].body.Position().Distance(circlesForJoints[j].body.Position()),
						stiffness,
						damping,
					),
				)
				c.SetCollideBodies(false)
			}
		}
	}
	color := TofuColorRed
	if rand.Float64() < 0.5 {
		color = TofuColorBlack
	}
	t := &Tofu{
		circles: circles,
		head:    circles[0][1],
		color:   color,
	}
	return t
}
