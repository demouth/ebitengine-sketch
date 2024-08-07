package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 320
	screenHeight = 240
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
)

type Game struct {
	count int

	space *cp.Space
}

func (g *Game) Update() error {
	g.count++
	if g.count%20 == 0 {
		addCircle(g.space, 10, rand.Float64()*10-5, -screenHeight/2+10)
	}
	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)
			vec := circle.TransformC()

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
			op.GeoM.Rotate(circle.Body().Angle())
			op.GeoM.Translate(screenWidth/2, screenHeight/2)
			op.GeoM.Translate(vec.X, vec.Y)
			i := (g.count / 5) % frameCount
			sx, sy := frameOX+i*frameWidth, frameOY
			screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), op)
		}
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}

	space := cp.NewSpace()
	space.SetGravity(cp.Vector{X: 0, Y: 200})
	game.space = space

	sides := []cp.Vector{
		{X: -hwidth, Y: -hheight}, {X: -hwidth, Y: hheight},
		{X: hwidth, Y: -hheight}, {X: hwidth, Y: hheight},
		{X: -hwidth, Y: -hheight}, {X: hwidth, Y: -hheight},
		{X: -hwidth, Y: hheight}, {X: hwidth, Y: hheight},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(0.9)
		seg.SetFriction(0.9)
	}

	for i := 0; i < 1; i++ {
		addCircle(space, 10, rand.Float64()*2-1, rand.Float64()*2-1)
	}

	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Animation (Ebitengine Demo)")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0.96)
	shape.SetFriction(0.96)
}
