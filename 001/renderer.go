package main

import (
	"image"
	"image/color"
	"math"

	"github.com/demouth/colorgradient-go"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Renderer struct {
	particleImage *ebiten.Image
	emitterImage  *ebiten.Image

	whiteImage    *ebiten.Image
	whiteSubImage *ebiten.Image
	vertices      []ebiten.Vertex
	indices       []uint16

	op ebiten.DrawImageOptions
}

func NewRenderer() *Renderer {
	r := &Renderer{}
	r.generateParticleImage()
	r.generateEmitterImage()

	r.whiteImage = ebiten.NewImage(3, 3)
	r.whiteImage.Fill(color.White)
	r.whiteSubImage = r.whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	return r
}

func (r *Renderer) generateParticleImage() {
	w := 80
	h := 80
	fw := float64(w)

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	grad, _ := colorgradient.NewGradient(
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 0, 0, 50},
		color.RGBA{255, 0, 0, 25},
		color.RGBA{255, 0, 0, 0},
	)
	fh := float64(h)
	cw := fw / 2
	ch := fh / 2

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			dx := float64(x) - cw
			dy := float64(y) - ch

			delta := math.Sqrt(dx*dx+dy*dy) / math.Min(cw, ch)
			col := grad.At(delta)
			img.Set(x, y, col)
		}
	}

	origEbitenImage := ebiten.NewImageFromImage(img)

	s := origEbitenImage.Bounds().Size()
	r.particleImage = ebiten.NewImage(s.X, s.Y)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(1)
	r.particleImage.DrawImage(origEbitenImage, op)
}

func (r *Renderer) generateEmitterImage() {
	w := 100
	h := 100
	fw := float64(w)

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	grad, _ := colorgradient.NewGradient(
		color.RGBA{255, 255, 255, 255},
		color.RGBA{255, 255, 255, 50},
		color.RGBA{255, 255, 255, 25},
		color.RGBA{0, 0, 0, 0},
	)
	fh := float64(h)
	cw := fw / 2
	ch := fh / 2

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			dx := float64(x) - cw
			dy := float64(y) - ch

			delta := math.Sqrt(dx*dx+dy*dy) / math.Min(cw, ch)
			col := grad.At(delta)
			img.Set(x, y, col)
		}
	}

	origEbitenImage := ebiten.NewImageFromImage(img)

	s := origEbitenImage.Bounds().Size()
	r.emitterImage = ebiten.NewImage(s.X, s.Y)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(1)
	r.emitterImage.DrawImage(origEbitenImage, op)
}

func (r *Renderer) Draw(screen *ebiten.Image, particles []*Particle, emitter *Emitter) {
	screen.Fill(color.NRGBA{0x00, 0x00, 0x00, 0xff})
	r.drawParticlesTrail(screen, particles)
	r.drawParticles(screen, particles)
	r.drawEmitter(screen, emitter.loc.x, emitter.loc.y)
}

func projectionTo2D(x, y, z float64) (float64, float64, float64) {
	const zFar float64 = 1000
	perspective := zFar / (zFar - z)
	px := x*perspective + screenWidth/2
	py := y*perspective + screenHeight/2
	return px, py, perspective
}

func (r *Renderer) drawParticlesTrail(screen *ebiten.Image, particles []*Particle) {
	for _, s := range particles {
		r.drawParticleTrail(screen, s)
	}
}
func (r *Renderer) drawParticleTrail(screen *ebiten.Image, particle *Particle) {

	if len(particle.trail) < 2 {
		return
	}

	for i, l := 1, len(particle.trail); i < l; i++ {
		var path vector.Path
		x1, y1, _ := projectionTo2D(particle.trail[i-1].x, particle.trail[i-1].y, particle.trail[i-1].z)
		path.MoveTo(float32(x1), float32(y1))
		x2, y2, _ := projectionTo2D(particle.trail[i].x, particle.trail[i].y, particle.trail[i].z)
		path.LineTo(float32(x2), float32(y2))
		op := &vector.StrokeOptions{}
		op.Width = float32(i)/float32(l)*4 + 0.4
		vs, is := path.AppendVerticesAndIndicesForStroke(r.vertices[:0], r.indices[:0], op)
		for i := range vs {
			vs[i].ColorR = 1
			vs[i].ColorG = 0.5
			vs[i].ColorB = 1.0 - float32(i)/float32(l)
			vs[i].ColorA = float32(i)/float32(l)*0.9 + 0.1
		}
		screen.DrawTriangles(vs, is, r.whiteSubImage, &ebiten.DrawTrianglesOptions{
			AntiAlias: false,
		})
	}
}

func (r *Renderer) drawParticles(screen *ebiten.Image, particles []*Particle) {
	for _, s := range particles {
		r.drawParticle(screen, s, s.radius*s.AgePer()*0.01)
	}
}

func (r *Renderer) drawParticle(screen *ebiten.Image, particle *Particle, diam float64) {
	w, h := float64(r.particleImage.Bounds().Dx()), float64(r.particleImage.Bounds().Dy())
	x, y, perspective := projectionTo2D(particle.loc.x, particle.loc.y, particle.loc.z)
	diam *= perspective

	r.op.GeoM.Reset()
	r.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	r.op.GeoM.Scale(particle.radius/w, particle.radius/h)
	r.op.GeoM.Scale(diam, diam)
	r.op.GeoM.Translate(x, y)
	// r.op.GeoM.Translate(screenWidth/2, screenHeight/2)
	// r.op.Blend = ebiten.BlendLighter
	r.op.Filter = ebiten.FilterLinear
	r.op.ColorScale.Reset()
	r.op.ColorScale.Scale(
		float32(particle.AgePer()),
		float32(particle.AgePer()*0.75),
		float32(1-particle.AgePer()),
		0.1,
	)

	// Do not use the colorm package because it is too heavy
	// var c colorm.ColorM
	// c.Reset()
	// c.Scale(1, 1, 1, particle.AgePer())
	// c.Translate(particle.AgePer(), particle.AgePer()*0.75, 1-particle.AgePer(), 0.1)
	// colorm.DrawImage(screen, r.ebitenImage, c, &r.op)

	screen.DrawImage(r.particleImage, &r.op)
}

func (r *Renderer) drawEmitter(screen *ebiten.Image, x, y float64) {
	// const zFar float64 = 1000
	w, h := float64(r.emitterImage.Bounds().Dx()), float64(r.emitterImage.Bounds().Dy())
	radius := float64(100)

	r.op.GeoM.Reset()
	r.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	r.op.GeoM.Scale(radius/w, radius/h)
	r.op.GeoM.Translate(x, y)
	r.op.GeoM.Translate(screenWidth/2, screenHeight/2)
	// r.op.Blend = ebiten.BlendLighter
	r.op.ColorScale.Reset()
	r.op.Filter = ebiten.FilterLinear

	screen.DrawImage(r.emitterImage, &r.op)
}
