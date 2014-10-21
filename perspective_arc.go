package main

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

func drawArcGrid(vis *image.RGBA) {

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

	for y := height / 2; y < height; y = y + int(float64(height)/yLog2) {
		for x := 0; x < width; x++ {
			vis.Set(x, y, gridColor)
			vis.Set(x, height-y, gridColor)
		}
	}
}

func generateArcVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	vis := initializeVisualization()
	drawArcGrid(vis)

	// Shorthand variables
	s := colorSteps
	h := height
	w := width

	for {

		tA := float64(minTime) // Pre-cast lower limit of time range
		tΩ := float64(maxTime) // Pre-cast upper limit of time range

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			tMin := float64(event.StartTime)
			tMax := float64(event.StartTime + event.RunTime)
			y := h / 2
			for t := tMin; t <= tMax; t++ {
				x := int(float64(w) * (t - tA) / (tΩ - tA))
				yMin := h/2 - int(yLog2*(math.Log2(math.Max(1, t-tMin))))
				for yʹ := y; yʹ > yMin; yʹ-- {
					y = yʹ
					if event.Status == 0 {
						r16, g16, b16, _ := vis.At(x, y).RGBA()
						r16 = uint32(min(65535, int(r16)+65536/s/4))
						g16 = uint32(min(65535, int(g16)+65536/s/4))
						b16 = uint32(min(65535, int(b16)+65536/s))
						vis.Set(
							x,
							y,
							color.RGBA{
								uint8(r16 >> 8),
								uint8(g16 >> 8),
								uint8(b16 >> 8),
								255})
					} else {
						r16, g16, b16, _ := vis.At(x, h-y).RGBA()
						r16 = uint32(min(65535, int(r16)+65536/s))
						vis.Set(
							x,
							h-y,
							color.RGBA{
								uint8(r16 >> 8),
								uint8(g16 >> 8),
								uint8(b16 >> 8),
								255})
					}
				}
			}
		}
	}

	png.Encode(oFile, vis)
}
