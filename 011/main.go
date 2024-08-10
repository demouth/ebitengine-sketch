package main

// This is based on "jakecoffman/cp-examples/theojansen".

import (
	"fmt"
	_ "image/png"
	"log"
	"math"

	"github.com/demouth/ebitengine-sketch/011/ebitencp"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/jakecoffman/cp/v2"
)

var (
	motor *cp.SimpleMotor
)

const (
	screenWidth  = 640
	screenHeight = 480
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2
)

type Game struct {
	space    *cp.Space
	drawer   *ebitencp.Drawer
	touchIDs []ebiten.TouchID
}

func (g *Game) Update() error {

	clickLeft, clickRight := false, false
	mouseX, _ := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if mouseX < screenWidth/2 {
			clickLeft = true
		} else {
			clickRight = true
		}
	}

	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || g.leftTouched() || clickLeft {
		motor.Rate = -5
		motor.SetMaxForce(100000)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || g.rightTouched() || clickRight {
		motor.Rate = 5
		motor.SetMaxForce(100000)
	} else {
		motor.SetMaxForce(0)
	}

	g.space.Step(1 / 60.0)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawer.Screen = screen
	cp.DrawSpace(g.space, g.drawer)

	msg := fmt.Sprintf(
		"FPS: %0.2f\n"+
			"Press left or right arrow key to rotate the motor\n"+
			"Press left or right half of the screen to rotate the motor",
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
	space.Iterations = 20
	space.SetGravity(cp.Vector{X: 0, Y: -500})

	var shape *cp.Shape
	var a, b cp.Vector

	walls := []cp.Vector{
		{X: -320, Y: -240}, {X: -320, Y: 240},
		{X: 320, Y: -240}, {X: 320, Y: 240},
		{X: -320, Y: -240}, {X: 320, Y: -240},
	}

	for i := 0; i < len(walls)-1; i += 2 {
		shape = space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 0))
		shape.SetElasticity(0.9)
		shape.SetFriction(0.9)
	}

	offset := 30.0
	chassisMass := 2.0
	a = cp.Vector{X: -offset, Y: 0}
	b = cp.Vector{X: offset, Y: 0}
	chassis := space.AddBody(cp.NewBody(chassisMass, cp.MomentForSegment(chassisMass, a, b, 0)))

	shape = space.AddShape(cp.NewSegment(chassis, a, b, segRadius))
	shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	crankMass := 1.0
	crankRadius := 13.0
	crank := space.AddBody(cp.NewBody(crankMass, cp.MomentForCircle(crankMass, crankRadius, 0, cp.Vector{})))

	shape = space.AddShape(cp.NewCircle(crank, crankRadius, cp.Vector{}))
	shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	space.AddConstraint(cp.NewPivotJoint2(chassis, crank, cp.Vector{}, cp.Vector{}))

	side := 30.0

	const numLegs = 2
	for i := 0; i < numLegs; i++ {
		makeLeg(space, side, offset, chassis, crank, cp.ForAngle(float64(2*i+0)/numLegs*math.Pi).Mult(crankRadius))
		makeLeg(space, side, -offset, chassis, crank, cp.ForAngle(float64(2*i+1)/numLegs*math.Pi).Mult(crankRadius))
	}

	motor = space.AddConstraint(cp.NewSimpleMotor(chassis, crank, 6)).Class.(*cp.SimpleMotor)

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

const segRadius = 3.0

func makeLeg(space *cp.Space, side, offset float64, chassis, crank *cp.Body, anchor cp.Vector) {
	var a, b cp.Vector
	var shape *cp.Shape

	legMass := 1.0

	// make a leg
	a = cp.Vector{}
	b = cp.Vector{X: 0, Y: side}
	upperLeg := space.AddBody(cp.NewBody(legMass, cp.MomentForSegment(legMass, a, b, 0)))
	upperLeg.SetPosition(cp.Vector{X: offset, Y: 0})

	shape = space.AddShape(cp.NewSegment(upperLeg, a, b, segRadius))
	shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	space.AddConstraint(cp.NewPivotJoint2(chassis, upperLeg, cp.Vector{X: offset, Y: 0}, cp.Vector{}))

	// lower leg
	a = cp.Vector{}
	b = cp.Vector{X: 0, Y: -1 * side}
	lowerLeg := space.AddBody(cp.NewBody(legMass, cp.MomentForSegment(legMass, a, b, 0)))
	lowerLeg.SetPosition(cp.Vector{X: offset, Y: -side})

	shape = space.AddShape(cp.NewSegment(lowerLeg, a, b, segRadius))
	shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	shape = space.AddShape(cp.NewCircle(lowerLeg, segRadius*2.0, b))
	shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	shape.SetElasticity(0)
	shape.SetFriction(1)

	space.AddConstraint(cp.NewPinJoint(chassis, lowerLeg, cp.Vector{X: offset, Y: 0}, cp.Vector{}))

	space.AddConstraint(cp.NewGearJoint(upperLeg, lowerLeg, 0, 1))

	var constraint *cp.Constraint
	diag := math.Sqrt(side*side + offset*offset)

	constraint = space.AddConstraint(cp.NewPinJoint(crank, upperLeg, anchor, cp.Vector{X: 0, Y: side}))
	constraint.Class.(*cp.PinJoint).Dist = diag

	constraint = space.AddConstraint(cp.NewPinJoint(crank, lowerLeg, anchor, cp.Vector{}))
	constraint.Class.(*cp.PinJoint).Dist = diag
}

func (g *Game) leftTouched() bool {
	for _, id := range g.touchIDs {
		x, _ := ebiten.TouchPosition(id)
		if x < screenWidth/2 {
			return true
		}
	}
	return false
}

func (g *Game) rightTouched() bool {
	for _, id := range g.touchIDs {
		x, _ := ebiten.TouchPosition(id)
		if x >= screenWidth/2 {
			return true
		}
	}
	return false
}
