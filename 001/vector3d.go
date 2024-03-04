package main

import (
	"fmt"
	"math"
	"math/rand"
)

type Vec3D struct {
	x, y, z float64
}

func NewVec3D(x, y, z float64) Vec3D {
	return Vec3D{x, y, z}
}

func (v Vec3D) Normalize() Vec3D {
	magnitude := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	return Vec3D{v.x / magnitude, v.y / magnitude, v.z / magnitude}
}

func (v Vec3D) Add(value Vec3D) Vec3D {
	return Vec3D{v.x + value.x, v.y + value.y, v.z + value.z}
}

func (v Vec3D) Sub(value Vec3D) Vec3D {
	return Vec3D{v.x - value.x, v.y - value.y, v.z - value.z}
}

func (v Vec3D) Cross(value Vec3D) Vec3D {
	return Vec3D{
		v.y*value.z - v.z*value.y,
		v.z*value.x - v.x*value.z,
		v.x*value.y - v.y*value.x,
	}
}

func (v Vec3D) Scale(value float64) Vec3D {
	return Vec3D{v.x * value, v.y * value, v.z * value}
}

func (v *Vec3D) Randomize() {
	v.x = rand.Float64()
	v.y = rand.Float64()
	v.z = rand.Float64()
}

func (v *Vec3D) Set(vx, vy, vz float64) {
	v.x, v.y, v.z = vx, vy, vz
}

func (v *Vec3D) ScaleSelf(value float64) {
	v.x *= value
	v.y *= value
	v.z *= value
}

func (v *Vec3D) AddSelf(value Vec3D) {
	v.x += value.x
	v.y += value.y
	v.z += value.z
}

func (v *Vec3D) InterpolateToSelf(value Vec3D, scale float64) {
	v.x = value.x * scale
	v.y = value.y * scale
	v.z = value.z * scale
}

func (v Vec3D) Clone() Vec3D {
	return Vec3D{v.x, v.y, v.z}
}

func (v Vec3D) String() string {
	return fmt.Sprintf("[Vec3D] x=%f, y=%f, z=%f", v.x, v.y, v.z)
}
