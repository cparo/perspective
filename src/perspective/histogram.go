package perspective

import (
	"image"
	"image/color"
	"math"
)

type histogram struct {
	w     int     // Width of the visualization
	h     int     // Height of the visualization
	yLog2 float64 // Number of pixels over which elapsed times double
	pass  []int   // Counts of successful events by x-axis position
	fail  []int   // Counts of failed events by x-axis position
}

// NewHistogram returns a histogram-visualization generator.
func NewHistogram(width int, height int, yLog2 float64) *histogram {
	return &histogram{
		width,
		height,
		yLog2,
		make([]int, width),
		make([]int, width)}
}

// Record accepts an EventDataPoint and plots it onto the visualization.
func (v *histogram) Record(e EventDataPoint) {

	// Run time is hacked to a floor of 1 because a log of zero doesn't
	// make a lot of sense, and there are some fun cases of events with
	// negative recorded run times because of clock skew.
	x := int(v.yLog2 * math.Log2(math.Max(1, float64(e.Run))))

	// Discard data which lies beyond the specified bounds for the
	// rendered visualization.
	if x < v.w {
		if e.Status == 0 {
			v.pass[x] = v.pass[x] + 1
		} else {
			v.fail[x] = v.fail[x] + 1
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *histogram) Render() *image.RGBA {

	// Initialize our image canvas and grid.
	vis := initializeVisualization(v.w, v.h)
	v.drawGrid(vis)

	// Find the highest point of the histogram to normalize the height of the
	// masts.
	maxCount := 0
	for x := 0; x < v.w; x++ {
		maxCount = max(maxCount, v.pass[x]+v.fail[x])
	}
	scale := float64(v.h) / float64(maxCount)

	// Set up our pass and fail colors.
	passColor := color.RGBA{83, 83, 191, 255}
	failColor := color.RGBA{191, 33, 33, 255}

	// Draw the masts, with successes stacked atop failures.
	for x := 0; x < v.w; x++ {
		fail := int(math.Ceil(float64(v.fail[x]) * scale))
		pass := int(math.Ceil(float64(v.pass[x]) * scale))
		for y := 0; y < fail; y++ {
			vis.Set(x, v.h-y, failColor)
		}
		for y := fail; y < fail+pass; y++ {
			vis.Set(x, v.h-y, passColor)
		}
	}

	return vis
}

func (v *histogram) drawGrid(vis *image.RGBA) {

	// Draw vertical grid lines on each doubling of the run time in seconds.
	for x := 0; x < v.w; x += int(float64(v.w) / v.yLog2) {
		drawXGridLine(vis, x)
	}

	// Draw lines bounding the reset of the graph, for the sake of a tidy
	// appearance. We don't need to draw a line on the left edge here, since
	// we would have done that already for the grid.
	drawXGridLine(vis, v.w)
	drawYGridLine(vis, 0)
	drawYGridLine(vis, v.h)
}
