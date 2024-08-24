package minigui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type RangeInterpolator[T float32 | float64 | int] struct {
	Min   T
	Max   T
	Value T
}

func (r *RangeInterpolator[T]) SetRatio(ratio float64) {
	// r.Value = r.Min + (r.Max-r.Min)*T(ratio)

	v64 := float64(r.Min) + (float64(r.Max)-float64(r.Min))*(ratio)
	r.Value = T(v64)
}
func (r *RangeInterpolator[T]) Ratio() float64 {
	return float64(r.Value-r.Min) / float64(r.Max-r.Min)
}
func (r *RangeInterpolator[T]) String() string {
	switch any(r.Value).(type) {
	case float32:
		v := any(r.Value).(float32)
		return fmt.Sprintf("%.3f", v)
	case float64:
		v := any(r.Value).(float64)
		return fmt.Sprintf("%.3f", v)
	default:
		v := any(r.Value).(int)
		return fmt.Sprintf("%d", v)
	}
}

type slider[T float32 | float64 | int] struct {
	label    string
	hovered  bool
	callback func(v T)

	Interpolator RangeInterpolator[T]
}

func UpdateSlider[T float32 | float64 | int](s *slider[T], x, y, width, height, scale float32) {
	width *= scale
	height *= scale
	clicked := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if x >= width/2 && x <= width && y > 0 && y <= height {
		s.hovered = true
	} else {
		s.hovered = false
	}

	if s.hovered && clicked {
		ratio := float64((x - width/2) / width * 2)
		s.Interpolator.SetRatio(ratio)
		s.callback(s.Interpolator.Value)
	}
}
func DrawSliderShape[T float32 | float64 | int](s *slider[T], image *ebiten.Image, whiteImage *ebiten.Image, left, top, width, height, scale float32) {
	padding := 2.0 * scale
	width *= scale
	height *= scale

	x := left + width/2 + padding
	y := top + padding
	w := width/2 - padding*2
	h := height - padding*2

	c := color.NRGBA{0x66, 0x66, 0x66, 0xff}
	if s.hovered {
		c = color.NRGBA{0x79, 0x79, 0x79, 0xff}
	}
	drawRect(image, whiteImage, x, y, w, h, c)

	textPadding := 5.0 * scale
	fontSize := height - textPadding*2
	drawText(image, s.Interpolator.String(), x+textPadding, top+textPadding, fontSize, color.NRGBA{R: 0xeb, G: 0xeb, B: 0xeb, A: 0xff})

	ratio := s.Interpolator.Ratio()
	drawRange := w
	drawWidth := drawRange * float32(ratio)

	drawLine(
		image,
		whiteImage,
		x+drawWidth,
		y,
		x+drawWidth,
		y+h,
		2.0*scale,
		color.NRGBA{0x2c, 0xc9, 0xff, 0xff},
	)
}

// SliderFloat64

type sliderFloat64 struct {
	slider slider[float64]
}

func (g *GUI) AddSliderFloat64(label string, value, min, max float64, callback func(v float64)) {
	s := slider[float64]{
		label:        label,
		Interpolator: RangeInterpolator[float64]{Min: min, Max: max, Value: value},
		callback:     callback,
	}
	slider := &sliderFloat64{slider: s}
	g.components = append(g.components, slider)
}

func (s *sliderFloat64) Label() string {
	return s.slider.label
}
func (s *sliderFloat64) Update(x, y, width, height, scale float32) {
	slider := &s.slider
	UpdateSlider(slider, x, y, width, height, scale)
}
func (s *sliderFloat64) Draw(image *ebiten.Image, whiteImage *ebiten.Image, top, left, width, height, scale float32) {
	slider := &s.slider
	DrawSliderShape(slider, image, whiteImage, top, left, width, height, scale)
}

// SliderFloat32

type sliderFloat32 struct {
	slider slider[float32]
}

func (g *GUI) AddSliderFloat32(label string, value, min, max float32, callback func(v float32)) {
	s := slider[float32]{
		label:        label,
		Interpolator: RangeInterpolator[float32]{Min: min, Max: max, Value: value},
		callback:     callback,
	}
	slider := &sliderFloat32{slider: s}
	g.components = append(g.components, slider)
}

func (s *sliderFloat32) Label() string {
	return s.slider.label
}
func (s *sliderFloat32) Update(x, y, width, height, scale float32) {
	slider := &s.slider
	UpdateSlider(slider, x, y, width, height, scale)
}
func (s *sliderFloat32) Draw(image *ebiten.Image, whiteImage *ebiten.Image, top, left, width, height, scale float32) {
	slider := &s.slider
	DrawSliderShape(slider, image, whiteImage, top, left, width, height, scale)
}

// SliderInt

type sliderInt struct {
	slider slider[int]
}

func (g *GUI) AddSliderInt(label string, value, min, max int, callback func(v int)) {
	s := slider[int]{
		label:        label,
		Interpolator: RangeInterpolator[int]{Min: min, Max: max, Value: value},
		callback:     callback,
	}
	slider := &sliderInt{slider: s}
	g.components = append(g.components, slider)
}

func (s *sliderInt) Label() string {
	return s.slider.label
}
func (s *sliderInt) Update(x, y, width, height, scale float32) {
	slider := &s.slider
	UpdateSlider(slider, x, y, width, height, scale)
}
func (s *sliderInt) Draw(image *ebiten.Image, whiteImage *ebiten.Image, top, left, width, height, scale float32) {
	slider := &s.slider
	DrawSliderShape(slider, image, whiteImage, top, left, width, height, scale)
}
