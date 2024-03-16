package main

import (
	_ "embed"
	"image"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	spriteImage *ebiten.Image

	//go:embed assets/sprite.png
	sprite_png []byte
)

type SpriteDrawer struct {
	op ebiten.DrawImageOptions
}

func init() {
	spriteImage = loadImage(sprite_png)
}

func (d *SpriteDrawer) World(screen *ebiten.Image, world World) {
	vector.DrawFilledRect(
		screen,
		float32(world.X), float32(world.Y), float32(world.Width), float32(world.Height),
		color.RGBA{0x66, 0x66, 0x66, 0xff},
		false,
	)
}

func (d *SpriteDrawer) Fruit(screen *ebiten.Image, world World, f *Fruit) {
	var img *ebiten.Image
	img = spriteImage

	w := 48
	h := 104
	offsetX := 0
	offsetY := 0

	tm := int64(f.TotalMovement)
	frame := (tm / 30) % 4

	switch frame {
	case 0:
		offsetX = 8
	case 1:
		offsetX = 72
	case 2:
		offsetX = 136
	case 3:
		offsetX = 200
	}
	switch f.Direction {
	case BOTTOM:
		offsetY = 12
	case RIGHT:
		offsetY = 144
	case TOP:
		offsetY = 276
	case LEFT:
		offsetY = 412
	}
	scale := f.Radius / float64(w) * 2
	wc := float64(w) / 2
	d.op.Filter = ebiten.FilterLinear
	d.op.GeoM.Reset()
	d.op.GeoM.Translate(-wc, wc-float64(h))
	d.op.GeoM.Scale(scale, scale)
	d.op.GeoM.Translate(float64(world.X), float64(world.Y))
	d.op.GeoM.Translate(float64(f.X), float64(f.Y))
	subImg := img.SubImage(image.Rect(offsetX, offsetY, w+offsetX, h+offsetY)).(*ebiten.Image)
	screen.DrawImage(subImg, &d.op)
}

func (d *SpriteDrawer) Fruits(screen *ebiten.Image, world World, fruits []*Fruit) {
	sortedFruits := sortFruitsByY(fruits)

	l := len(sortedFruits)
	for i := 0; i < l; i++ {
		f := sortedFruits[i]
		d.Fruit(screen, world, f)
	}
}

func sortFruitsByY(fruits []*Fruit) []*Fruit {
	l := len(fruits)
	sortedFruits := make([]*Fruit, l)
	copy(sortedFruits, fruits)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			if sortedFruits[i].Y > sortedFruits[j].Y {
				sortedFruits[i], sortedFruits[j] = sortedFruits[j], sortedFruits[i]
			}
		}
	}
	return sortedFruits
}
