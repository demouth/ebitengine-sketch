// porting https://beautifl.net/run/106/
package main

import (
	"fmt"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 600
	screenHeight = 600
	maxAngle     = 256
	ALLOWFLOOR   = true
	ALLOWGRAVITY = true
	floorLevel   = 200
	gravity      = .35
)

type Particle struct {
	vel      Vec3D // velocity
	loc      Vec3D // position
	age      float64
	lifeSpan float64
	radius   float64
	trail    []Vec3D
}

func (p *Particle) Move() {
	if ALLOWGRAVITY {
		p.vel.y += gravity
	}
	if ALLOWFLOOR {
		if p.loc.y > floorLevel {
			p.loc.y = floorLevel
			p.vel.Scale(0.75)
			p.vel.y *= -0.5
		}
	}

	// ref to https://beautifl.net/run/192/
	d := p.loc.x + p.loc.y + p.loc.z
	p.vel.x += math.Cos(d) * 0.5
	p.vel.y += math.Sin(d)*0.5 + 0.35
	p.vel.z += math.Cos(d) * -0.5

	p.loc.AddSelf(p.vel)
	p.trail = append(p.trail, p.loc.Clone())
	const max = 10
	if len(p.trail) > max {
		p.trail = p.trail[len(p.trail)-max:]
	}
	p.age += 1
}

func (p *Particle) IsDead() bool {
	return p.AgePer() <= 0
}

// range from 1.0 (birth) to 0.0 (death)
func (p *Particle) AgePer() float64 {
	a := p.age / p.lifeSpan
	agePer := 1 - a
	return agePer
}

type Game struct {
	touchIDs      []ebiten.TouchID
	particles     []*Particle
	inited        bool
	renderer      *Renderer
	emitter       *Emitter
	isMouseDown   bool
	isTouching    bool
	isTouchDevice bool
}

func (g *Game) init() {
	defer func() {
		g.inited = true
	}()
	g.emitter = &Emitter{}
	g.renderer = NewRenderer()
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}

	g.isTouching = false
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	for _ = range g.touchIDs {
		g.isTouching = true
		g.isTouchDevice = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.isMouseDown = true
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.isMouseDown = false
	}

	if g.isTouching {
		x, y := ebiten.TouchPosition(g.touchIDs[0])
		x -= screenWidth / 2
		y -= screenHeight / 2
		g.emitter.Move(float64(x), float64(y))
	} else if !g.isTouchDevice {
		mouseX, mouseY := ebiten.CursorPosition()
		mouseX -= screenWidth / 2
		mouseY -= screenHeight / 2
		g.emitter.Move(float64(mouseX), float64(mouseY))
	}
	if g.isTouching || g.isMouseDown {
		g.particles = append(g.particles, g.emitter.Emit(10)...)
	}

	newParticles := make([]*Particle, 0)
	for _, p := range g.particles {
		p.Move()
		if !p.IsDead() {
			newParticles = append(newParticles, p)
		}
	}
	g.particles = newParticles

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.particles, g.emitter)

	msg := fmt.Sprintf(
		"particles: %d\nFPS: %0.2f",
		len(g.particles),
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine Demo")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
