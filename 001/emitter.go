package main

import "math/rand"

type Emitter struct {
	loc        Vec3D
	vel        Vec3D
	velToMouse Vec3D
}

func (e *Emitter) Move(toX, toY float64) {
	e.velToMouse.Set(toX-e.loc.x, toY-e.loc.y, 0)
	e.vel.InterpolateToSelf(e.velToMouse, 0.35)
	e.loc.AddSelf(e.vel)
}

func (e *Emitter) Emit(amount uint) []*Particle {
	particles := make([]*Particle, 0)
	for i := uint(0); i < amount; i++ {
		r := rand.Float64()*60 + 20
		loc := e.loc.Clone().Add(
			NewVec3D(rand.Float64()-0.5, rand.Float64()-0.5, rand.Float64()-0.5).Scale(5.0),
		)
		vel := e.vel.Scale(.5).Add(
			NewVec3D(rand.Float64()-0.5, rand.Float64()-0.5, rand.Float64()-0.5).Scale(10.0),
		)

		particles = append(particles, &Particle{
			age:      0,
			lifeSpan: r,
			radius:   r,
			loc:      loc,
			vel:      vel,
		})
	}
	return particles
}
