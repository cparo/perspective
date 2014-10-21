package perspective

import (
	"image"
	"math"
)

type statusStack struct {
	w int           // Width of the visualization
	h int           // Height of the visualization
	n map[int16]int // Event counts by exit status code
	σ float64       // Total count of recorded events
}

// NewStatusStack returns a status-stack-visualization generator.
func NewStatusStack(width int, height int) *statusStack {
	return &statusStack{width, height, make(map[int16]int), 0}
}

// Record accepts an EventDataPoint and plots it onto the visualization.
func (v *statusStack) Record(e EventDataPoint) {
	for int(e.Status)+1 > len(v.n) {
		v.n[int16(len(v.n))] = 0
	}
	v.n[e.Status]++
	v.σ++
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *statusStack) Render() *image.RGBA {

	// Initialize our image canvas.
	vis := initializeVisualization(v.w, v.h)

	// Draw the stack, giving each failure type a different color and scaling
	// the overall stack stuck that each failure case occupies space
	// proportionate to its relative frequency as observed in across all
	// recoded events.
	y := 0
	for i := 1; i <= len(v.n); i++ {
		color := getErrorStackColor(i, len(v.n))
		yʹ := y + int(math.Ceil(float64(v.n[int16(i)]*v.h)/v.σ))
		for ; y < yʹ; y++ {
			for x := 0; x < v.w; x++ {
				vis.Set(x, v.h-y, color)
			}
		}
	}

	return vis
}
