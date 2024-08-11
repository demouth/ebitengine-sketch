package main

// This is based on "jakecoffman/cp-examples/chains".

import (
	"fmt"
	_ "image/png"
	"log"

	"github.com/demouth/ebitengine-sketch/010/ebitencp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2

	CHAIN_COUNT = 8
	LINK_COUNT  = 10
)

type Game struct {
	count  int
	space  *cp.Space
	drawer *ebitencp.Drawer
}

func (g *Game) Update() error {
	g.drawer.HandleMouseEvent(g.space)

	g.space.Step(1 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	g.drawer.Screen = screen
	cp.DrawSpace(g.space, g.drawer)

	msg := fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	// chipmunk init

	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{X: 0, Y: -100})
	space.SleepTimeThreshold = 0.5

	walls := []cp.Vector{
		{X: -320, Y: -240}, {X: -320, Y: 240},
		{X: 320, Y: -240}, {X: 320, Y: 240},
		{X: -320, Y: -240}, {X: 320, Y: -240},
		{X: -320, Y: 240}, {X: 320, Y: 240},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 0))
		shape.SetElasticity(1)
		shape.SetFriction(1)
	}

	mass := 1.0
	width := 20.0
	height := 30.0

	spacing := width * 0.3

	var i, j float64
	for i = 0; i < CHAIN_COUNT; i++ {
		var prev *cp.Body

		for j = 0; j < LINK_COUNT; j++ {
			pos := cp.Vector{X: 40 * (i - (CHAIN_COUNT-1)/2.0), Y: 240 - (j+0.5)*height - (j+1)*spacing}

			body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
			body.SetPosition(pos)

			shape := space.AddShape(cp.NewSegment(body, cp.Vector{X: 0, Y: (height - width) / 2}, cp.Vector{X: 0, Y: (width - height) / 2}, width/2))
			shape.SetFriction(0.8)

			breakingForce := 80000.0

			var constraint *cp.Constraint
			if prev == nil {
				constraint = space.AddConstraint(cp.NewSlideJoint(body, space.StaticBody, cp.Vector{X: 0, Y: height / 2}, cp.Vector{X: pos.X, Y: 240}, 0, spacing))
			} else {
				constraint = space.AddConstraint(cp.NewSlideJoint(body, prev, cp.Vector{X: 0, Y: height / 2}, cp.Vector{X: 0, Y: -height / 2}, 0, spacing))
			}

			constraint.SetMaxForce(breakingForce)
			constraint.PostSolve = BreakableJointPostSolve
			constraint.SetCollideBodies(false)

			prev = body
		}
	}

	radius := 15.0
	body := space.AddBody(cp.NewBody(10, cp.MomentForCircle(10, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: 0, Y: -240 + radius + 5})
	body.SetVelocity(0, 300)

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.9)

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

func BreakableJointPostStepRemove(space *cp.Space, joint interface{}, _ interface{}) {
	space.RemoveConstraint(joint.(*cp.Constraint))
}

func BreakableJointPostSolve(joint *cp.Constraint, space *cp.Space) {
	dt := space.TimeStep()

	// Convert the impulse to a force by dividing it by the timestep.
	force := joint.Class.GetImpulse() / dt
	maxForce := joint.MaxForce()

	// If the force is almost as big as the joint's max force, break it.
	if force > 0.9*maxForce {
		space.AddPostStepCallback(BreakableJointPostStepRemove, joint, nil)
	}
}
