package main

import (
	"bufio"
	"encoding/binary"
	"image/png"
	"math"
)

func generateStatusStackVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	n := make(map[int16]int) // Event counts by exit status code
	σ := float64(0)          // Total count of events

	for {

		var event eventData

		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			for int(event.Status)+1 > len(n) {
				n[int16(len(n))] = 0
			}
			n[event.Status] = n[event.Status] + 1
			σ++
		}
	}

	vis := initializeVisualization()

	y := 0
	for i := 1; i <= len(n); i++ {
		color := getErrorStackColor(i, len(n))
		yʹ := y + int(math.Ceil(float64(n[int16(i)]*height)/σ))
		for ; y < yʹ; y++ {
			for x := 0; x < width; x++ {
				vis.Set(x, height-y, color)
			}
		}
	}

	png.Encode(oFile, vis)
}
