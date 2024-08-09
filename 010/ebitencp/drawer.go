package ebitencp

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp/v2"
)

var (
	whiteImage = ebiten.NewImage(3, 3)
)

type v2f struct {
	x, y float32
}

const DrawPointLineScale = 1

func V2f(v cp.Vector) v2f {
	return v2f{float32(v.X), float32(v.Y)}
}
func v2f0() v2f {
	return v2f{0, 0}
}

type Vertex struct {
	vertex, aa_coord          v2f
	fill_color, outline_color cp.FColor
}

func init() {
	whiteImage.Fill(color.White)
}

type Drawer struct {
	whiteImage *ebiten.Image
	Screen     *ebiten.Image

	ScreenWidth  int
	ScreenHeight int

	vertices []ebiten.Vertex
	indices  []uint16
}

func NewDrawer(screenWidth, screenHeight int) *Drawer {
	return &Drawer{
		whiteImage:   ebiten.NewImage(3, 3),
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

func (d *Drawer) drawWithVerticesAndIndices(screen *ebiten.Image, vertices []ebiten.Vertex, indices []uint16) {
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.NonZero
	// op.AntiAlias = true
	screen.DrawTriangles(vertices, indices, whiteImage, op)
}

func (d *Drawer) draw(
	screen *ebiten.Image,
	path vector.Path,
	r, g, b, a float32,
) {
	var vs []ebiten.Vertex
	var is []uint16
	sop := &vector.StrokeOptions{}
	sop.Width = 1
	sop.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = a
	}
	d.drawWithVerticesAndIndices(screen, vs, is)

	vs, is = path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = r
		vs[i].ColorG = g
		vs[i].ColorB = b
		vs[i].ColorA = a * 0.5
	}
	d.drawWithVerticesAndIndices(screen, vs, is)
}

func (d *Drawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data interface{}) {

	var path *vector.Path = &vector.Path{}
	path.Arc(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2,
		float32(radius),
		0, 2*math.Pi, vector.Clockwise)
	path.MoveTo(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2)
	path.LineTo(
		float32(pos.X+math.Cos(angle)*radius)+float32(d.ScreenWidth)/2,
		-float32(pos.Y+math.Sin(angle)*radius)+float32(d.ScreenHeight)/2)
	path.Close()

	d.draw(d.Screen, *path, outline.R, outline.G, outline.B, outline.A)
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
	d.draw(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
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
	d.draw(d.Screen, path, outline.R, outline.G, outline.B, outline.A)
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
		v0 := V2f(verts[0].Add(extrude[0].offset.Mult(inset)))
		v1 := V2f(verts[i+1].Add(extrude[i+1].offset.Mult(inset)))
		v2 := V2f(verts[i+2].Add(extrude[i+2].offset.Mult(inset)))

		path.MoveTo(
			float32(v0.x)+float32(d.ScreenWidth)/2,
			-float32(v0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v1.x)+float32(d.ScreenWidth)/2,
			-float32(v1.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v2.x)+float32(d.ScreenWidth)/2,
			-float32(v2.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(v0.x)+float32(d.ScreenWidth)/2,
			-float32(v0.y)+float32(d.ScreenHeight)/2)
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

		inner0 := V2f(innerA)
		inner1 := V2f(innerB)
		outer0 := V2f(innerA.Add(nB.Mult(outset)))
		outer1 := V2f(innerB.Add(nB.Mult(outset)))
		outer2 := V2f(innerA.Add(offsetA.Mult(outset)))
		outer3 := V2f(innerA.Add(nA.Mult(outset)))

		path.MoveTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner1.x)+float32(d.ScreenWidth)/2,
			-float32(inner1.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer1.x)+float32(d.ScreenWidth)/2,
			-float32(outer1.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer0.x)+float32(d.ScreenWidth)/2,
			-float32(outer0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer1.x)+float32(d.ScreenWidth)/2,
			-float32(outer1.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer0.x)+float32(d.ScreenWidth)/2,
			-float32(outer0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer2.x)+float32(d.ScreenWidth)/2,
			-float32(outer2.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)

		path.MoveTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer2.x)+float32(d.ScreenWidth)/2,
			-float32(outer2.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(outer3.x)+float32(d.ScreenWidth)/2,
			-float32(outer3.y)+float32(d.ScreenHeight)/2)
		path.LineTo(
			float32(inner0.x)+float32(d.ScreenWidth)/2,
			-float32(inner0.y)+float32(d.ScreenHeight)/2)

		j = i
		i++
	}

	d.draw(d.Screen, *path, outline.R, outline.G, outline.B, outline.A)
}
func (d *Drawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data interface{}) {

	var path *vector.Path = &vector.Path{}
	path.Arc(
		float32(pos.X)+float32(d.ScreenWidth)/2,
		-float32(pos.Y)+float32(d.ScreenHeight)/2,
		float32(2),
		0, 2*math.Pi, vector.Clockwise)
	path.Close()

	d.draw(d.Screen, *path, fill.R, fill.G, fill.B, fill.A)
}

func (d *Drawer) Flags() uint {
	return 0
}

func (d *Drawer) OutlineColor() cp.FColor {
	return cp.FColor{R: 0.1, G: 0.6, B: 0.3, A: 1}
}

func (d *Drawer) ShapeColor(shape *cp.Shape, data interface{}) cp.FColor {
	body := shape.Body()
	if body.IsSleeping() {
		return cp.FColor{R: .2, G: .2, B: .2, A: 1}
	}

	if body.IdleTime() > shape.Space().SleepTimeThreshold {
		return cp.FColor{R: .66, G: .66, B: .66, A: 1}
	}
	return cp.FColor{R: 0.1, G: 0.6, B: 0.3, A: 1}
}

func (d *Drawer) ConstraintColor() cp.FColor {
	return cp.FColor{R: 0.9, G: 0.6, B: 0.3, A: 1}
}

func (d *Drawer) CollisionPointColor() cp.FColor {
	return cp.FColor{R: 1, G: 0.1, B: 0.2, A: 1}
}

func (d *Drawer) Data() interface{} {
	return nil
}

func (d *Drawer) Draw() {
	d.drawWithVerticesAndIndices(d.Screen, d.vertices, d.indices)
	d.vertices = []ebiten.Vertex{}
	d.indices = []uint16{}
}

/*
func (d *EbitenDrawer) addPath(
	path vector.Path,
	r, g, b, a float32,
) {
	var vs []ebiten.Vertex
	var is []uint16
	sop := &vector.StrokeOptions{}
	sop.Width = 1
	sop.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	// for i := range vs {
	// 	vs[i].SrcX = 1
	// 	vs[i].SrcY = 1
	// 	vs[i].ColorR = r
	// 	vs[i].ColorG = g
	// 	vs[i].ColorB = b
	// 	vs[i].ColorA = a
	// }
	d.vertices = append(d.vertices, vs...)
	d.indices = append(d.indices, is...)
}
*/
