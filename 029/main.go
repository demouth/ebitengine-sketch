package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math/rand"

	"github.com/demouth/ebitengine-sketch/029/colorpallet"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1152
	screenHeight = 1152
)

type Game struct {
	canvas      *ebiten.Image
	timer       float32
	tiles       Tiles
	colorpallet *colorpallet.Colors
}
type Tile struct {
	image  *ebiten.Image
	drawer TileDrawer
	color  color.RGBA
}

func (t *Tile) Draw(r float32) {

	// パフォーマンスは悪いが、楽に実装するため画像をコピーして描画する

	size := t.image.Bounds().Size()
	copyImage := ebiten.NewImage(size.X, size.Y)
	copyImage.DrawImage(t.image, nil)

	t.drawer.Draw(copyImage, t.color, r)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(t.image.Bounds().Min.X), float64(t.image.Bounds().Min.Y))
	t.image.DrawImage(copyImage, op)
}

type Tiles []*Tile
type TileDrawer interface {
	Draw(*ebiten.Image, color.RGBA, float32)
}

func (g *Game) makeTiles(canvas *ebiten.Image) Tiles {
	const step = 96
	const middleStep = step * 2
	const bigStep = step * 4
	tiles := make(Tiles, 0)

	for x := 0; x < screenWidth; x += bigStep {
		for y := 0; y < screenHeight; y += bigStep {
			if rand.Float32() < 0.9 {
				continue
			}
			tileImage, _ := canvas.SubImage(
				image.Rect(x, y, x+bigStep, y+bigStep),
			).(*ebiten.Image)
			tile := &Tile{
				image:  tileImage,
				drawer: BigTileDrawerFactory(),
				color:  g.colorpallet.Random(),
			}
			tiles = append(tiles, tile)
		}
	}

	for x := 0; x < screenWidth; x += middleStep {
		for y := 0; y < screenHeight; y += middleStep {
			if rand.Float32() < 0.6 {
				continue
			}
			tileImage, _ := canvas.SubImage(
				image.Rect(x, y, x+middleStep, y+middleStep),
			).(*ebiten.Image)

			overlaps := false
			for _, t := range tiles {
				if t.image.Bounds().Overlaps(tileImage.Bounds()) {
					overlaps = true
					continue
				}
			}
			if overlaps {
				continue
			}

			tile := &Tile{
				image:  tileImage,
				drawer: MiddleTileDrawerFactory(),
				color:  g.colorpallet.Random(),
			}
			tiles = append(tiles, tile)
		}
	}

	for x := 0; x < screenWidth; x += step {
		for y := 0; y < screenHeight; y += step {
			if rand.Float32() < 0.2 {
				continue
			}
			tileImage, _ := canvas.SubImage(image.Rectangle{
				Min: image.Point{X: x, Y: y},
				Max: image.Point{x + step, y + step},
			}).(*ebiten.Image)

			overlaps := false
			for _, t := range tiles {
				if t.image.Bounds().Overlaps(tileImage.Bounds()) {
					overlaps = true
					continue
				}
			}
			if overlaps {
				continue
			}

			tile := &Tile{
				image:  tileImage,
				drawer: SmallTileDrawerFactory(x/step, y/step),
				color:  g.colorpallet.Random(),
			}
			tiles = append(tiles, tile)
		}
	}
	return tiles
}

func (g *Game) Update() error {
	if g.timer > 1.2 {
		g.timer = 0
		g.tiles = g.makeTiles(g.canvas)
	}
	g.timer += 0.01
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, tile := range g.tiles {
		tile.Draw(g.timer)
	}
	screen.DrawImage(g.canvas, nil)
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %0.2f",
		ebiten.ActualFPS(),
	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{
		canvas:      ebiten.NewImage(screenWidth, screenHeight),
		timer:       1,
		colorpallet: colorpallet.NewColors(2),
	}
	game.canvas.Fill(color.White)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("easing")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
