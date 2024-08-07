package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp/v2"
)

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	whiteImage.Fill(color.White)
}

type Ebitencp struct {
}

func (c *Ebitencp) Draw(screen *ebiten.Image, space *cp.Space) {
	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()
	var path vector.Path
	space.EachShape(func(shape *cp.Shape) {
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
			// segment.B()
			path.MoveTo(
				float32(ta.X)+float32(screenWidth)/2,
				float32(ta.Y)+float32(screenHeight)/2)
			path.LineTo(
				float32(tb.X)+float32(screenWidth)/2,
				float32(tb.Y)+float32(screenHeight)/2)
			path.Close()
		}
	})
	var vs []ebiten.Vertex
	var is []uint16
	sop := &vector.StrokeOptions{}
	sop.Width = 1
	sop.LineJoin = vector.LineJoinRound
	vs, is = path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0x33 / float32(0xff)
		vs[i].ColorG = 0xcc / float32(0xff)
		vs[i].ColorB = 0x66 / float32(0xff)
		vs[i].ColorA = 1
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.NonZero
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}
