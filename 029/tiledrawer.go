package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/demouth/ebitengine-sketch/029/drawer"
	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func BigTileDrawerFactory() TileDrawer {
	drawers := []TileDrawer{
		// OutBack
		&TileDrawer11{directionType: 0},
		&TileDrawer11{directionType: 1},
		// shutter
		&TileDrawer12{directionType: 0},
		&TileDrawer12{directionType: 1},
	}
	r := rand.Intn(len(drawers))
	return drawers[r]
}
func MiddleTileDrawerFactory() TileDrawer {
	drawers := []TileDrawer{
		// circle
		&TileDrawer1{},
		&TileDrawer2{},
		&TileDrawer3{},
		&TileDrawer4{},
		// packman
		&TileDrawer9{},
		&TileDrawer9{baseAngle: math.Pi / 2},
		&TileDrawer9{baseAngle: math.Pi},
		&TileDrawer9{baseAngle: math.Pi * 1.5},
	}
	r := rand.Intn(len(drawers))
	return drawers[r]
}
func SmallTileDrawerFactory(x, y int) TileDrawer {
	drawers := []TileDrawer{
		// rect
		&TileDrawer5{},
		&TileDrawer6{},
		&TileDrawer7{},
		&TileDrawer8{},
		// arc
		&TileDrawer10{x: x, y: y},
	}
	r := rand.Intn(len(drawers))
	return drawers[r]
}

func easing(t float32) float32 {
	return float32(ease.OutExpo(float64(t)))
}

type TileDrawer1 struct{}

func (d *TileDrawer1) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	drawer.DrawCircle(screen,
		-x/2+x*t,
		y/2,
		x/2,
		c,
	)
}

type TileDrawer2 struct{}

func (d *TileDrawer2) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	drawer.DrawCircle(screen,
		x*1.5-x*t,
		y/2,
		x/2,
		c,
	)
}

type TileDrawer3 struct{}

func (d *TileDrawer3) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	drawer.DrawCircle(screen,
		x/2,
		-y/2+y*t,
		x/2,
		c,
	)
}

type TileDrawer4 struct{}

func (d *TileDrawer4) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	drawer.DrawCircle(screen,
		x/2,
		y*1.5-y*t,
		x/2,
		c,
	)
}

type TileDrawer5 struct{}

func (d *TileDrawer5) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path.MoveTo(0, 0)
	path.LineTo(0, y*t)
	path.LineTo(x, y*t)
	path.LineTo(x, 0)
	path.Close()
	drawer.DrawFill(screen, path, c)
}

type TileDrawer6 struct{}

func (d *TileDrawer6) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path.MoveTo(0, y)
	path.LineTo(0, y-y*t)
	path.LineTo(x, y-y*t)
	path.LineTo(x, y)
	path.Close()
	drawer.DrawFill(screen, path, c)
}

type TileDrawer7 struct{}

func (d *TileDrawer7) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path.MoveTo(0, 0)
	path.LineTo(x*t, 0)
	path.LineTo(x*t, y)
	path.LineTo(0, y)
	path.Close()
	drawer.DrawFill(screen, path, c)
}

type TileDrawer8 struct{}

func (d *TileDrawer8) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path.MoveTo(x, 0)
	path.LineTo(x-x*t, 0)
	path.LineTo(x-x*t, y)
	path.LineTo(x, y)
	path.Close()
	drawer.DrawFill(screen, path, c)
}

type TileDrawer9 struct {
	baseAngle float32
}

func (d *TileDrawer9) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path.MoveTo(x/2, y/2)
	path.Arc(x/2, y/2, x/2, d.baseAngle, d.baseAngle+math.Pi*2*(1-t), vector.CounterClockwise)
	path.Close()
	drawer.DrawFill(screen, path, c)
}

type TileDrawer10 struct {
	x, y int
}

func (d *TileDrawer10) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	if d.x%2 == 1 && d.y%2 == 1 {
		path.MoveTo(0, 0)
		path.Arc(0, 0, x, 0, math.Pi/2*(1-t), vector.CounterClockwise)
		path.Close()
	} else if d.x%2 == 1 && d.y%2 == 0 {
		path.MoveTo(0, y)
		path.Arc(0, y, x, math.Pi/2, -math.Pi/2+math.Pi/2*(1-t), vector.CounterClockwise)
		path.Close()
	} else if d.x%2 == 0 && d.y%2 == 0 {
		path.MoveTo(x, y)
		path.Arc(x, y, x, math.Pi, -math.Pi+math.Pi/2*(1-t), vector.CounterClockwise)
		path.Close()
	} else if d.x%2 == 0 && d.y%2 == 1 {
		path.MoveTo(x, 0)
		path.Arc(x, 0, x, -math.Pi/2, math.Pi/2+math.Pi/2*(1-t), vector.CounterClockwise)
		path.Close()
	}
	drawer.DrawFill(screen, path, c)
}

type TileDrawer11 struct {
	original      *ebiten.Image
	directionType int
}

func (d *TileDrawer11) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	if d.original == nil {
		size := screen.Bounds().Size()
		original := ebiten.NewImage(size.X, size.Y)
		original.DrawImage(screen, nil)
		d.original = original
	}

	// t.drawer.Draw(copyImage, t.color, r)

	screen.DrawImage(d.original, nil)

	path := vector.Path{}
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	if t > 1 {
		t = 1
	}
	t = float32(ease.OutBack(float64(t)))

	path.MoveTo(0, 0)
	path.LineTo(0, y)
	path.LineTo(x, y)
	path.LineTo(x, 0)
	path.Close()

	geoM := ebiten.GeoM{}
	switch d.directionType {
	case 0:
		geoM.Translate(0, float64(y-t*y))
	case 1:
		geoM.Translate(float64(x-t*x), 0)
	}
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		dstX, dstY := float64(vs[i].DstX), float64(vs[i].DstY)
		dstX, dstY = geoM.Apply(dstX, dstY)
		vs[i].DstX, vs[i].DstY = float32(dstX), float32(dstY)
		vs[i].ColorR = float32(c.R) / float32(0xff)
		vs[i].ColorG = float32(c.G) / float32(0xff)
		vs[i].ColorB = float32(c.B) / float32(0xff)
		vs[i].ColorA = float32(c.A) / float32(0xff)
	}

	drawer.DrawTriangles(screen, vs, is)
}

type TileDrawer12 struct {
	directionType int
}

func (d *TileDrawer12) Draw(screen *ebiten.Image, c color.RGBA, t float32) {
	size := screen.Bounds().Size()
	x, y := float32(size.X), float32(size.Y)
	t = easing(t)
	path := vector.Path{}
	num := 4
	var step float32
	if d.directionType == 0 {
		step = y / float32(num)
	} else {
		step = x / float32(num)
	}
	for i := 0; i < num; i++ {
		if d.directionType == 0 {
			path.MoveTo(0, step*float32(i))
			path.LineTo(0, step*float32(i)+step*t)
			path.LineTo(x, step*float32(i)+step*t)
			path.LineTo(x, step*float32(i))
			path.Close()
		} else {
			path.MoveTo(step*float32(i), 0)
			path.LineTo(step*float32(i)+step*t, 0)
			path.LineTo(step*float32(i)+step*t, y)
			path.LineTo(step*float32(i), y)
			path.Close()
		}
	}
	drawer.DrawFill(screen, path, c)
}
