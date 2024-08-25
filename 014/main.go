package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand/v2"

	"github.com/demouth/ebitencp"
	"github.com/demouth/ebitengine-sketch/014/minigui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jakecoffman/cp/v2"
)

const (
	screenWidth  = 600
	screenHeight = 600
	hwidth       = screenWidth / 2
	hheight      = screenHeight / 2

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

var (
	runnerImage   *ebiten.Image
	whiteSubImage = ebiten.NewImage(3, 3)
)

type Game struct {
	elapsedTime float64
	genTime     float64
	numGen      int
	debugMode   bool
	debugDrawer *ebitencp.Drawer

	space *cp.Space

	gui  *minigui.GUI
	gui2 *minigui.GUI

	gx         float64
	gy         float64
	elasticity float64
	step       float64
}

func (g *Game) Update() error {
	if g.genTime > 0.1 {
		g.genTime = 0
		for i := 0; i < g.numGen; i++ {
			addCircle(
				g.space,
				rand.Float64()*rand.Float64()*rand.Float64()*30+10,
				rand.Float64()*20-10,
				rand.Float64()*20-10-200,
				g.elasticity,
			)
		}
	}

	margin := 10.0
	g.space.EachBody(func(body *cp.Body) {
		remove := false
		if body.Position().Y > hheight+margin {
			remove = true
		} else if body.Position().X < -hwidth-margin {
			remove = true
		} else if body.Position().X > hwidth+margin {
			remove = true
		} else if body.Position().Y < -hheight-margin {
			remove = true
		}
		if remove {
			g.space.AddPostStepCallback(removeBodyCallback, body, nil)
		}
	})

	g.gui.Update()
	g.gui2.Update()

	g.space.Step(g.step)
	g.genTime += g.step
	g.elapsedTime += g.step
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0x66, 0x66, 0x66, 0xff})
	numCircle := 0

	if g.debugMode {
		g.space.EachShape(func(shape *cp.Shape) {
			switch shape.Class.(type) {
			case *cp.Circle:
				numCircle++
			}
		})
		cp.DrawSpace(g.space, g.debugDrawer.WithScreen(screen))
	} else {
		g.space.StaticBody.EachShape(func(shape *cp.Shape) {
			switch shape.Class.(type) {
			case *cp.Segment:
				seg := shape.Class.(*cp.Segment)
				path := vector.Path{}
				path.MoveTo(float32(hwidth+seg.A().X), float32(hheight+seg.A().Y))
				path.LineTo(float32(hwidth+seg.B().X), float32(hheight+seg.B().Y))
				g.drawLine(screen, path, color.NRGBA{0xff, 0xff, 0xff, 0xff}, 5)
			}
		})
		g.space.EachShape(func(shape *cp.Shape) {
			switch shape.Class.(type) {
			case *cp.Circle:
				numCircle++
				circle := shape.Class.(*cp.Circle)
				vec := circle.TransformC()

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
				op.GeoM.Rotate(circle.Body().Angle())
				op.GeoM.Scale(circle.Radius()/10, circle.Radius()/10)
				op.GeoM.Translate(screenWidth/2, screenHeight/2)
				op.GeoM.Translate(vec.X, vec.Y)
				i := int(g.elapsedTime*10.0) % frameCount
				sx, sy := frameOX+i*frameWidth, frameOY
				screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), op)
			}
		})
	}

	g.gui.Draw(screen)
	g.gui2.Draw(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f\nNumCircle: %v",
		ebiten.ActualFPS(),
		numCircle,
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func init() {
	whiteSubImage.Fill(color.White)
}
func main() {
	game := &Game{}
	game.step = 1.0 / 60.0
	game.gy = 100
	game.elasticity = 0.9
	game.numGen = 2
	game.debugDrawer = ebitencp.NewDrawer(screenWidth, screenHeight)
	game.debugDrawer.FlipYAxis = true

	space := cp.NewSpace()
	gravity := cp.Vector{X: game.gx, Y: game.gy}
	space.SetGravity(gravity)
	game.space = space

	addWall(space, -100, 200, +100, 200, 5, game.elasticity)

	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	// gui
	gui := minigui.NewGUI()
	gui.X = screenWidth
	gui.HorizontalAlign = minigui.HorizontalAlignRight
	gui.AddSliderFloat64("Gravity X", game.gx, -500, 500, func(v float64) {
		game.gx = v
		gravity := cp.Vector{X: v, Y: game.gy}
		space.SetGravity(gravity)
	})
	gui.AddSliderFloat64("Gravity Y", game.gy, -500, 500, func(v float64) {
		game.gy = v
		gravity := cp.Vector{X: game.gx, Y: v}
		space.SetGravity(gravity)
	})
	gui.AddSliderFloat64("Elasticity", game.elasticity, 0, 1, func(v float64) {
		game.elasticity = v
		game.space.EachShape(func(shape *cp.Shape) {
			if shape.UserData == "wall" {
				shape.SetElasticity(v)
			}
		})
	})
	gui.AddSliderFloat64("Step", game.step, 1.0/360.0, 1.0/30.0, func(v float64) {
		game.step = v
	})
	gui.AddSliderInt("NumGen", game.numGen, 1, 10, func(v int) {
		game.numGen = v
	})
	gui.AddButton("Debug mode", game.debugMode, func(v bool) {
		game.debugMode = v
	})
	game.gui = gui

	gui2 := minigui.NewGUI()
	gui2.X = 0
	gui2.Y = 100
	gui2.Scale = 1
	gui2.Width = 400
	gui2.AddSliderFloat64("Scale", float64(gui.Scale), 0.1, 2, func(v float64) {
		gui.Scale = float32(v)
	})
	gui2.AddSliderFloat64("X", float64(gui.X), 0, 1000, func(v float64) {
		gui.X = float32(v)
	})
	gui2.AddSliderFloat64("Y", float64(gui.Y), 0, 1000, func(v float64) {
		gui.Y = float32(v)
	})
	gui2.AddSliderFloat64("Width", float64(gui.Width), 10, 1000, func(v float64) {
		gui.Width = float32(v)
	})
	gui2.AddSliderFloat32("ComponentHeight", gui.ComponentHeight, 5, 100, func(v float32) {
		gui.ComponentHeight = v
	})
	gui2.AddButton("Set HorizontalAlign to Right", true, func(v bool) {
		if v {
			gui.HorizontalAlign = minigui.HorizontalAlignRight
		} else {
			gui.HorizontalAlign = minigui.HorizontalAlignLeft
		}
	})
	game.gui2 = gui2

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Chipmunk")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func addCircle(space *cp.Space, radius float64, x, y, elasticity float64) {
	mass := radius * radius / 25.0
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(cp.Vector{X: x, Y: y})

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.96)
	shape.UserData = "circle"
}
func addWall(space *cp.Space, x1, y1, x2, y2, radius, elasticity float64) {
	pos1 := cp.Vector{X: x1, Y: y1}
	pos2 := cp.Vector{X: x2, Y: y2}
	shape := space.AddShape(cp.NewSegment(space.StaticBody, pos1, pos2, radius))
	shape.SetElasticity(elasticity)
	shape.SetFriction(0.5)
	shape.UserData = "wall"
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
func removeShapeCallback(space *cp.Space, key interface{}, data interface{}) {
	var s *cp.Shape
	var ok bool
	if s, ok = key.(*cp.Shape); !ok {
		return
	}
	space.RemoveBody(s.Body())
	space.RemoveShape(s)
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
