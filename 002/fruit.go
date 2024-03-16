package main

const (
	APPLE = iota
)

type Fruit struct {
	X      float64
	Y      float64
	VX     float64
	VY     float64
	Radius float64
	Type   int
	Remove bool
}

func NewApple(x float64, y float64) *Fruit {
	return &Fruit{
		X:      x,
		Y:      y,
		Radius: 20,
		Type:   APPLE,
	}
}
