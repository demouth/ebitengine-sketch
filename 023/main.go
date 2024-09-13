package main

import (
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/023/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 640
	screenHeight = 640
)

var (
	//go:embed radialblur.kage
	radialblur_kage []byte
)

type Game struct {
	count  int
	canvas *ebiten.Image
	space  *cp.Space
	shader *ebiten.Shader
	ecp    *ebitencp.Drawer
}

func (g *Game) Update() error {
	g.count++
	for i := 0; i < 3; i++ {
		addCircle(g.space, 2.5, rand.Float64()*40-20, -screenHeight/2)
	}

	margin := 10.0
	g.space.EachBody(func(body *cp.Body) {
		remove := false
		if body.Position().Y > screenHeight/2+margin {
			remove = true
		} else if body.Position().X < -screenWidth/2-margin {
			remove = true
		} else if body.Position().X > screenWidth/2+margin {
			remove = true
		}
		if remove {
			g.space.AddPostStepCallback(removeBodyCallback, body, nil)
		}
	})
	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	numCircle := 0
	g.canvas.Fill(color.Black)
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Circle:
			numCircle++
			circle := shape.Class.(*cp.Circle)
			vec := circle.TransformC()
			drawer.DrawCircle(
				g.canvas,
				float32(vec.X+screenWidth/2),
				float32(vec.Y+screenHeight/2),
				float32(circle.Radius()),
				color.NRGBA{0xff, 0xff, 0xff, 0xff},
			)
		}
	})
	// cp.DrawSpace(g.space, g.ecp.WithScreen(g.canvas))

	cx, cy := ebiten.CursorPosition()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Cursor": []float32{float32(cx), float32(cy)},
	}
	op.Images[0] = g.canvas
	screen.DrawRectShader(screenWidth, screenHeight, g.shader, op)
	// screen.DrawImage(g.canvas, nil)

	g.space.StaticBody.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.Segment:
			seg := shape.Class.(*cp.Segment)
			drawer.DrawLine(
				screen,
				float32(screenWidth/2+seg.A().X),
				float32(screenHeight/2+seg.A().Y),
				float32(screenWidth/2+seg.B().X),
				float32(screenHeight/2+seg.B().Y),
				float32(seg.Radius()*2),
				color.RGBA{0x99, 0x99, 0x99, 0xff},
			)
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

func main() {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{X: 0, Y: 400})

	s, err := ebiten.NewShader([]byte(radialblur_kage))
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{}
	game.space = space
	game.canvas = ebiten.NewImage(screenWidth, screenHeight)
	game.ecp = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.ecp.FlipYAxis = true
	game.shader = s

	addWall(space, -150, -120, -20, -80, 5, 0)
	addWall(space, 150, -40, 0, 0, 5, 0)
	addWall(space, -150, 160, -20, 160, 5, 0)
	addWall(space, -150, 100, -150, 160, 5, 0)
	addWall(space, -20, 100, -20, 160, 5, 0)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("shader")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0)
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
