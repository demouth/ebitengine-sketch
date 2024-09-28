package main

import (
	_ "embed"
	"fmt"
	_ "image/png"

	"github.com/demouth/ebitencp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 1000
	screenHeight = 1000
)

type Game struct {
	space *cp.Space
	ecp   *ebitencp.Drawer
}

func (g *Game) Update() error {
	g.ecp.HandleMouseEvent(g.space)
	g.space.Step(1.0 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	cp.DrawSpace(g.space, g.ecp.WithScreen(screen))

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{X: 0, Y: 100})

	{
		const step = 50.0
		radius := step * 0.6
		circles := [][]*cp.Body{}
		circlesForJoints := []*cp.Body{}
		startX, startY := 100.0, 100.0
		for y := 0; y < 10; y++ {
			circles = append(circles, []*cp.Body{})
			for x := 0; x < 5; x++ {
				circle, _ := addCircle(space, radius, startX+step*float64(x), startY+step*float64(y), 0.0)
				circles[y] = append(circles[y], circle)
				circlesForJoints = append(circlesForJoints, circle)
			}
		}

		l := len(circlesForJoints)
		for i := 0; i < l; i++ {
			for j := i + 1; j < l; j++ {
				if circlesForJoints[i].Position().Distance(circlesForJoints[j].Position()) <= step*1.5 {
					c := space.AddConstraint(
						cp.NewDampedSpring(
							circlesForJoints[i], circlesForJoints[j],
							cp.Vector{X: 0, Y: 0}, cp.Vector{X: 0, Y: 0},
							circlesForJoints[i].Position().Distance(circlesForJoints[j].Position()),
							200.1, 3.2,
						),
					)
					c.SetCollideBodies(false)
				}
			}
		}
	}

	addWall(space, 0, 0, 0, screenHeight, 1, 0.1)
	addWall(space, screenWidth, 0, screenWidth, screenHeight, 1, 0.1)
	addWall(space, 0, 0, screenWidth, 0, 1, 0.1)
	addWall(space, 0, screenHeight, screenWidth, screenHeight, 1, 0.1)

	g := &Game{}
	g.space = space

	g.ecp = ebitencp.NewDrawer(0, 0)
	g.ecp.FlipYAxis = true
	g.ecp.OptStroke.AntiAlias = true

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("tofu")
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
func addWall(space *cp.Space, x1, y1, x2, y2, radius, elasticity float64) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	shape := space.AddShape(cp.NewSegment(space.StaticBody, pos1, pos2, radius))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.99)
	shape.UserData = "wall"
}
