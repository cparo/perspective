package main

import (
	"bufio"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

func drawHistogramGrid(vis *image.RGBA) {

	gridColor := color.RGBA{45, 45, 45, 255}

	for x := 0; x < width; x++ {
		vis.Set(x, 0, gridColor)
		vis.Set(x, height-1, gridColor)
	}

	for y := 0; y < height; y++ {
		vis.Set(0, y, gridColor)
		vis.Set(width-1, y, gridColor)
	}

	for i := float64(0); int(i*yLog2) < width; i++ {
		x := int(i * yLog2)
		for y := 0; y < height; y++ {
			vis.Set(x, y, gridColor)
		}
	}
}

func generateHistogramVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	pass := make([]int, width)
	fail := make([]int, width)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {

			// Run time is hacked to a floor of 1 because a log of zero doesn't
			// make a lot of sense, and there are some fun cases of events with
			// negative recorded run times because of clock skew.
			runTime := float64(max(1, int(event.RunTime)))
			x := int(yLog2 * math.Log2(runTime))

			// Discard data which lies beyond the specified bounds for the
			// rendered visualization.
			if x < width {
				if event.Status == 0 {
					pass[x] = pass[x] + 1
				} else {
					fail[x] = fail[x] + 1
				}
			}
		}
	}

	vis := initializeVisualization()
	drawHistogramGrid(vis)

	maxCount := 0
	for x := 0; x < width; x++ {
		maxCount = max(maxCount, pass[x]+fail[x])
	}
	scale := float64(height) / float64(maxCount)
	passColor := color.RGBA{83, 83, 191, 255}
	failColor := color.RGBA{191, 33, 33, 255}
	for x := 0; x < width; x++ {
		fail := int(math.Ceil(float64(fail[x]) * scale))
		pass := int(math.Ceil(float64(pass[x]) * scale))
		for y := 0; y < fail; y++ {
			vis.Set(x, height-y, failColor)
		}
		for y := fail; y < fail+pass; y++ {
			vis.Set(x, height-y, passColor)
		}
	}

	png.Encode(oFile, vis)
}
