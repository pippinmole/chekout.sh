//go:build !windows

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
