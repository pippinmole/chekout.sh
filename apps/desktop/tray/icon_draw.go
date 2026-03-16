package tray

import (
	"image"
	"image/color"
)

// drawLine draws a line between (x0,y0) and (x1,y1) using Bresenham's algorithm.
func drawLine(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	bounds := img.Bounds()
	for {
		if x0 >= bounds.Min.X && x0 < bounds.Max.X &&
			y0 >= bounds.Min.Y && y0 < bounds.Max.Y {
			img.SetNRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
