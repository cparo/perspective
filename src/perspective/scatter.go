package perspective

import (
	"image"
	"math"
)

type scatter struct {
	w      int         // Width of the visualization
	h      int         // Height of the visualization
	vis    *image.RGBA // Visualization canvas
	tA     float64     // Lower limit of time range to be visualized
	tΩ     float64     // Upper limit of time range to be visualized
	yLog2  float64     // Number of pixels over which elapsed times double
	colors float64     // Number of color steps before saturation
}

// NewScatter returns a scatter-visualization generator.
func NewScatter(
	width int,
	height int,
	minTime int,
	maxTime int,
	yLog2 float64,
	colorSteps int,
	xGrid int) *scatter {

	return (&scatter{
		width,
		height,
		initializeVisualization(width, height),
		float64(minTime),
		float64(maxTime),
		float64(yLog2),
		float64(colorSteps)}).drawGrid(xGrid)
}

// Record accepts an EventDataPoint and plots it onto the visualization.
func (s *scatter) Record(e EventDataPoint) {

	start := float64(e.Start)
	x := int(float64(s.w) * (start - s.tA) / (s.tΩ - s.tA))
	y := s.h - int(s.yLog2*math.Log2(float64(e.Run)))

	// Since recorded events may collide in space with other recorded points in
	// this visualization, we use a color progression to indicate the density
	// of events in a given pixel of the visualization. This requires that we
	// take into account the existing color of the point on the canvas to which
	// the event will be plotted and calculate its new color as a function of
	// its existing color.
	r16, g16, b16, _ := s.vis.At(x, y).RGBA()
	if e.Status == 0 {
		// We desturate success colors both for aesthetics and because this
		// allows them an additional range of visual differentiation (from
		// bright blue to white) beyond their normal clipping point in the blue
		// band.
		r16 = uint32(math.Min(maxC16, float64(r16)+maxC16/s.colors/4))
		g16 = uint32(math.Min(maxC16, float64(g16)+maxC16/s.colors/4))
		b16 = uint32(math.Min(maxC16, float64(b16)+maxC16/s.colors))
	} else {
		// Failures are not desaturated to help make them more visible and to
		// prevent a dense cluster of failures from looking like a dense cluster
		// of successes.
		r16 = uint32(math.Min(maxC16, float64(r16)+maxC16/s.colors))
	}
	plot(s.vis, x, y, r16, g16, b16)
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (s *scatter) Render() *image.RGBA {
	return s.vis
}

func (s *scatter) drawGrid(xGrid int) *scatter {

	// Draw vertical grid lines, if vertical divisions were specified
	if xGrid > 0 {
		for x := 0; x < s.w; x = x + s.w/xGrid {
			drawXGridLine(s.vis, x)
		}
	}

	// Draw horizontal grid lines on each doubling of the run time in seconds
	for y := s.h; y > 0; y = y - int(float64(s.h)/s.yLog2) {
		drawYGridLine(s.vis, y)
	}

	// Draw a line up top, for the sake of tidy appearance
	drawYGridLine(s.vis, 0)

	// Return the scatter visualization struct, so this can be conveniently
	// used in the visualization's constructor.
	return s
}
