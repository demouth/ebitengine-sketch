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
	"github.com/demouth/ebitengine-sketch/013/ui"
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
	containerHeight = 600
	paddingBottom   = 100
)

type Game struct {
	count   int
	space   *cp.Space
	drawer  *ebitencp.Drawer
	next    next
	buttons ui.Components

	debug bool
}

type next struct {
	kind  assets.Kind
	x     float64
	y     float64
	angle float64
}

func (g *Game) Update() error {
	g.count++
	g.space.EachBody(func(body *cp.Body) {
		if body.Position().Y < screenHeight-containerHeight {
			g.space.EachShape(func(shape *cp.Shape) {
				if shape.Body().UserData != nil {
					g.space.AddPostStepCallback(removeShapeCallback, shape, nil)
				}
				hiscore = int(math.Max(float64(score), float64(hiscore)))
				score = 0
			})
			return
		}
	})
	g.next.angle += 0.01
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.drop()
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.moveRight()
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.moveLeft()
	}
	g.buttons.Update()
	g.drawer.HandleMouseEvent(g.space)
	g.space.Step(1 / 60.0)
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.drawFruit(screen, g.next.kind, g.next.x, g.next.y-paddingBottom, g.next.angle)
	g.space.EachShape(func(shape *cp.Shape) {
		switch shape.Class.(type) {
		case *cp.PolyShape:
			circle := shape.Class.(*cp.PolyShape)
			vec := circle.Body().Position()
			g.drawFruit(screen, circle.Body().UserData.(assets.Kind), vec.X, vec.Y-paddingBottom, circle.Body().Angle())
		}
	})
	if g.debug {
		cp.DrawSpace(g.space, g.drawer.WithScreen(screen))
	}
	g.buttons.Draw(screen)
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

func (g *Game) moveRight() {
	if g.next.x < screenWidth-50 {
		g.next.x += 4
	}
}
func (g *Game) moveLeft() {
	if g.next.x > 40 {
		g.next.x -= 4
	}
}
func (g *Game) drop() {
	if g.count > 40 {
		k := g.next.kind
		addShapeOptions := addShapeOptions{
			kind:  g.next.kind,
			pos:   cp.Vector{X: g.next.x, Y: g.next.y},
			angle: g.next.angle,
		}
		g.space.AddPostStepCallback(addShapeCallback, k, addShapeOptions)
		g.count = 0
		g.next.kind = assets.Kind(rand.Intn(2) + int(assets.Min))
	}
}
func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
	screen.DrawImage(bgImage, nil)

	var path vector.Path

	path = vector.Path{}
	path.MoveTo(0, 0)
	path.LineTo(0, screenHeight-paddingBottom)
	path.LineTo(screenWidth, screenHeight-paddingBottom)
	path.LineTo(screenWidth, 0)
	g.drawFill(screen, path, color.NRGBA{0xff, 0xcc, 0x99, 0xff})
	g.drawLine(screen, path, color.NRGBA{0xaa, 0x66, 0x33, 0xff}, 5)

	path = vector.Path{}
	path.MoveTo(0, screenHeight-containerHeight-paddingBottom)
	path.LineTo(screenWidth, screenHeight-containerHeight-paddingBottom)
	g.drawLine(screen, path, color.NRGBA{0xdd, 0xaa, 0x99, 0xff}, 3)

	path = vector.Path{}
	path.MoveTo(0, 0)
	path.LineTo(0, 50)
	path.LineTo(100, 50)
	path.LineTo(100, 0)
	g.drawFill(screen, path, color.NRGBA{0x00, 0x00, 0x00, 0xff})
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

func (g *Game) drawFill(screen *ebiten.Image, path vector.Path, c color.NRGBA) {
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
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

func init() {
	whiteSubImage.Fill(color.White)
}

func main() {

	// chipmunk init

	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{X: 0, Y: 500})
	space.SleepTimeThreshold = 0.5
	space.SetDamping(1)

	walls := []cp.Vector{
		{X: 0, Y: 0}, {X: 0, Y: screenHeight},
		{X: screenWidth, Y: 0}, {X: screenWidth, Y: screenHeight},
		{X: 0, Y: screenHeight}, {X: screenWidth, Y: screenHeight},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 1))
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
	game.drawer.FlipYAxis = true
	game.drawer.Camera.Offset = cp.Vector{X: screenWidth / 2, Y: screenHeight/2 + paddingBottom}
	game.next = next{kind: assets.Tomato, x: screenWidth / 2, y: screenHeight - containerHeight + 10, angle: 0}
	game.buttons = ui.Components{
		&ui.Button{X: 20, Y: screenHeight - 70, Width: 60, Height: 40, FontSize: 8, Text: "debug", OnMouseDown: func() {
			game.debug = !game.debug
		}},
		&ui.Button{X: 120, Y: screenHeight - 80, Width: 80, Height: 60, FontSize: 14, Text: "<-", OnMouseDownHold: func() {
			game.moveLeft()
		}},
		&ui.Button{X: 220, Y: screenHeight - 80, Width: 100, Height: 60, FontSize: 14, Text: "Drop", OnMouseDownHold: func() {
			game.drop()
		}},
		&ui.Button{X: 340, Y: screenHeight - 80, Width: 80, Height: 60, FontSize: 14, Text: "->", OnMouseDownHold: func() {
			game.moveRight()
		}},
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk Physics")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addRandomFruit(space *cp.Space) {
	j := assets.Tomato
	pos := cp.Vector{X: screenWidth / 2, Y: screenHeight - containerHeight + 10}
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
func (g *Game) drawFruit(screen *ebiten.Image, kind assets.Kind, x, y, angle float64) {
	imgSet := assets.Get(kind)
	img := imgSet.EbitenImage
	size := img.Bounds().Size()

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(-float64(size.X)/2, -float64(size.Y)/2)
	op.GeoM.Rotate(angle)
	op.GeoM.Scale(imgSet.Scale, imgSet.Scale)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
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
