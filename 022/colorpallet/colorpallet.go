package colorpallet

import (
	"image/color"
	"math/rand"
)

type Colors struct {
	colors []color.RGBA
}

func NewColors(palette int) *Colors {
	var colors []color.RGBA
	if palette == 0 {
		// https://openprocessing.org/sketch/1845890
		colors = []color.RGBA{
			{0xDE, 0x18, 0x3C, 0xFF},
			{0xF2, 0xB5, 0x41, 0xFF},
			{0x0C, 0x79, 0xBB, 0xFF},
			{0x2D, 0xAC, 0xB2, 0xFF},
			{0xE4, 0x64, 0x24, 0xFF},
			{0xEC, 0xAC, 0xBE, 0xFF},
			// {0x00, 0x00, 0x00, 0xFF},
			{0x19, 0x44, 0x6B, 0xFF},
		}
	} else if palette == 1 {

		// Vincent Willem van Gogh
		// https://goworkship.com/magazine/artist-inspired-color-palettes/
		colors = []color.RGBA{
			{0x00, 0x39, 0x55, 0xff}, // 003955
			{0x39, 0x7e, 0xc0, 0xff}, // 397ec0
			{0x73, 0x38, 0x37, 0xff}, // 733837
			{0xeb, 0xc7, 0x4b, 0xff}, // ebc74b
			{0x60, 0x7a, 0x4d, 0xff}, // 607a4d
		}
	} else {

		colors = []color.RGBA{
			{0x68, 0x8C, 0x89, 0xFF}, //#688C89
			{0xF2, 0xC1, 0x85, 0xFF}, //#F2C185
			{0x73, 0x02, 0x02, 0xFF}, //#730202
			{0xA6, 0x17, 0x17, 0xFF}, //#A61717
			{0xF2, 0x38, 0x38, 0xFF}, //#F23838
		}
	}
	return &Colors{colors: colors}
}
func (c *Colors) Random() color.RGBA {
	i := rand.Intn(len(c.colors))
	return c.colors[i]
}
func (c *Colors) Color(colorNo uint8) color.RGBA {
	return c.colors[colorNo%uint8(len(c.colors))]
}
func (c *Colors) Len() int {
	return len(c.colors)
}
