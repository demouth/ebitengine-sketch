package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 640
	screenHeight = 640
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

var (
	runnerImage   *ebiten.Image
	whiteSubImage = ebiten.NewImage(3, 3)
)

type Game struct {
	count int

	space *cp.Space
}

func (g *Game) Update() error {
	g.count++
	for i := 0; i < 5; i++ {
		addCircle(g.space, 10, rand.Float64()*10-5, -screenHeight/2+10)
	}

	g.space.EachBody(func(body *cp.Body) {
		if body.Position().Y > hheight {
			g.space.AddPostStepCallback(removeBodyCallback, body, nil)
			return
		}
	})

	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.space.StaticBody.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Segment:
			seg := shape.Class.(*cp.Segment)
			path := vector.Path{}
			path.MoveTo(float32(hwidth+seg.A().X), float32(hheight+seg.A().Y))
			path.LineTo(float32(hwidth+seg.B().X), float32(hheight+seg.B().Y))
			g.drawLine(screen, path, color.NRGBA{0xff, 0xff, 0xff, 0xff}, 5)
		}
	})

	numCircle := 0
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Circle:
			numCircle++
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

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f\nNumCircle: %v",
		ebiten.ActualFPS(),
		numCircle,
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

	space := cp.NewSpace()
	space.SetGravity(cp.Vector{X: 0, Y: 200})
	game.space = space

	for i := 0; i < 1; i++ {
		addCircle(space, 10, rand.Float64()*2-1, rand.Float64()*2-1)
	}
	addWall(space, -100, 100, +100, 100, 5)

	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0.1)
	shape.SetFriction(0.96)
}
func addWall(space *cp.Space, x1, y1, x2, y2, radius float64) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	shape := space.AddShape(cp.NewSegment(space.StaticBody, pos1, pos2, radius))
	shape.SetElasticity(0.1)
	shape.SetFriction(0.5)
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
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}
func removeShapeCallback(space *cp.Space, key interface{}, data interface{}) {
	var s *cp.Shape
	var ok bool
	if s, ok = key.(*cp.Shape); !ok {
		return
	}
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
