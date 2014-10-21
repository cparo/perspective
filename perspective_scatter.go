package main

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"perspective"
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

	scatter := perspective.NewScatter(
		width,
		height,
		minTime,
		maxTime,
		yLog2,
		colorSteps,
		xGrid)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			scatter.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, scatter.Render())
}
