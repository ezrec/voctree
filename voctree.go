//
// Copyright (c) 2020, Jason S. McMullan <jason.mcmullan@gmail.com>
//

package voctree

import (
	"fmt"
	"image"
	"image/color"
)

type Point struct {
	X, Y, Z uint16
}

type Cube struct {
	Point
	SideShift uint8
}

func octIndex(cube Cube) (index int, here Cube) {
	side := uint16(1 << (cube.SideShift - 1))
	p := cube.Point

	isx := (p.X >> (cube.SideShift - 1)) > 0
	isy := (p.Y >> (cube.SideShift - 1)) > 0
	isz := (p.Z >> (cube.SideShift - 1)) > 0

	index = 0
	if isx {
		index |= (1 << 0)
	}
	if isy {
		index |= (1 << 1)
	}
	if isz {
		index |= (1 << 2)
	}

	here = Cube{
		SideShift: cube.SideShift - 1,
		Point: Point{
			X: p.X & (side - 1),
			Y: p.Y & (side - 1),
			Z: p.Z & (side - 1),
		},
	}

	return
}

type Vocelish interface {
	Set(cube Cube, pixel color.Gray) (vnew Vocelish)
	At(cube Cube) (pixel color.Gray)
	Nodes() (count int)
	String(cube Cube) (str string)
}

// Leaf case: one pixel in the cube, no subtress
type Vocel1 struct {
	pixel color.Gray // Color of the voxel
}

func (v1 *Vocel1) Nodes() (count int) {
	return 1
}

func (v1 *Vocel1) String(cube Cube) (str string) {
	return fmt.Sprintf("v1{size: %v @(%+v), pixel: %+v}", (1 << cube.SideShift), cube.Point, v1.pixel)
}

/// Set - set the pixel of the volume
///
/// If the new pixel does not match the old pixel, split into 8 voxels
func (v1 *Vocel1) Set(cube Cube, pixel color.Gray) (vnew Vocelish) {
	if cube.SideShift == 0 || pixel == v1.pixel {
		v1.pixel = pixel
		vnew = v1
		return
	}

	// Split!
	v8 := &Vocel8{}
	for i := 0; i < len(v8.pixel); i++ {
		v8.pixel[i] = v1.pixel
	}

	vnew = v8.Set(cube, pixel)
	return
}

func (v1 *Vocel1) At(cube Cube) (pixel color.Gray) {
	pixel = v1.pixel
	return
}

// Leaf case: 8 pixels in the cube, no subtrees
type Vocel8 struct {
	pixel [8]color.Gray
}

func (v8 *Vocel8) Nodes() (count int) {
	return 8
}

func (v8 *Vocel8) String(cube Cube) (str string) {
	return fmt.Sprintf("v8{size: %v @(%+v), pixels: %+v}", (1 << cube.SideShift), cube.Point, v8.pixel)
}

/// Set - set the pixel of the volume
///
func (v8 *Vocel8) Set(cube Cube, pixel color.Gray) (vnew Vocelish) {
	index, _ := octIndex(cube)

	if v8.pixel[index] == pixel {
		vnew = v8
		return
	}

	// At the pixel level...
	if cube.SideShift == 1 {
		allSame := true
		for i, c := range v8.pixel {
			if index != i && c != pixel {
				allSame = false
				break
			}
		}

		// All the pixels would not be the same
		if !allSame {
			v8.pixel[index] = pixel
			vnew = v8
			return
		}

		// Coalesce into a Vocel1
		v1 := &Vocel1{pixel: pixel}
		vnew = v1
		return
	}

	// Split!
	vt := &VocelTree{}
	for i, c := range v8.pixel {
		vt.subtree[i] = &Vocel1{pixel: c}
	}

	vnew = vt.Set(cube, pixel)
	return
}

func (v8 *Vocel8) At(cube Cube) (pixel color.Gray) {
	index, _ := octIndex(cube)
	pixel = v8.pixel[index]
	return
}

// Tree case: subtrees
type VocelTree struct {
	subtree [8]Vocelish
}

func (vt *VocelTree) Nodes() (count int) {
	sum := 0
	for _, sub := range vt.subtree {
		sum += sub.Nodes()
	}

	return 1 + sum
}

func (vt *VocelTree) String(cube Cube) (str string) {
	var substr string

	side := 1 << (cube.SideShift - 1)
	here := Cube{SideShift: cube.SideShift - 1}
	for n, sub := range vt.subtree {
		cube.Point.X = uint16(((n >> 0) & 1) * side)
		cube.Point.Y = uint16(((n >> 1) & 1) * side)
		cube.Point.Z = uint16(((n >> 2) & 1) * side)
		substr += "    " + sub.String(here) + ",\n"
	}
	return fmt.Sprintf("vt{cube: %v, sub:\n%v}\n", cube, substr)
}

