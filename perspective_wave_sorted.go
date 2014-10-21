package main

import (
	"bufio"
	"encoding/binary"
	"image/color"
	"image/png"
	"sort"
)

func generateSortedWaveVisualization(iPath string, oPath string) {

	// NOTE: Event input is expected to be received in chronological order. If
	//       it is not received in chronological order, the graph will not be
	//       rendered properly (with the severity of the issue being dependent
	//       upon the degree of deviation between the input order and the ideal
	//       chronologically-sorted input.

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	vis := initializeVisualization()

	// Shorthand variables
	h := height
	w := width

	var (
		pass  = []eventData{}
		fail  = []eventData{}
		passʹ []eventData
		failʹ []eventData
	)

	x := 0
	for {

		tA := float64(minTime) // Pre-cast lower limit of time range
		tΩ := float64(maxTime) // Pre-cast upper limit of time range

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			passʹ = make([]eventData, 0, len(pass)+64)
			failʹ = make([]eventData, 0, len(fail)+64)
			for _, e := range pass {
				if e.StartTime+e.RunTime > event.StartTime {
					passʹ = append(passʹ, e)
				}
			}
			for _, e := range fail {
				if e.StartTime+e.RunTime > event.StartTime {
					failʹ = append(failʹ, e)
				}
			}
			pass = passʹ
			fail = failʹ
			if event.Status == 0 {
				pass = append(passʹ, event)
			} else {
				fail = append(failʹ, event)
			}
			t := float64(event.StartTime)
			xʹ := int(float64(w) * (t - tA) / (tΩ - tA))
			var points []float64
			for xʹ > x {
				x++
				yP := 0
				yF := 0
				points = make([]float64, 0, len(pass))
				for i := 0; i < len(pass); i++ {
					e := pass[len(pass)-i-1]
					start := e.StartTime
					run := e.RunTime
					prog := float64(event.StartTime-start) / float64(run+1)
					points = append(points, prog)
				}
				sort.Sort(sort.Float64Slice(points))
				for _, prog := range points {
					r := uint8(min(255, int(33+255*prog/4)))
					g := uint8(min(255, int(33+255*prog/4)))
					b := uint8(min(255, int(33+255*prog)))
					yPʹ := yP + 1
					for ; yP < yPʹ; yP++ {
						vis.Set(
							x,
							h/2-yP,
							color.RGBA{r, g, b, 255})
					}
				}
				points = make([]float64, 0, len(fail))
				for i := 0; i < len(fail); i++ {
					e := fail[len(fail)-i-1]
					start := e.StartTime
					run := e.RunTime
					prog := float64(event.StartTime-start) / float64(run+1)
					points = append(points, prog)
				}
				sort.Sort(sort.Float64Slice(points))
				for _, prog := range points {
					r := uint8(min(255, int(33+255*prog)))
					g := uint8(min(255, int(33+255*prog/4)))
					b := uint8(min(255, int(33+255*prog/4)))
					yFʹ := yF + 1
					for ; yF < yFʹ; yF++ {
						vis.Set(
							x,
							h/2+yF,
							color.RGBA{r, g, b, 255})
					}
				}
			}
		}
	}
	png.Encode(oFile, vis)
}
