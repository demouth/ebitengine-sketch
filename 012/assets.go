package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
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

	//go:embed assets/apple.png
	apple_png []byte
	//go:embed assets/avacado.png
	grape_png []byte
	//go:embed assets/kiwi.png
	orange_png []byte
	//go:embed assets/strawberry.png
	pineapple_png []byte
	//go:embed assets/melon.png
	melon_png []byte
	//go:embed assets/watermelon.png
	watermelon_png []byte

	assets map[int]ImageSet
)

type ImageSet struct {
	EbitenImage *ebiten.Image
	Image       image.Image
}

func init() {
	applePngImage, appleImage = loadImage(apple_png)
	grapePngImage, grapeImage = loadImage(grape_png)
	pineapplePngImage, pineappleImage = loadImage(pineapple_png)
	watermelonPngImage, watermelonImage = loadImage(watermelon_png)
	// orangePngImage, orangeImage = loadImage(orange_png)
	// melonPngImage, melonImage = loadImage(melon_png)

	assets = map[int]ImageSet{
		0: {EbitenImage: appleImage, Image: applePngImage},
		1: {EbitenImage: grapeImage, Image: grapePngImage},
		2: {EbitenImage: pineappleImage, Image: pineapplePngImage},
		3: {EbitenImage: watermelonImage, Image: watermelonPngImage},
		// 4: {EbitenImage: orangeImage, Image: orangePngImage},
		// 5: {EbitenImage: melonImage, Image: melonPngImage},
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

func GetImageSet(tp int) ImageSet {
	is, ok := assets[tp]
	if !ok {
		log.Fatalf("image %d not found", tp)
	}
	return is
}

func LenImageSet() int {
	return len(assets)
}
