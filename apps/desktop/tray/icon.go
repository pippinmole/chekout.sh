package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// iconData returns the bytes of a 22x22 PNG tray icon: a white checkmark on
// a transparent background, suitable for macOS menu bar and other platforms.
func iconData() []byte {
	const size = 22
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Draw a checkmark using Bresenham's line algorithm.
	// Left leg: from (3, 11) to (9, 17)
	drawLine(img, 3, 11, 9, 17, white)
	// Right leg: from (9, 17) to (19, 5)
	drawLine(img, 9, 17, 19, 5, white)

	// Thicken slightly by drawing adjacent lines.
	drawLine(img, 3, 12, 9, 18, white)
	drawLine(img, 9, 18, 19, 6, white)
	drawLine(img, 4, 11, 10, 17, white)
	drawLine(img, 10, 17, 19, 7, white)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

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
