// Copyright 2020 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed shader.kage
	shader_kage []byte
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	shader *ebiten.Shader
	idx    int
	time   int
}

func (g *Game) Update() error {
	g.time++
	if g.shader == nil {
		s, err := ebiten.NewShader([]byte(shader_kage))
		if err != nil {
			return err
		}
		g.shader = s
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	s := g.shader

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	cx, cy := ebiten.CursorPosition()

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"Time":   float32(g.time) / 60,
		"Cursor": []float32{float32(cx), float32(cy)},
	}
	// op.Images[0] = gopherImage
	// op.Images[1] = normalImage
	// op.Images[2] = gopherBgImage
	// op.Images[3] = noiseImage
	screen.DrawRectShader(w, h, s, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Shader (Ebitengine Demo)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
