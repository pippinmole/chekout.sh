//go:build windows

package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// iconData returns ICO-format bytes for the Windows system tray.
// Windows requires ICO format; modern ICO files can embed PNG data directly.
func iconData() []byte {
	const size = 32
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Scale checkmark to 32x32
	drawLine(img, 4, 16, 13, 25, white)
	drawLine(img, 13, 25, 28, 7, white)
	drawLine(img, 4, 17, 13, 26, white)
	drawLine(img, 13, 26, 28, 8, white)
	drawLine(img, 5, 16, 14, 25, white)
	drawLine(img, 14, 25, 28, 9, white)

	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	// Wrap PNG in ICO container (Vista+ supports PNG inside ICO).
	var ico bytes.Buffer
	// ICONDIR
	ico.Write([]byte{0, 0})                            // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1)) // type: 1 = ICO
	binary.Write(&ico, binary.LittleEndian, uint16(1)) // count: 1 image
	// ICONDIRENTRY (16 bytes)
	ico.WriteByte(byte(size))                                     // width
	ico.WriteByte(byte(size))                                     // height
	ico.WriteByte(0)                                              // color count (0 = >8bpp)
	ico.WriteByte(0)                                              // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1))            // planes
	binary.Write(&ico, binary.LittleEndian, uint16(32))           // bit count
	binary.Write(&ico, binary.LittleEndian, uint32(len(pngData))) // bytes in resource
	binary.Write(&ico, binary.LittleEndian, uint32(6+16))         // image offset (after ICONDIR + 1 ICONDIRENTRY)
	// PNG payload
	ico.Write(pngData)
	return ico.Bytes()
}
