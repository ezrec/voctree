//
// Copyright (c) 2020, Jason S. McMullan <jason.mcmullan@gmail.com>
//

package voctree

import (
	"testing"

	"fmt"
	"image"
	"image/color"
)

func TestVoctreeCreate(t *testing.T) {
	const testX = 128
	const testY = 128
	const testZ = 128

	v := NewVoctree(testX, testY)

	if v.Bounds() != image.Rect(0, 0, testX, testY) {
		t.Fatalf("expected %v, got %v", image.Rect(0, 0, testX, testY), v.Bounds())
	}

	// Insert 1024 black images (even power of 2)
	black := image.NewGray(image.Rect(0, 0, testX, testY))

	if black.Stride != testX {
		t.Fatalf("expected %v, got %v", testX, black.Stride)
	}

	for z := 0; z < testZ; z++ {
		v.SetPlane(z, black)
	}

	// This should only take one v element
	if v.Nodes() != 1 {
		fmt.Printf("%v\n", v)
		t.Fatalf("expected %v, got %v", 1, v.Nodes())
	}

	// Set the bottom half of the v to all white
	white := image.NewGray(image.Rect(0, 0, testX, testY))

	for l := 0; l < len(white.Pix); l++ {
		white.Pix[l] = 0xff
	}

	for z := 0; z < testZ/2; z++ {
		v.SetPlane(z, white)
	}

	// Verify that there are now 8 nodes
	if v.Nodes() != 8 {
		fmt.Printf("%v\n", v)
		t.Fatalf("expected %v, got %v", 8, v.Nodes())
	}

	// Set the top half of the v to all white
	for z := testZ / 2; z < testZ; z++ {
		v.SetPlane(z, white)
	}

	// Verify that there are now 8 nodes
	if v.Nodes() != 1 {
		fmt.Printf("%v\n", v)
		t.Fatalf("expected %v, got %v", 1, v.Nodes())
	}
}

func TestVoctreeModify(t *testing.T) {
	const testX = 128
	const testY = 128
	const testZ = 128

	v := NewVoctree(testX, testY)

	// V starts as all black. Insert a horizontal bar of y pixels on the middle plane
	z := testZ / 2
	y := testY / 2
	for x := 0; x < v.Dx(); x++ {
		v.Set(Point{X: uint16(x), Y: uint16(y), Z: uint16(z)}, color.Gray{0xff})
	}

	// Collect the image
	plane := v.GetPlane(z)

	// Check the image
	for py := 0; py < testY; py++ {
		for px := 0; px < testX; px++ {
			var expected uint8
			if py == y {
				expected = 0xff
			}
			c := plane.At(px, py).(color.Gray).Y
			if c != expected {
				t.Errorf("%v, %v: expected %+#v, got %+#v", px, py, expected, c)
			}
		}
	}
}
