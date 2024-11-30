package main

import (
	"image/color"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/033/colorpallet"
	"github.com/demouth/ebitengine-sketch/033/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth   = 960
	screenHeight  = 960
	hScreenWidth  = screenWidth / 2
	hScreenHeight = screenHeight / 2
)

type Game struct {
	space   *cp.Space
	drawer  *ebitencp.Drawer
	counter uint64
}

func (g *Game) Update() error {
	// Handling dragging
	g.drawer.HandleMouseEvent(g.space)
	g.counter++

	if g.counter == 140 {
		walls := []cp.Vector{
			{X: 0, Y: 0}, {X: 0, Y: screenHeight},
			{X: screenWidth, Y: 0}, {X: screenWidth, Y: screenHeight},
			{X: 0, Y: screenHeight}, {X: screenWidth, Y: screenHeight},
		}
		for i := 0; i < len(walls)-1; i += 2 {
			shape := g.space.AddShape(cp.NewSegment(g.space.StaticBody, walls[i], walls[i+1], 0))
			shape.SetElasticity(0.5)
			shape.SetFriction(0.5)
		}
	} else if g.counter == 430 {
		g.space.EachShape(func(shape *cp.Shape) {
			switch shape.Class.(type) {
			case *cp.Segment:
				g.space.AddPostStepCallback(removeShapeCallback, shape, nil)
			}
		})
	} else if g.counter > 610 {
		g.counter = 0
		g.initCP()
	}

	g.space.EachBody(func(body *cp.Body) {
		if body.Position().Y > screenHeight*2 {
			g.space.AddPostStepCallback(removeBodyCallback, body, nil)
		}
	})

	g.space.Step(1 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawer.Screen = screen
	g.space.EachShape(func(shape *cp.Shape) {
		body := shape.Body()
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)
			c := shape.UserData.(color.RGBA)
			drawer.DrawCircle(
				screen,
				float32(body.Position().X),
				float32(body.Position().Y),
				float32(circle.Radius()),
				c,
			)
		case *cp.PolyShape:
			poly := shape.Class.(*cp.PolyShape)
			c := shape.UserData.(color.RGBA)
			path := vector.Path{}
			for i, l := 0, poly.Count(); i < l; i++ {
				v := poly.Vert(i)
				v = body.LocalToWorld(v)
				if i == 0 {
					path.MoveTo(float32(v.X), float32(v.Y))
				} else {
					path.LineTo(float32(v.X), float32(v.Y))
				}
			}
			path.Close()
			drawer.DrawFill(
				screen,
				path,
				c,
			)
		}
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) initCP() {
	space := g.space

	boxes := NewBoxes(0, 0, screenWidth, screenHeight)
	const m = 0.999
	for _, box := range boxes {
		shiftY := -float64(screenHeight)
		if rand.Float64() < 0.95 {
			addBox(
				space, box.w*m,
				box.h*m,
				box.x+box.w/2,
				(box.y+box.h/2)*2+shiftY*2,
			)
		} else {
			addBall(
				space,
				box.x+box.w/2,
				(box.y+box.h/2)*2+shiftY*2,
				box.w/2*m,
			)
		}
	}
}

func main() {
	// Initialising Chipmunk
	space := cp.NewSpace()
	space.SleepTimeThreshold = 0.5
	space.SetGravity(cp.Vector{X: 0, Y: 100})

	// Initialising Ebitengine/v2
	game := &Game{}
	game.space = space
	game.drawer = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.drawer.GeoM.Translate(-hScreenWidth, -hScreenHeight)
	game.drawer.FlipYAxis = true
	game.initCP()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ebitengine-sketch 033")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

var pallet = colorpallet.NewColors(2)

func addBall(space *cp.Space, x, y, radius float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})
	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0.0)
	shape.SetFriction(1.0)
	shape.UserData = pallet.Random()
}

func addBox(space *cp.Space, w, h float64, x, y float64) {
	mass := w * h / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, w, h)))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewBox(body, w, h, 0))
	shape.SetElasticity(0.0)
	shape.SetFriction(1.0)
	shape.UserData = pallet.Random()
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
func removeShapeCallback(space *cp.Space, key interface{}, data interface{}) {
	var s *cp.Shape
	var ok bool
	if s, ok = key.(*cp.Shape); !ok {
		return
	}
	s.Body().EachConstraint(func(c *cp.Constraint) {
		space.RemoveConstraint(c)
	})
	space.RemoveBody(s.Body())
	space.RemoveShape(s)
}

type Box struct {
	w, h  float64
	x, y  float64
	count int
}
type Boxes []*Box

func (b *Box) SplitRandom() []*Box {
	return b.split()
}
func (b *Box) split() []*Box {
	splited := make([]*Box, 4)
	splited[0] = &Box{w: b.w / 2, h: b.h / 2, x: b.x, y: b.y, count: b.count + 1}
	splited[1] = &Box{w: b.w / 2, h: b.h / 2, x: b.x, y: b.y + b.h/2, count: b.count + 1}
	splited[2] = &Box{w: b.w / 2, h: b.h / 2, x: b.x + b.w/2, y: b.y, count: b.count + 1}
	splited[3] = &Box{w: b.w / 2, h: b.h / 2, x: b.x + b.w/2, y: b.y + b.h/2, count: b.count + 1}

	ret := make([]*Box, 0, len(splited))

	for _, box := range splited {
		if b.count < 5 && rand.Float64() < 0.6 {
			ret = append(ret, box.split()...)
		} else {
			ret = append(ret, box)
		}
	}
	return ret
}

func NewBoxes(x, y, w, h float64) Boxes {
	box := &Box{w: w, h: h, x: x, y: y, count: 0}
	return box.SplitRandom()
}
