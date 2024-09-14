package main

import (
	"bytes"
	"fmt"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math/rand"

	"github.com/demouth/ebitengine-sketch/025/drawer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/tdewolff/canvas"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	screenWidth  = 600
	screenHeight = 600
)

var (
	textFaceSource *text.GoTextFaceSource
	segments       []canvas.Segment
)

func init() {
	/*
		// s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
		s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
		if err != nil {
			log.Fatal(err)
		}
		textFaceSource = s

		b, err := io.ReadAll(bytes.NewReader(goregular.TTF))
		if err != nil {
			log.Fatal(err)
		}
		font_, err := truetype.Parse(b)
		freetype.ParseFont(b)
		if err != nil {
			log.Fatal(err)
		}
		font_.VMetric(10, 10)
		face := truetype.NewFace(font_, &truetype.Options{
			Size: 16, // これがフォントサイズ。
		})
		face.Glyph(fixed.Point26_6{X: 1, Y: 2}, 'a')

		// 描画用の構造体を準備する。
		img := image.NewRGBA(image.Rect(0, 0, 256, 128))
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.Black),
			Face: face,
			Dot:  fixed.Point26_6{X: fixed.I(30)},
		}
		d.DrawString("Hello, World! こんにちは、世界！\nThis is a test.")
	*/
	// path := vector.Path{}
	// text.AppendVectorPath(path, "A")

	fontFamily := canvas.NewFontFamily("goregular")
	b2, err := io.ReadAll(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	if err := fontFamily.LoadFont(b2, 0, canvas.FontRegular); err != nil {
		panic(err)
	}
	face__ := fontFamily.Face(14.0, canvas.Black, canvas.FontBold, canvas.FontNormal)
	path, _, err := face__.ToPath("A")
	if err != nil {
		panic(err)
	}
	segments = path.Segments()
	fmt.Println(segments)
}

type Game struct {
	segments []canvas.Segment
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{0x66, 0x66, 0x66, 0xff})

	for _, segment := range g.segments {
		drawer.DrawLine(screen,
			float32(segment.Start.X)*80+100+rand.Float32()*8, -float32(segment.Start.Y)*80+400+rand.Float32()*8,
			float32(segment.End.X)*80+100+rand.Float32()*8, -float32(segment.End.Y)*80+400+rand.Float32()*8,
			1,
			color.RGBA{0xff, 0xff, 0xff, 0xff},
		)
	}

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
		segments: segments,
	}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine + Canvas")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
