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
	Tomato Kind = 1 + iota
	Onion
	Eggplant
	Cucumber
	Carrot
	Pumpkin
	Whiteradish

	Min Kind = Tomato
	Max Kind = Whiteradish
)

var (
	//go:embed tomato.png
	tomato_png []byte
	//go:embed cucumber.png
	cucumber_png []byte
	//go:embed onion.png
	onion_png []byte
	//go:embed pumpkin.png
	pumpkin_png []byte
	//go:embed eggplant.png
	eggplant_png []byte
	//go:embed carrot.png
	carrot_png []byte
	//go:embed whiteradish.png
	whiteradish_png []byte

	assets map[Kind]ImageSet
)

type Kind int

func (k Kind) Next() (hasNext bool, next Kind) {
	if k < Max {
		return true, k + 1
	}
	return false, 0
}

func (k Kind) Score() int {
	return Get(k).Score
}

type ImageSet struct {
	EbitenImage *ebiten.Image
	Image       image.Image
	Scale       float64
	Vectors     []cp.Vector
	Score       int
}

func init() {
	tomatoPngImage, tomatoImage := loadImage(tomato_png)
	cucumberPngImage, cucumberImage := loadImage(cucumber_png)
	pumpkinPngImage, pumpkinImage := loadImage(pumpkin_png)
	carrotPngImage, carrotImage := loadImage(carrot_png)
	onionPngImage, onionImage := loadImage(onion_png)
	eggplantPngImage, eggplantImage := loadImage(eggplant_png)
	whiteradishPngImage, whiteradishImage := loadImage(whiteradish_png)

	assets = map[Kind]ImageSet{
		Tomato:      makeImageSet(tomatoImage, tomatoPngImage, 0.4, 10),
		Onion:       makeImageSet(onionImage, onionPngImage, 1.0, 20),
		Eggplant:    makeImageSet(eggplantImage, eggplantPngImage, 1.2, 30),
		Cucumber:    makeImageSet(cucumberImage, cucumberPngImage, 1.4, 40),
		Carrot:      makeImageSet(carrotImage, carrotPngImage, 1.1, 50),
		Pumpkin:     makeImageSet(pumpkinImage, pumpkinPngImage, 1.9, 60),
		Whiteradish: makeImageSet(whiteradishImage, whiteradishPngImage, 1.2, 70),
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

func Get(tp Kind) ImageSet {
	is, ok := assets[tp]
	if !ok {
		log.Fatalf("image %d not found", tp)
	}
	return is
}

func Length() int {
	return len(assets)
}

func Exists(tp Kind) bool {
	return tp >= Min && tp <= Max
}

func ForEach(f func(Kind, ImageSet)) {
	for i, v := range assets {
		f(i, v)
	}
}

func makeImageSet(
	ebitenImage *ebiten.Image,
	image image.Image,
	scale float64,
	score int,
) ImageSet {
	is := ImageSet{
		EbitenImage: ebitenImage,
		Image:       image,
		Scale:       scale,
		Vectors:     makeVector(image, scale),
		Score:       score,
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
