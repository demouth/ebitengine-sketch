package main

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

func init() {
	whiteImage.Fill(color.White)
}

type Ebitencp struct {
}

func (c *Ebitencp) Draw(screen *ebiten.Image, space *cp.Space) {
	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()
	var awakingPath *vector.Path = &vector.Path{}
	var sleptPath *vector.Path = &vector.Path{}
	var path *vector.Path = nil
	space.EachShape(func(shape *cp.Shape) {
		if shape.Body().IsSleeping() {
			path = sleptPath
		} else if shape.Body().IdleTime() > shape.Space().SleepTimeThreshold {
			path = sleptPath
		} else {
			path = awakingPath
		}
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)
			vec := circle.TransformC()
			path.Arc(
				float32(vec.X)+float32(screenWidth)/2,
				float32(vec.Y)+float32(screenHeight)/2,
				float32(circle.Radius()),
				0, 2*math.Pi, vector.Clockwise)
			path.MoveTo(
				float32(vec.X)+float32(screenWidth)/2,
				float32(vec.Y)+float32(screenHeight)/2)
			path.LineTo(
				float32(vec.X+math.Cos(circle.Body().Angle())*circle.Radius())+float32(screenWidth)/2,
				float32(vec.Y+math.Sin(circle.Body().Angle())*circle.Radius())+float32(screenHeight)/2)
			path.Close()
		case *cp.PolyShape:
			poly := shape.Class.(*cp.PolyShape)

			count := poly.Count()
			for i := 0; i < count; i++ {
				vec := poly.TransformVert(i)
				if count == 0 {
					path.MoveTo(
						float32(vec.X)+float32(screenWidth)/2,
						float32(vec.Y)+float32(screenHeight)/2)
				} else {
					path.LineTo(
						float32(vec.X)+float32(screenWidth)/2,
						float32(vec.Y)+float32(screenHeight)/2)
				}
				if i == count-1 {
					vec := poly.TransformVert(i)
					path.LineTo(
						float32(vec.X)+float32(screenWidth)/2,
						float32(vec.Y)+float32(screenHeight)/2)
					path.Close()
				}

			}

		case *cp.Segment:
			segment := shape.Class.(*cp.Segment)
			ta := segment.TransformA()
			tb := segment.TransformB()
			path.MoveTo(
				float32(ta.X)+float32(screenWidth)/2,
				float32(ta.Y)+float32(screenHeight)/2)
			path.LineTo(
				float32(tb.X)+float32(screenWidth)/2,
				float32(tb.Y)+float32(screenHeight)/2)
			path.Close()
		}
	})
	draw(screen, *sleptPath, 0.5, 0.5, 0.5, 1)
	draw(screen, *awakingPath, 0.2, 0.9, 0.5, 1)
}

func draw(
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
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.NonZero
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}
