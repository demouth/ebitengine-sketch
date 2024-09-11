package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/solarlune/tetra3d"
	"github.com/solarlune/tetra3d/colors"
	"github.com/solarlune/tetra3d/examples"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/demouth/ebitengine-sketch/022/colorpallet"
)

type Game struct {
	Scene         *tetra3d.Scene
	Camera        examples.BasicFreeCam
	SystemHandler examples.BasicSystemHandler
	cubes         []*Particle
}
type Particle struct {
	model    *tetra3d.Model
	velocity tetra3d.Vector
}

func NewGame() *Game {
	game := &Game{}
	game.Init()
	return game
}

func (g *Game) Init() {
	g.Scene = tetra3d.NewScene("boids")
	g.Scene.World.LightingOn = true
	g.SystemHandler = examples.NewBasicSystemHandler(g)
	colors := colorpallet.NewColors(0)
	for i := 0; i < 100; i++ {
		cube := tetra3d.NewModel("Cube", tetra3d.NewCubeMesh())
		color := colors.Random()
		cube.Color = tetra3d.NewColor(float32(color.R)/255.0, float32(color.G)/255.0, float32(color.B)/255.0, 1)
		g.Scene.Root.AddChildren(cube)
		cube.SetWorldPosition((rand.Float64()-0.5)*30, (rand.Float64()-0.5)*30, (rand.Float64()-0.5)*30)
		vec := tetra3d.Vector{X: rand.Float64() - 0.5, Y: rand.Float64() - 0.5, Z: rand.Float64() - 0.5}
		vec = vec.Unit().Scale(0.04)
		p := &Particle{model: cube, velocity: vec}
		g.cubes = append(g.cubes, p)
	}
	g.Camera = examples.NewBasicFreeCam(g.Scene)
	g.Camera.CameraTilt = -1.4
	g.Camera.Camera.SetLocalPosition(-5, 130, -5)
	g.Camera.SetFar(1000)

	light := tetra3d.NewDirectionalLight("camera light", 1, 1, 1, 1)
	g.Camera.AddChildren(light)
}

const (
	CohesionNeighborhoodRadius float64 = 16.0
	AlignNeighborhoodRadius            = 16.0
	SeparateNeighborhoodRadius         = 8.0
	MaxSpeed                           = 0.2
	MaxSteerForce                      = 0.005
	SeparateWeight                     = 3
	AlignmentWeight                    = 1
	CohesionWeight                     = 1
	AvoidWallWeight                    = 10
	WallSize                           = 40
)

