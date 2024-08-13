package main

// This is based on "jakecoffman/cp-examples/march".

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/013/assets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 600
	screenHeight = 600
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2
)

type Game struct {
	count  int
	space  *cp.Space
	drawer *ebitencp.Drawer
}

func (g *Game) Update() error {
	g.count++
	if g.count%10 == 0 {
		addRandomFruit(g.space)
	}

	// cp.SpaceCollideShapesFunc(shape, shape2, func(arb *cp.Arbiter) {
	// 	// fmt.Println("collision")
	// })

	g.drawer.HandleMouseEvent(g.space)

	g.space.Step(1 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	g.drawer.Screen = screen

	g.space.EachShape(func(shape *cp.Shape) {

		switch shape.Class.(type) {
		case *cp.PolyShape:
			circle := shape.Class.(*cp.PolyShape)
			vec := circle.Body().Position()

			img := assets.Get(circle.Body().UserData.(int)).EbitenImage
			size := img.Bounds().Size()

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(size.X), 0)
			op.GeoM.Translate(-float64(size.X)/2, -float64(size.Y)/2)
			op.GeoM.Rotate(-circle.Body().Angle() + math.Pi)
			op.GeoM.Translate(screenWidth/2, screenHeight/2)
			op.GeoM.Translate(vec.X, -vec.Y)
			screen.DrawImage(img, op)
		}
	})
	// cp.DrawSpace(g.space, g.drawer)

	msg := fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
func BeginFunc(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	shape, shape2 := arb.Shapes()

	// arb.Ignore()
	space.AddPostStepCallback(func(space *cp.Space, key interface{}, data interface{}) {
		if shape.Space() != nil {
			space.RemoveShape(shape)
			space.RemoveBody(shape.Body())
			shape.Body().RemoveShape(shape)
		}
		if shape2.Space() != nil {
			space.RemoveShape(shape2)
			space.RemoveBody(shape2.Body())
			shape2.Body().RemoveShape(shape2)
		}
	}, nil, nil)

	return false
}

func main() {

	// chipmunk init

	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{X: 0, Y: -500})
	space.SleepTimeThreshold = 0.5
	space.SetDamping(.99)

	walls := []cp.Vector{
		{X: -hwidth, Y: -hheight}, {X: -hwidth, Y: screenHeight * 10},
		{X: hwidth, Y: -hheight}, {X: hwidth, Y: screenHeight * 10},
		{X: -hwidth, Y: -hheight}, {X: hwidth, Y: -hheight},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 0))
		shape.SetElasticity(0.6)
		shape.SetFriction(0.4)
	}

	space.NewCollisionHandler(assets.Apple, assets.Apple).BeginFunc = BeginFunc
	space.NewCollisionHandler(assets.Grape, assets.Grape).BeginFunc = BeginFunc
	space.NewCollisionHandler(assets.Pineapple, assets.Pineapple).BeginFunc = BeginFunc
	space.NewCollisionHandler(assets.Watermelon, assets.Watermelon).BeginFunc = BeginFunc

	// ebitengine init

	game := &Game{}
	game.space = space
	game.drawer = ebitencp.NewDrawer(screenWidth, screenHeight)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk Physics")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addRandomFruit(space *cp.Space) *cp.Shape {
	j := rand.Intn(4)
	var shape *cp.Shape
	switch j {
	case 0:
		is := assets.Get(assets.Apple)
		shape := addFruit(space, is.Image, assets.Apple)
		shape.SetCollisionType(assets.Apple)
	case 1:
		is := assets.Get(assets.Grape)
		shape := addFruit(space, is.Image, assets.Grape)
		shape.SetCollisionType(assets.Grape)
	case 2:
		is := assets.Get(assets.Pineapple)
		shape := addFruit(space, is.Image, assets.Pineapple)
		shape.SetCollisionType(assets.Pineapple)
	case 3:
		is := assets.Get(assets.Watermelon)
		shape := addFruit(space, is.Image, assets.Watermelon)
		shape.SetCollisionType(assets.Watermelon)
	}
	return shape
}

func addFruit(space *cp.Space, img image.Image, tp int) *cp.Shape {
	b := img.Bounds()
	bb := cp.BB{L: float64(b.Min.X), B: float64(b.Min.Y), R: float64(b.Max.X), T: float64(b.Max.Y)}

	sampleFunc := func(point cp.Vector) float64 {
		x := point.X
		y := point.Y
		rect := img.Bounds()

		if x < float64(rect.Min.X) || x > float64(rect.Max.X) || y < float64(rect.Min.Y) || y > float64(rect.Max.Y) {
			return 0.0
		}
		_, _, _, a := img.At(int(x), int(y)).RGBA()
		return float64(a) / 0xffff
	}

	//lineSet := MarchHard(bb, 100, 100, 0.2, PolyLineCollectSegment, sampleFunc)
	lineSet := cp.MarchSoft(bb, 300, 300, 0.5, cp.PolyLineCollectSegment, sampleFunc)

	line := lineSet.Lines[0].SimplifyCurves(.9)
	offset := cp.Vector{X: float64(b.Max.X-b.Min.X) / 2., Y: float64(b.Max.Y-b.Min.Y) / 2.}
	// center the verts on origin
	for i, l := range line.Verts {
		line.Verts[i] = l.Sub(offset)
	}

	body := space.AddBody(cp.NewBody(10, cp.MomentForPoly(10, len(line.Verts), line.Verts, cp.Vector{}, 1)))
	body.SetPosition(cp.Vector{X: float64(rand.Intn(screenWidth)-hwidth) * 0.99, Y: float64(hheight - rand.Intn(100))})
	body.SetAngle(rand.Float64() * math.Pi * 2)
	body.UserData = tp
	fruit := space.AddShape(cp.NewPolyShape(body, len(line.Verts), line.Verts, cp.NewTransformIdentity(), 0))
	fruit.SetElasticity(.7)
	fruit.SetFriction(0.5)
	return fruit
}
