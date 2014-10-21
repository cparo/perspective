package perspective

import (
	"image"
	"math"
)

// TODO: Rename. This is a stupid name. Or at least a pretty vague one.
//       Leaning toward "sweep" at the moment.
type arc struct {
	w      int         // Width of the visualization
	h      int         // Height of the visualization
	vis    *image.RGBA // Visualization canvas
	tA     float64     // Lower limit of time range to be visualized
	tΩ     float64     // Upper limit of time range to be visualized
	yLog2  float64     // Number of pixels over which elapsed times double
	colors float64     // Number of color steps before saturation
}

// NewArc returns an arc-visualization generator.
func NewArc(
	width int,
	height int,
	minTime int,
	maxTime int,
	yLog2 float64,
	colorSteps int,
	xGrid int) *arc {

	return (&arc{
		width,
		height,
		initializeVisualization(width, height),
		float64(minTime),
		float64(maxTime),
		float64(yLog2),
		float64(colorSteps)}).drawGrid(xGrid)
}

// Record accepts an EventDataPoint and plots it onto the visualization.
func (v *arc) Record(e EventDataPoint) {
	tMin := float64(e.Start)
	tMax := float64(e.Start + e.Run)
	y := v.h / 2
	for t := tMin; t <= tMax; t++ {
		x := int(float64(v.w) * (t - v.tA) / (v.tΩ - v.tA))
		yMin := v.h/2 - int(v.yLog2*(math.Log2(math.Max(1, t-tMin))))
		for yʹ := y; yʹ > yMin; yʹ-- {
			y = yʹ
			if e.Status == 0 {
				r16, g16, b16, _ := v.vis.At(x, y).RGBA()
				r16 = uint32(math.Min(maxC16, float64(r16)+maxC16/v.colors/4))
				g16 = uint32(math.Min(maxC16, float64(g16)+maxC16/v.colors/4))
				b16 = uint32(math.Min(maxC16, float64(b16)+maxC16/v.colors))
				plot(v.vis, x, y, r16, g16, b16)
			} else {
				r16, g16, b16, _ := v.vis.At(x, v.h-y).RGBA()
				r16 = uint32(math.Min(maxC16, float64(r16)+maxC16/v.colors))
				plot(v.vis, x, v.h-y, r16, g16, b16)
			}
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *arc) Render() *image.RGBA {
	return v.vis
}

func (v *arc) drawGrid(xGrid int) *arc {

	// Draw vertical grid lines, if vertical divisions were specified
	if xGrid > 0 {
		for x := 0; x < v.w; x = x + v.w/xGrid {
			drawXGridLine(v.vis, x)
		}
	}

	// Draw horizontal grid lines on each doubling of the run time in seconds
	for y := v.h / 2; y < v.h; y = y + int(float64(v.h)/v.yLog2) {
		drawYGridLine(v.vis, y)
		drawYGridLine(v.vis, v.h - y)
	}

	// Draw a line up top, for the sake of tidy appearance
	drawYGridLine(v.vis, 0)

	// Return the arc visualization struct, so this can be conveniently
	// used in the visualization's constructor.
	return v
}
