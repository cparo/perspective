package perspective

import (
	"image"
	"math"
)

type rollingStack struct {
	w  int             // Width of the visualization
	h  int             // Height of the visualization
	tA float64         // Lower limit of time range to be visualized
	tΩ float64         // Upper limit of time range to be visualized
	n  map[int16][]int // Event counts by status and x-axis position
	σ  []float64       // Event totals by and x-axis position
}

// NewRollingStack returns a rolling-stack-visualization generator.
func NewRollingStack(
	width int,
	height int,
	minTime int,
	maxTime int) Visualizer {

	return &rollingStack{
		width,
		height,
		float64(minTime),
		float64(maxTime),
		make(map[int16][]int),
		make([]float64, width)}
}

// Record accepts an EventDataPoint and plots it onto the visualization.
func (v *rollingStack) Record(e EventDataPoint) {
	for int(e.Status)+1 > len(v.n) {
		v.n[int16(len(v.n))] = make([]int, v.w)
	}
	w := float64(v.w)
	s := float64(e.Start)
	x := int(math.Min(w-1, w*(s-v.tA)/(v.tΩ-v.tA)))
	v.n[e.Status][x]++
	v.σ[x]++
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *rollingStack) Render() *image.RGBA {

	// Initialize our image canvas.
	vis := initializeVisualization(v.w, v.h)

	// Draw the rolling stack, giving each failure type a different color and
	// scaling the height of the visualization to normalize for the x-position
	// which represents the greatest density of recorded events.
	for x := 0; x < v.w; x++ {
		y := 0
		for i := 1; i < len(v.n); i++ {
			color := getErrorStackColor(i, len(v.n))
			if v.σ[x] > 0 {
				yʹ := y + int(float64(v.n[int16(i)][x]*v.h)/v.σ[x])
				for ; y < yʹ; y++ {
					vis.Set(x, v.h-y, color)
				}
			}
		}
	}

	return vis
}
