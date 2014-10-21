package main

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

func drawScatterGrid(vis *image.RGBA) {

	gridColor := color.RGBA{45, 45, 45, 255}

	for x := 0; x < width; x++ {
		vis.Set(x, 0, gridColor)
		vis.Set(x, height-1, gridColor)
	}

	if xGrid > 0 {
		for x := 0; x < width; x = x + width/xGrid {
			for y := 0; y < height; y++ {
				vis.Set(x, y, gridColor)
			}
		}
	}

	for y := 0; y < height; y = y + int(float64(height)/yLog2) {
		for x := 0; x < width; x++ {
			vis.Set(x, height-y, gridColor)
		}
	}
}

func generateScatterVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	vis := initializeVisualization()
	drawScatterGrid(vis)

	for {

		tA := float64(minTime) // Pre-cast lower limit of time range
		tΩ := float64(maxTime) // Pre-cast upper limit of time range

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			start := float64(event.StartTime)
			x := int(float64(width) * (start - tA) / (tΩ - tA))
			y := height - int(yLog2*math.Log2(float64(event.RunTime)))
			r16, g16, b16, _ := vis.At(x, y).RGBA()
			if event.Status == 0 {
				// We desturate success colors in part for aesthetics and in
				// part because this allows them an additional range of visual
				// differentiation (from bright blue to white) beyond their
				// normal clipping point in the blue band.
				r16 = uint32(min(65535, int(r16)+65536/colorSteps/4))
				g16 = uint32(min(65535, int(g16)+65536/colorSteps/4))
				b16 = uint32(min(65535, int(b16)+65536/colorSteps))
			} else {
				// Errors are not desaturated to help make them more visible
				// and to prevent a dense cluster of errors from looking like
				// a dense cluster of successes.
				r16 = uint32(min(65535, int(r16)+65536/colorSteps))
			}
			r8 := uint8(r16 >> 8)
			g8 := uint8(g16 >> 8)
			b8 := uint8(b16 >> 8)
			vis.Set(x, y, color.RGBA{r8, g8, b8, 255})
		}
	}

	png.Encode(oFile, vis)
}
