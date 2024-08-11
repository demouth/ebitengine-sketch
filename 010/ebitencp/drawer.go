package ebitencp

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp/v2"
)

const DrawPointLineScale = 1.0

type Drawer struct {
	whiteImage *ebiten.Image
	Screen     *ebiten.Image

	ScreenWidth  int
	ScreenHeight int
	AntiAlias    bool
	StrokeWidth  float32

	handler mouseEventHandler
}

func NewDrawer(screenWidth, screenHeight int) *Drawer {
	whiteImage := ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)
	return &Drawer{
		whiteImage:   whiteImage,
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		AntiAlias:    true,
		StrokeWidth:  1,
	}
}

func (d *Drawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {

	path := &vector.Path{}
	path.Arc(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2,
		float32(radius),
		0, 2*math.Pi, vector.Clockwise)
	d.drawFill(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)

	path.MoveTo(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2)
	path.LineTo(
		float32(pos.X+math.Cos(angle)*radius)+float32(d.ScreenWidth)/2,
		-float32(pos.Y+math.Sin(angle)*radius)+float32(d.ScreenHeight)/2)
	path.Close()

	d.drawOutline(d.Screen, *path, outline.R, outline.G, outline.B, outline.A)
}

func (d *Drawer) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {

	var path *vector.Path = &vector.Path{}
	path.MoveTo(
		float32(a.X)+float32(d.ScreenWidth)/2,
		-float32(a.Y)+float32(d.ScreenHeight)/2)
	path.LineTo(
		float32(b.X)+float32(d.ScreenWidth)/2,
		-float32(b.Y)+float32(d.ScreenHeight)/2)
	path.Close()
	d.drawFill(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
	d.drawOutline(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
}

func (d *Drawer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	var path vector.Path = vector.Path{}
	t1 := -float32(math.Atan2(b.Y-a.Y, b.X-a.X)) + math.Pi/2
	t2 := t1 + math.Pi
	path.Arc(
		float32(a.X)+float32(d.ScreenWidth)/2,
		-float32(a.Y)+float32(d.ScreenHeight)/2,
		float32(radius),
		t1, t1+math.Pi, vector.Clockwise)
	path.Arc(
		float32(b.X)+float32(d.ScreenWidth)/2,
		-float32(b.Y)+float32(d.ScreenHeight)/2,
		float32(radius),
		t2, t2+math.Pi, vector.Clockwise)
	path.Close()
	d.drawFill(d.Screen, path, fill.R, fill.G, fill.B, fill.A)
	d.drawOutline(d.Screen, path, outline.R, outline.G, outline.B, outline.A)
}

func (d *Drawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data interface{}) {
	type ExtrudeVerts struct {
		offset, n cp.Vector
	}
	extrude := make([]ExtrudeVerts, count)

	for i := 0; i < count; i++ {
		v0 := verts[(i-1+count)%count]
		v1 := verts[i]
		v2 := verts[(i+1)%count]

		n1 := v1.Sub(v0).ReversePerp().Normalize()
		n2 := v2.Sub(v1).ReversePerp().Normalize()

		offset := n1.Add(n2).Mult(1.0 / (n1.Dot(n2) + 1.0))
		extrude[i] = ExtrudeVerts{offset, n2}
	}

	var path *vector.Path = &vector.Path{}

	inset := -math.Max(0, 1.0/DrawPointLineScale-radius)
	for i := 0; i < count-2; i++ {
		v0 := verts[0].Add(extrude[0].offset.Mult(inset))
		v1 := verts[i+1].Add(extrude[i+1].offset.Mult(inset))
		v2 := verts[i+2].Add(extrude[i+2].offset.Mult(inset))

		path.MoveTo(
			float32(v0.X)+float32(d.ScreenWidth)/2,
			-float32(v0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v1.X)+float32(d.ScreenWidth)/2,
			-float32(v1.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v2.X)+float32(d.ScreenWidth)/2,
			-float32(v2.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v0.X)+float32(d.ScreenWidth)/2,
			-float32(v0.Y)+float32(d.ScreenHeight)/2)
	}

	outset := 1.0/DrawPointLineScale + radius - inset
	j := count - 1
	for i := 0; i < count; {
		vA := verts[i]
		vB := verts[j]

		nA := extrude[i].n
		nB := extrude[j].n

		offsetA := extrude[i].offset
		offsetB := extrude[j].offset

		innerA := vA.Add(offsetA.Mult(inset))
		innerB := vB.Add(offsetB.Mult(inset))

		inner0 := innerA
		inner1 := innerB
		outer0 := innerA.Add(nB.Mult(outset))
		outer1 := innerB.Add(nB.Mult(outset))
		outer2 := innerA.Add(offsetA.Mult(outset))
		outer3 := innerA.Add(nA.Mult(outset))

		path.MoveTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner1.X)+float32(d.ScreenWidth)/2,
			-float32(inner1.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer1.X)+float32(d.ScreenWidth)/2,
			-float32(outer1.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer0.X)+float32(d.ScreenWidth)/2,
			-float32(outer0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer1.X)+float32(d.ScreenWidth)/2,
			-float32(outer1.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer0.X)+float32(d.ScreenWidth)/2,
			-float32(outer0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer2.X)+float32(d.ScreenWidth)/2,
			-float32(outer2.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer2.X)+float32(d.ScreenWidth)/2,
			-float32(outer2.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer3.X)+float32(d.ScreenWidth)/2,
			-float32(outer3.Y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.X)+float32(d.ScreenWidth)/2,
			-float32(inner0.Y)+float32(d.ScreenHeight)/2)

		j = i
		i++
	}

	d.drawFill(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
	d.drawOutline(d.Screen, *path, outline.R, outline.G, outline.B, outline.A)
}
func (d *Drawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {

	var path *vector.Path = &vector.Path{}
	path.Arc(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2,
		float32(2),
		0, 2*math.Pi, vector.Clockwise)
	path.Close()

	d.drawFill(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
}

func (d *Drawer) Flags() uint {
	return 0
}

func (d *Drawer) OutlineColor() cp.FColor {
	return cp.FColor{R: 200.0 / 255.0, G: 210.0 / 255.0, B: 230.0 / 255.0, A: 1}
}

func (d *Drawer) ShapeColor(shape *cp.Shape, data interface{}) cp.FColor {
	body := shape.Body()
	if body.IsSleeping() {
		return cp.FColor{R: .2, G: .2, B: .2, A: 1}
	}

	if body.IdleTime() > shape.Space().SleepTimeThreshold {
		return cp.FColor{R: .66, G: .66, B: .66, A: 1}
	}
	return cp.FColor{R: 0.7, G: 0.3, B: 0.6, A: 0.5}
}

func (d *Drawer) ConstraintColor() cp.FColor {
	return cp.FColor{R: 0, G: 0.75, B: 0, A: 1}
}

func (d *Drawer) CollisionPointColor() cp.FColor {
	return cp.FColor{R: 1, G: 0.1, B: 0.2, A: 1}
}

func (d *Drawer) Data() interface{} {
	return nil
}

func (d *Drawer) drawOutline(
	screen *ebiten.Image,
	path vector.Path,
	r, g, b, a float32,
) {
	sop := &vector.StrokeOptions{}
	sop.Width = d.StrokeWidth
	sop.LineJoin = vector.LineJoinRound
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = a
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillAll
	op.AntiAlias = d.AntiAlias
	screen.DrawTriangles(vs, is, d.whiteImage, op)
}

func (d *Drawer) drawFill(
	screen *ebiten.Image,
	path vector.Path,
	r, g, b, a float32,
) {
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = a
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillAll
	op.AntiAlias = d.AntiAlias
	screen.DrawTriangles(vs, is, d.whiteImage, op)
}

func (d *Drawer) HandleMouseEvent(space *cp.Space) {
	d.handler.handleMouseEvent(
		space,
		d.ScreenWidth,
		d.ScreenHeight,
	)
}

// event handling

const GRABBABLE_MASK_BIT uint = 1 << 31

var grabFilter cp.ShapeFilter = cp.ShapeFilter{
	Group:      cp.NO_GROUP,
	Categories: GRABBABLE_MASK_BIT,
	Mask:       GRABBABLE_MASK_BIT,
}

type mouseEventHandler struct {
	mouseJoint *cp.Constraint
	mouseBody  *cp.Body
	touchIDs   []ebiten.TouchID
}

func (h *mouseEventHandler) handleMouseEvent(space *cp.Space, screenWidth, screenHeight int) {
	if h.mouseBody == nil {
		h.mouseBody = cp.NewKinematicBody()
	}

	var x, y int

	// touch position
	for _, id := range h.touchIDs {
		x, y = ebiten.TouchPosition(id)
		if x == 0 && y == 0 || inpututil.IsTouchJustReleased(id) {
			h.onMouseUp(space)
			h.touchIDs = []ebiten.TouchID{}
			break
		}
	}
	isJuestTouched := false
	touchIDs := inpututil.AppendJustPressedTouchIDs(h.touchIDs[:0])
	for _, id := range touchIDs {
		isJuestTouched = true
		h.touchIDs = []ebiten.TouchID{id}
		x, y = ebiten.TouchPosition(id)
		break
	}
	// h.touchIDs = inpututil.AppendJustPressedTouchIDs(h.touchIDs[:0])
	// for _, id := range h.touchIDs {
	// 	x, y = ebiten.TouchPosition(id)
	// 	if inpututil.IsTouchJustReleased(id) {
	// 		h.onMouseUp(space)
	// 		return
	// 	}
	// }

	// mouse position
	if len(h.touchIDs) == 0 {
		x, y = ebiten.CursorPosition()
	}

	x -= screenWidth / 2
	y = -y + screenHeight/2

	cursorPosition := cp.Vector{X: float64(x), Y: float64(y)}
	if isJuestTouched {
		h.mouseBody.SetVelocityVector(cp.Vector{})
		h.mouseBody.SetPosition(cursorPosition)
	} else {
		newPoint := h.mouseBody.Position().Lerp(cursorPosition, 0.25)
		h.mouseBody.SetVelocityVector(newPoint.Sub(h.mouseBody.Position()).Mult(60.0))
		h.mouseBody.SetPosition(newPoint)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || isJuestTouched {
		h.onMouseDown(space, cursorPosition)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		h.onMouseUp(space)
	}
}

func (h *mouseEventHandler) onMouseDown(space *cp.Space, cursorPosition cp.Vector) {
	// give the mouse click a little radius to make it easier to click small shapes.
	radius := 5.0

	info := space.PointQueryNearest(cursorPosition, radius, grabFilter)

	if info.Shape != nil && info.Shape.Body().Mass() < cp.INFINITY {
		var nearest cp.Vector
		if info.Distance > 0 {
			nearest = info.Point
		} else {
			nearest = cursorPosition
		}

		body := info.Shape.Body()
		h.mouseJoint = cp.NewPivotJoint2(h.mouseBody, body, cp.Vector{}, body.WorldToLocal(nearest))
		h.mouseJoint.SetMaxForce(50000)
		h.mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
		space.AddConstraint(h.mouseJoint)
	}
}

func (h *mouseEventHandler) onMouseUp(space *cp.Space) {
	if h.mouseJoint == nil {
		return
	}
	space.RemoveConstraint(h.mouseJoint)
	h.mouseJoint = nil
}