func (vt *VocelTree) Set(cube Cube, pixel color.Gray) (vnew Vocelish) {

	index, here := octIndex(cube)

	vt.subtree[index] = vt.subtree[index].Set(here, pixel)

	// If all the subtrees are Vocel1, the coalesce into a Vocel8
	allVocel1 := true
	for _, sub := range vt.subtree {
		_, ok := sub.(*Vocel1)
		if !ok {
			allVocel1 = false
			break
		}
	}

	if !allVocel1 {
		vnew = vt
		return
	}

	// Collect the pixels of all the subnodes
	allSame := true
	var pixels [8]color.Gray
	for i, sub := range vt.subtree {
		v1 := sub.(*Vocel1)
		pixels[i] = v1.pixel
		if i > 0 && pixels[i] != pixels[0] {
			allSame = false
		}
		vt.subtree[i] = nil
	}

	// If all the pixels are the same, use a Vocel1, otherwise use a Vocel8
	if !allSame {
		v8 := &Vocel8{pixel: pixels}
		vnew = v8
	} else {
		v1 := &Vocel1{pixel: pixels[0]}
		vnew = v1
	}

	return
}

func (vt *VocelTree) At(cube Cube) (pixel color.Gray) {
	index, here := octIndex(cube)

	pixel = vt.subtree[index].At(here)
	return
}

// Voxel Octree
type Voctree struct {
	Vocelish
	image.Rectangle
	SideShift uint8 // Power-of-8 size (0 = 1 voxel/side, 1 = 2 voxels/side, 2 = 4 voxels/side, ...)
}

func NewVoctree(sizex, sizey int) (v *Voctree) {
	sideShift := 0

	for (1 << sideShift) < sizex {
		sideShift++
	}

	for (1 << sideShift) < sizey {
		sideShift++
	}

	v = &Voctree{
		Rectangle: image.Rect(0, 0, sizex, sizey),
		Vocelish:  &Vocel1{},
		SideShift: uint8(sideShift),
	}

	return
}

func (v *Voctree) resizeSideShift(z int) {
	// Crank up the shift size if needed for extra Z
	// Create a new subtree, and put the existing root in the 0,0,0 corner
	for (1 << v.SideShift) < z {
		v.SideShift++
		vt := &VocelTree{}
		vt.subtree[0] = v.Vocelish
		for i := 1; i < len(vt.subtree); i++ {
			vt.subtree[i] = &Vocel1{}
		}
		v.Vocelish = vt
	}
}

func (v *Voctree) Set(point Point, pixel color.Gray) {
	v.Vocelish = v.Vocelish.Set(Cube{SideShift: v.SideShift, Point: point}, pixel)
}

func (v *Voctree) At(point Point) (pixel color.Gray) {
	pixel = v.Vocelish.At(Cube{SideShift: v.SideShift, Point: point})
	return
}

func (v *Voctree) SetPlane(z int, gray *image.Gray) (err error) {
	size := v.Size()

	pix := gray.Pix

	if len(pix) != size.X*size.Y {
		err = fmt.Errorf("pix: expected %v bytes, got %v", size.X*size.Y, len(pix))
		return
	}

	v.resizeSideShift(z)

	cube := Cube{SideShift: v.SideShift, Point: Point{Z: uint16(z)}}
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			cube.Point.X = uint16(x)
			cube.Point.Y = uint16(y)
			v.Vocelish = v.Vocelish.Set(cube, color.Gray{Y: pix[y*size.X+x]})
		}
	}

	return
}

func (v *Voctree) GetPlane(z int) (gray *image.Gray) {
	size := v.Size()

	v.resizeSideShift(z)

	pix := make([]byte, size.X*size.Y)

	cube := Cube{SideShift: v.SideShift, Point: Point{Z: uint16(z)}}
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			cube.Point.X = uint16(x)
			cube.Point.Y = uint16(y)
			pix[y*size.X+x] = v.Vocelish.At(cube).Y
		}
	}

	gray = &image.Gray{
		Rect:   v.Rectangle,
		Stride: size.X,
		Pix:    pix,
	}

	return
}

func (v *Voctree) String() (str string) {
	str = v.Rectangle.String() + v.Vocelish.String(Cube{SideShift: v.SideShift})
	return
}
