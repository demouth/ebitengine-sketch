package main

const (
	BOTTOM = iota
	RIGHT
	TOP
	LEFT
)

type Fruit struct {
	X      float64
	Y      float64
	VX     float64
	VY     float64
	Radius float64

	Direction     uint8
	TotalMovement float64
}

func NewApple(x float64, y float64) *Fruit {
	return &Fruit{
		X:      x,
		Y:      y,
		Radius: 20,

		Direction: BOTTOM,
	}
}
