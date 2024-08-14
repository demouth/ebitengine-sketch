package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp/v2"
)

const (
	Apple = 1 + iota
	Grape
	Pineapple
	Watermelon
	Orange
	Melon
	Whiteradish
)

var (
	appleImage       *ebiten.Image
	grapeImage       *ebiten.Image
	orangeImage      *ebiten.Image
	pineappleImage   *ebiten.Image
	melonImage       *ebiten.Image
	watermelonImage  *ebiten.Image
	whiteradishImage *ebiten.Image

	applePngImage       image.Image
	grapePngImage       image.Image
	orangePngImage      image.Image
	pineapplePngImage   image.Image
	melonPngImage       image.Image
	watermelonPngImage  image.Image
	whiteradishPngImage image.Image

	//go:embed tomato.png
	apple_png []byte
	//go:embed kyuri.png
	grape_png []byte
	//go:embed tama.png
	orange_png []byte
	//go:embed kabo.png
	pineapple_png []byte
	//go:embed nasu.png
	melon_png []byte
	//go:embed nin.png
	watermelon_png []byte
	//go:embed dai.png
	whiteradish_png []byte

	assets map[int]ImageSet
)

type ImageSet struct {
	EbitenImage *ebiten.Image
	Image       image.Image
	Scale       float64
	Vectors     []cp.Vector
}

func init() {
	applePngImage, appleImage = loadImage(apple_png)
	grapePngImage, grapeImage = loadImage(grape_png)
	pineapplePngImage, pineappleImage = loadImage(pineapple_png)
	watermelonPngImage, watermelonImage = loadImage(watermelon_png)
	orangePngImage, orangeImage = loadImage(orange_png)
	melonPngImage, melonImage = loadImage(melon_png)
	whiteradishPngImage, whiteradishImage = loadImage(whiteradish_png)

	assets = map[int]ImageSet{
		Apple:       makeImageSet(appleImage, applePngImage, 0.5),
		Grape:       makeImageSet(grapeImage, grapePngImage, 0.8),
		Pineapple:   makeImageSet(pineappleImage, pineapplePngImage, 0.9),
		Watermelon:  makeImageSet(watermelonImage, watermelonPngImage, 0.4),
		Orange:      makeImageSet(orangeImage, orangePngImage, 0.7),
		Melon:       makeImageSet(melonImage, melonPngImage, 0.5),
		Whiteradish: makeImageSet(whiteradishImage, whiteradishPngImage, 0.5),
	}
}

func loadImage(b []byte) (image.Image, *ebiten.Image) {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	origImage := ebiten.NewImageFromImage(img)

	s := origImage.Bounds().Size()
	ebitenImage := ebiten.NewImage(s.X, s.Y)
	op := &ebiten.DrawImageOptions{}
	ebitenImage.DrawImage(origImage, op)
	return img, ebitenImage
}

func Get(tp int) ImageSet {
	is, ok := assets[tp]
	if !ok {
		log.Fatalf("image %d not found", tp)
	}
	return is
}

func Length() int {
	return len(assets)
}

func makeImageSet(
	ebitenImage *ebiten.Image,
	image image.Image,
	scale float64,
) ImageSet {
	is := ImageSet{
		EbitenImage: ebitenImage,
		Image:       image,
		Scale:       scale,
		Vectors:     makeVector(image, scale),
	}
	return is
}

func makeVector(img image.Image, scale float64) []cp.Vector {
	b := img.Bounds()
	bb := cp.BB{L: float64(b.Min.X), B: float64(b.Min.Y), R: float64(b.Max.X), T: float64(b.Max.Y)}

	sampleFunc := func(point cp.Vector) float64 {
		x := point.X
		y := point.Y
		rect := img.Bounds()

		if x < float64(rect.Min.X) || x > float64(rect.Max.X) || y < float64(rect.Min.Y) || y > float64(rect.Max.Y) {
			return 0.0
		}
		_, _, _, a := img.At(int(x), int(y)).RGBA()
		return float64(a) / 0xffff
	}

	lineSet := cp.MarchSoft(bb, 300, 300, 0.5, cp.PolyLineCollectSegment, sampleFunc)

	line := lineSet.Lines[0].SimplifyCurves(.9)
	offset := cp.Vector{X: float64(b.Max.X-b.Min.X) / 2., Y: float64(b.Max.Y-b.Min.Y) / 2.}
	// center the verts on origin
	for i, l := range line.Verts {
		line.Verts[i] = l.Sub(offset).Mult(scale)
	}
	return line.Verts
}
