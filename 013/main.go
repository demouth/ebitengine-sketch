package main

// This is based on "jakecoffman/cp-examples/march".

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/013/assets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jakecoffman/cp/v2"
)

var (
	bgImage       = ebiten.NewImage(100, 50)
	whiteSubImage = ebiten.NewImage(3, 3)
	score         = 0
	hiscore       = 0
)

const (
	screenWidth     = 480
	screenHeight    = 800
	hwidth          = screenWidth / 2
	hheight         = screenHeight / 2
	containerHeight = 600
)

type Game struct {
	count  int
	space  *cp.Space
	drawer *ebitencp.Drawer
}

func (g *Game) Update() error {
	g.count++
	if g.count%40 == 0 {
		addRandomFruit(g.space)
	}
	resetFlag := false

	g.space.EachBody(func(body *cp.Body) {
		if body.Position().Y > -hheight+containerHeight {
			resetFlag = true
			return
		}
	})
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		resetFlag = true
	}
	if resetFlag {
		g.space.EachShape(func(shape *cp.Shape) {
			if shape.Body().UserData != nil {
				g.space.AddPostStepCallback(removeShapeCallback, shape, nil)
			}
			hiscore = int(math.Max(float64(score), float64(hiscore)))
			score = 0
		})
	}

	g.drawer.HandleMouseEvent(g.space)

	g.space.Step(1 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{230, 190, 200, 255})
	screen.DrawImage(bgImage, nil)
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.PolyShape:
			circle := shape.Class.(*cp.PolyShape)
			vec := circle.Body().Position()

			imgSet := assets.Get(circle.Body().UserData.(assets.Kind))
			img := imgSet.EbitenImage
			size := img.Bounds().Size()

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(size.X), 0)
			op.GeoM.Translate(-float64(size.X)/2, -float64(size.Y)/2)
			op.GeoM.Rotate(-circle.Body().Angle() + math.Pi)
			op.GeoM.Scale(imgSet.Scale, imgSet.Scale)
			op.GeoM.Translate(screenWidth/2, screenHeight/2)
			op.GeoM.Translate(vec.X, -vec.Y)
			screen.DrawImage(img, op)
		}
	})
	// cp.DrawSpace(g.space, g.drawer.WithScreen(screen))
	g.drawUI(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f\nScore: %d\nHiScore: %d",
		ebiten.ActualFPS(),
		score,
		hiscore,
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) drawUI(screen *ebiten.Image) {

	var path vector.Path
	path.MoveTo(0, screenHeight-containerHeight)
	path.LineTo(screenWidth, screenHeight-containerHeight)
	path.Close()
	var vs []ebiten.Vertex
	var is []uint16
	sop := &vector.StrokeOptions{}
	sop.Width = 3
	sop.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0xff / float32(0xff)
		vs[i].ColorG = 0xcc / float32(0xff)
		vs[i].ColorB = 0xcc / float32(0xff)
		vs[i].ColorA = 1
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleFillAll

	screen.DrawTriangles(vs, is, whiteSubImage, op)
}

func init() {
	whiteSubImage.Fill(color.White)
}

func main() {

	// chipmunk init

	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{X: 0, Y: -500})
	space.SleepTimeThreshold = 0.5
	space.SetDamping(1)

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

	assets.ForEach(func(i assets.Kind, is assets.ImageSet) {
		ct := cp.CollisionType(i)
		space.NewCollisionHandler(ct, ct).BeginFunc = BeginFunc
	})

	// ebitengine init

	bgImage.Fill(color.Black)

	game := &Game{}
	game.space = space
	game.drawer = ebitencp.NewDrawer(screenWidth, screenHeight)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk Physics")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addRandomFruit(space *cp.Space) {
	j := assets.Tomato
	pos := cp.Vector{X: float64(rand.Intn(screenWidth)-hwidth) * 0.01, Y: float64(-hheight + containerHeight - 10)}
	addFruit(space, j, pos, rand.Float64()*math.Pi*2)
}

func addFruit(space *cp.Space, k assets.Kind, position cp.Vector, angle float64) {
	if !assets.Exists(k) {
		return
	}
	imgSet := assets.Get(k)

	body := space.AddBody(cp.NewBody(0, cp.MomentForPoly(10, len(imgSet.Vectors), imgSet.Vectors, cp.Vector{}, 1)))
	body.SetPosition(position)
	body.SetAngle(angle)
	body.UserData = k
	fruit := space.AddShape(cp.NewPolyShape(body, len(imgSet.Vectors), imgSet.Vectors, cp.NewTransformIdentity(), 0))
	body.SetMass(fruit.Area() * 0.001)
	fruit.SetElasticity(0.2)
	fruit.SetFriction(0.9)
	fruit.SetCollisionType(cp.CollisionType(k))
}

type addShapeOptions struct {
	kind  assets.Kind
	pos   cp.Vector
	angle float64
}

func addShapeCallback(space *cp.Space, key interface{}, data interface{}) {
	var opt addShapeOptions
	if i, ok := data.(addShapeOptions); ok {
		opt = i
	} else {
		return
	}
	addFruit(space, opt.kind, opt.pos, opt.angle)
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
func BeginFunc(arb *cp.Arbiter, space *cp.Space, data interface{}) bool {
	shape, shape2 := arb.Shapes()

	var k assets.Kind
	if ud, ok := shape.Body().UserData.(assets.Kind); ok {
		k = ud
	} else {
		return false
	}

	space.AddPostStepCallback(removeShapeCallback, shape, nil)
	space.AddPostStepCallback(removeShapeCallback, shape2, nil)

	score += k.Score()

	if hasNext, kk := k.Next(); hasNext {
		k = kk
	} else {
		return false
	}
	sp := shape.Body().Position().Clone()
	sp.Sub(shape2.Body().Position()).Mult(0.5).Add(shape2.Body().Position())
	a := (shape.Body().Angle() + shape2.Body().Angle()) / 2
	addShapeOptions := addShapeOptions{
		kind:  k,
		pos:   sp,
		angle: a,
	}
	space.AddPostStepCallback(addShapeCallback, k, addShapeOptions)
	return false
}
