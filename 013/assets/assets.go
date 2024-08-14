package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Apple = 1 + iota
	Grape
	Pineapple
	Watermelon
	// Orange
	// Melon
)

var (
	appleImage      *ebiten.Image
	grapeImage      *ebiten.Image
	orangeImage     *ebiten.Image
	pineappleImage  *ebiten.Image
	melonImage      *ebiten.Image
	watermelonImage *ebiten.Image

	applePngImage      image.Image
	grapePngImage      image.Image
	orangePngImage     image.Image
	pineapplePngImage  image.Image
	melonPngImage      image.Image
	watermelonPngImage image.Image

	//go:embed apple.png
	apple_png []byte
	//go:embed avacado.png
	grape_png []byte
	//go:embed kiwi.png
	orange_png []byte
	//go:embed strawberry.png
	pineapple_png []byte
	//go:embed melon.png
	melon_png []byte
	//go:embed watermelon.png
	watermelon_png []byte

	assets map[int]ImageSet
)

type ImageSet struct {
	EbitenImage *ebiten.Image
	Image       image.Image
	Scale       float64
}

func init() {
	applePngImage, appleImage = loadImage(apple_png)
	grapePngImage, grapeImage = loadImage(grape_png)
	pineapplePngImage, pineappleImage = loadImage(pineapple_png)
	watermelonPngImage, watermelonImage = loadImage(watermelon_png)
	orangePngImage, orangeImage = loadImage(orange_png)
	melonPngImage, melonImage = loadImage(melon_png)

	assets = map[int]ImageSet{
		Apple:      {EbitenImage: appleImage, Image: applePngImage, Scale: 1.2},
		Grape:      {EbitenImage: grapeImage, Image: grapePngImage, Scale: 1},
		Pineapple:  {EbitenImage: pineappleImage, Image: pineapplePngImage, Scale: 0.7},
		Watermelon: {EbitenImage: watermelonImage, Image: watermelonPngImage, Scale: 2},
		// Orange:     {EbitenImage: orangeImage, Image: orangePngImage},
		// Melon:      {EbitenImage: melonImage, Image: melonPngImage},
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