func limit(v *tetra3d.Vector, max float64) {
	if v.Distance(tetra3d.Vector{}) > max {
		*v = v.Unit().Scale(max)
	}
}
func (g *Game) Update() error {
	addVertices := make([]tetra3d.Vector, len(g.cubes))
	for i := 0; i < len(g.cubes); i++ {
		var (
			sepPosSum tetra3d.Vector
			sepCount  int = 0

			aliVelSum tetra3d.Vector
			aliCount  int = 0

			cohPosSum tetra3d.Vector
			cohCount  int = 0
		)
		cube := g.cubes[i].model

		for j := 0; j < len(g.cubes); j++ {
			if i == j {
				continue
			}

			otherCube := g.cubes[j].model
			dist := cube.LocalPosition().Distance(otherCube.LocalPosition())

			// separation
			if dist > 0 && dist < SeparateNeighborhoodRadius {
				repulse := cube.LocalPosition().Sub(otherCube.LocalPosition())
				sepPosSum.Add(repulse.Divide(dist))
				sepCount++
			}

			// alignment
			if dist > 0 && dist < AlignNeighborhoodRadius {
				aliVelSum = aliVelSum.Add(otherCube.LocalPosition())
				aliCount++
			}

			// cohesion
			if dist > 0 && dist < CohesionNeighborhoodRadius {
				cohPosSum = cohPosSum.Add(otherCube.LocalPosition())
				cohCount++
			}
		}

		// separation
		sepSteer := tetra3d.Vector{}
		if sepCount > 0 {
			sepSteer = sepPosSum.Divide(float64(sepCount))
			sepSteer = sepSteer.Unit().Scale(MaxSpeed)
			sepSteer = sepSteer.Sub(g.cubes[i].velocity)
			limit(&sepSteer, MaxSteerForce)
		}

		// alignment
		aliSteer := tetra3d.Vector{}
		if aliCount > 0 {
			aliSteer = aliVelSum.Divide(float64(aliCount))
			aliSteer = aliSteer.Unit().Scale(MaxSpeed)
			aliSteer = aliSteer.Sub(g.cubes[i].velocity)
			limit(&aliSteer, MaxSteerForce)
		}

		// cohesion
		cohSteer := tetra3d.Vector{}
		if cohCount > 0 {
			cohPosSum = cohPosSum.Divide(float64(cohCount))
			cohSteer = cohPosSum.Unit().Scale(MaxSpeed)
			cohSteer = cohSteer.Sub(g.cubes[i].velocity)
			limit(&cohSteer, MaxSteerForce)
		}

		acc := tetra3d.Vector{}
		{
			p := g.cubes[i]
			if p.model.LocalPosition().X > WallSize {
				acc.X = -1
			}
			if p.model.LocalPosition().X < -WallSize {
				acc.X = 1
			}
			if p.model.LocalPosition().Y > WallSize {
				acc.Y = -1
			}
			if p.model.LocalPosition().Y < -WallSize {
				acc.Y = 1
			}
			if p.model.LocalPosition().Z > WallSize {
				acc.Z = -1
			}
			if p.model.LocalPosition().Z < -WallSize {
				acc.Z = 1
			}
			limit(&acc, MaxSteerForce)
		}

		addVec := tetra3d.Vector{}
		addVec = addVec.Add(aliSteer.Scale(AlignmentWeight))
		addVec = addVec.Add(cohSteer.Scale(CohesionWeight))
		addVec = addVec.Add(sepSteer.Scale(SeparateWeight))
		addVec = addVec.Add(acc.Scale(AvoidWallWeight))
		addVertices[i] = addVec
	}

	for i := 0; i < len(g.cubes); i++ {
		cube := g.cubes[i].model
		addVec := addVertices[i]
		g.cubes[i].velocity = g.cubes[i].velocity.Add(addVec)
		g.cubes[i].velocity = g.cubes[i].velocity.Unit().Scale(MaxSpeed)
		cube.MoveVec(g.cubes[i].velocity)

		v := g.cubes[i].velocity
		math.Atan2(v.X, v.Z)
		mat := tetra3d.Matrix4{
			[4]float64{1, 0, 0, 0},
			[4]float64{0, 1, 0, 0},
			[4]float64{0, 0, 1, 0},
			[4]float64{0, 0, 0, 1},
		}
		mat = mat.Rotated(0, 1, 0, math.Atan2(v.X, v.Z))
		mat = mat.Rotated(1, 0, 0, math.Atan2(v.Y, v.Z))
		g.cubes[i].model.SetLocalRotation(mat)
	}

	g.Camera.Update()
	return g.SystemHandler.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{60, 70, 80, 255})
	g.Camera.ClearWithColor(g.Scene.World.FogColor)
	g.Camera.RenderScene(g.Scene)
	screen.DrawImage(g.Camera.ColorTexture(), nil)
	g.SystemHandler.Draw(screen, g.Camera.Camera)
	if g.SystemHandler.DrawDebugText {
		txt := fmt.Sprintf("Camera: %v %v", g.Camera.WorldPosition(), g.Camera.CameraTilt)
		g.Camera.DebugDrawText(screen, txt, 0, 200, 1, colors.LightGray())
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	// This is a fixed aspect ratio; we can change this to, say, extend for wider displays by using the provided w argument and
	// calculating the height from the aspect ratio, then calling Camera.Resize() with the new width and height.
	// return g.Width, g.Height
	return g.Camera.Size()
}

func main() {
	g := NewGame()
	g.Camera.Resize(1000, 1000)

	ebiten.SetWindowTitle("boids")

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	w, h := g.Camera.Size()
	ebiten.SetWindowSize(w, h)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
