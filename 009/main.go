package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

var (
	runnerImage *ebiten.Image

	simpleTerrainVerts = []cp.Vector{
		{X: 350.00, Y: 425.07}, {X: 336.00, Y: 436.55}, {X: 272.00, Y: 435.39}, {X: 258.00, Y: 427.63}, {X: 225.28, Y: 420.00}, {X: 202.82, Y: 396.00},
		{X: 191.81, Y: 388.00}, {X: 189.00, Y: 381.89}, {X: 173.00, Y: 380.39}, {X: 162.59, Y: 368.00}, {X: 150.47, Y: 319.00}, {X: 128.00, Y: 311.55},
		{X: 119.14, Y: 286.00}, {X: 126.84, Y: 263.00}, {X: 120.56, Y: 227.00}, {X: 141.14, Y: 178.00}, {X: 137.52, Y: 162.00}, {X: 146.51, Y: 142.00},
		{X: 156.23, Y: 136.00}, {X: 158.00, Y: 118.27}, {X: 170.00, Y: 100.77}, {X: 208.43, Y: 84.00}, {X: 224.00, Y: 69.65}, {X: 249.30, Y: 68.00},
		{X: 257.00, Y: 54.77}, {X: 363.00, Y: 45.94}, {X: 374.15, Y: 54.00}, {X: 386.00, Y: 69.60}, {X: 413.00, Y: 70.73}, {X: 456.00, Y: 84.89},
		{X: 468.09, Y: 99.00}, {X: 467.09, Y: 123.00}, {X: 464.92, Y: 135.00}, {X: 469.00, Y: 141.03}, {X: 497.00, Y: 148.67}, {X: 513.85, Y: 180.00},
		{X: 509.56, Y: 223.00}, {X: 523.51, Y: 247.00}, {X: 523.00, Y: 277.00}, {X: 497.79, Y: 311.00}, {X: 478.67, Y: 348.00}, {X: 467.90, Y: 360.00},
		{X: 456.76, Y: 382.00}, {X: 432.95, Y: 389.00}, {X: 417.00, Y: 411.32}, {X: 373.00, Y: 433.19}, {X: 361.00, Y: 430.02}, {X: 350.00, Y: 425.07},
	}
)

// copy from https://github.com/jakecoffman/cp-examples/blob/master/bench/bench.go
func simpleTerrain() *cp.Space {
	space := cp.NewSpace()
	space.Iterations = 10
	space.SleepTimeThreshold = 0.5
	space.SetCollisionSlop(0.5)

	offset := cp.Vector{X: -320, Y: -240}
	for i := 0; i < len(simpleTerrainVerts)-1; i++ {
		a := simpleTerrainVerts[i]
		b := simpleTerrainVerts[i+1]
		shape := cp.NewSegment(space.StaticBody, a.Add(offset), b.Add(offset), 0)
		space.AddShape(shape)
		shape.SetFriction(0.8)
	}

	return space
}

type Game struct {
	count int

	numShapes int

	space *cp.Space

	ebitencp *Ebitencp
}

func (g *Game) Update() error {
	g.count++
	if g.numShapes < 150 && g.count%15 == 0 {

		g.numShapes++
		addCircle(g.space, rand.Float64()*15+10, rand.Float64()*10, rand.Float64()*2-150)
		g.numShapes++
		addBox(g.space, rand.Float64()*15+10, rand.Float64()*15+10, rand.Float64()*10, rand.Float64()*2-150)
	}
	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// g.space.EachShape(func(shape *cp.Shape) {
	// 	switch shape.Class.(type) {
	// 	case *cp.Circle:
	// 		circle := shape.Class.(*cp.Circle)
	// 		vec := circle.TransformC()

	// 		op := &ebiten.DrawImageOptions{}
	// 		op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	// 		op.GeoM.Rotate(circle.Body().Angle())
	// 		op.GeoM.Scale(circle.Radius()/10, circle.Radius()/10)
	// 		op.GeoM.Translate(screenWidth/2, screenHeight/2)
	// 		op.GeoM.Translate(vec.X, vec.Y)
	// 		i := (g.count / 5) % frameCount
	// 		sx, sy := frameOX+i*frameWidth, frameOY
	// 		screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), op)
	// 	}
	// })

	g.ebitencp.Draw(screen, g.space)
	msg := fmt.Sprintf(
		"FPS: %0.2f\nNum Circles: %d",
		ebiten.ActualFPS(),
		g.numShapes,
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}

	game.ebitencp = &Ebitencp{}

	space := simpleTerrain()
	space.SetGravity(cp.Vector{X: 0, Y: 200})
	game.space = space

	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk Physics")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0.5)
	shape.SetFriction(0.5)
}

func addBox(space *cp.Space, w, h float64, x, y float64) {
	mass := w * h / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, w, h)))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewBox(body, w, h, 0))
	shape.SetElasticity(0.5)
	shape.SetFriction(0.5)
}
