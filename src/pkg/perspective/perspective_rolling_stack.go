package main

import (
	"bufio"
	"encoding/binary"
	"image/png"
)

func generateRollingStackVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	var event eventData

	tA := float64(minTime)      // Pre-cast lower limit of time range
	tΩ := float64(maxTime)      // Pre-cast upper limit of time range
	n := make(map[int16][]int)  // Event counts by status and start time
	σ := make([]float64, width) // Event totals by start time

	for {

		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			for int(event.Status)+1 > len(n) {
				n[int16(len(n))] = make([]int, width)
			}
			start := float64(event.StartTime)
			x := min(width-1, int(float64(width)*(start-tA)/(tΩ-tA)))
			n[event.Status][x] = n[event.Status][x] + 1
			σ[x]++
		}
	}

	vis := initializeVisualization()

	for x := 0; x < width; x++ {
		y := 0
		for i := 1; i < len(n); i++ {
			color := getErrorStackColor(i, len(n))
			if σ[x] > 0 {
				yʹ := y + int(float64(n[int16(i)][x]*height)/σ[x])
				for ; y < yʹ; y++ {
					vis.Set(x, height-y, color)
				}
			}
		}
	}

	png.Encode(oFile, vis)
}
