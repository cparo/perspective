// Perspective: Graphing library for quality control in event-driven systems

// Copyright (C) 2014 Christian Paro <christian.paro@gmail.com>,
//                                   <cparo@digitalocean.com>

// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License version 2 as published by the
// Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.

// You should have received a copy of the GNU General Public License along with
// this program. If not, see <http://www.gnu.org/licenses/>.

package perspective

import (
	"image"
	"math"
)

type scatter struct {
	w     int         // Width of the visualization
	h     int         // Height of the visualization
	vis   *image.RGBA // Visualization canvas
	tA    float64     // Lower limit of time range to be visualized
	tτ    float64     // Length of time range to be visualized
	yLog2 float64     // Number of pixels over which elapsed times double
	cΔ    float64     // Increment for color channel value increases
}

// NewScatter returns a scatter-visualization generator.
func NewScatter(
	width int,
	height int,
	minTime int,
	maxTime int,
	yLog2 float64,
	colorSteps int,
	xGrid int) Visualizer {

	return (&scatter{
		width,
		height,
		initializeVisualization(width, height),
		float64(minTime),
		float64(maxTime - minTime),
		float64(yLog2),
		saturated / float64(colorSteps)}).drawGrid(xGrid)
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *scatter) Record(e *EventData) {

	x := int(float64(v.w) * (float64(e.Start) - v.tA) / v.tτ)
	y := v.h - int(v.yLog2*math.Log2(float64(e.Run)))

	// Since recorded events may collide in space with other recorded points in
	// this visualization, we use a color progression to indicate the density
	// of events in a given pixel of the visualization. This requires that we
	// take into account the existing color of the point on the canvas to which
	// the event will be plotted and calculate its new color as a function of
	// its existing color.
	c := getRGBA(v.vis, x, y)
	if e.Status == 0 {
		// We desturate success colors both for aesthetics and because this
		// allows them an additional range of visual differentiation (from
		// bright blue to white) beyond their normal clipping point in the blue
		// band.
		c.R = uint8(math.Min(saturated, float64(c.R)+v.cΔ/4))
		c.G = uint8(math.Min(saturated, float64(c.G)+v.cΔ/4))
		c.B = uint8(math.Min(saturated, float64(c.B)+v.cΔ))
	} else if e.Status > 0 {
		// Failures are not desaturated to help make them more visible and to
		// prevent a dense cluster of failures from looking like a dense cluster
		// of successes.
		c.R = uint8(math.Min(saturated, float64(c.R)+v.cΔ))
	} else {
		// In-progress events are shown as grayscale points capping out at a
		// light-mid gray to avoid confusion with a high density of successful
		// events, unless the point is already beyond that intensity on one or
		// more channels due to other recorded events.
		c.R = uint8(math.Max(float64(c.R), math.Min(196, float64(c.R)+v.cΔ)))
		c.G = uint8(math.Max(float64(c.G), math.Min(196, float64(c.G)+v.cΔ)))
		c.B = uint8(math.Max(float64(c.B), math.Min(196, float64(c.B)+v.cΔ)))
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *scatter) Render() image.Image {
	return v.vis
}

func (v *scatter) drawGrid(xGrid int) *scatter {

	// Draw vertical grid lines, if vertical divisions were specified.
	if xGrid > 0 {
		for x := 0; x < v.w; x += v.w / xGrid {
			drawXGridLine(v.vis, x)
		}
	}

	// Draw horizontal grid lines on each doubling of the run time in seconds.
	for y := v.h; y > 0; y -= int(float64(v.h) / v.yLog2) {
		drawYGridLine(v.vis, y)
	}

	// Draw a line along the top and bottom, for the sake of tidy appearance.
	drawYGridLine(v.vis, 0)
	drawYGridLine(v.vis, v.h-1)

	// Return the scatter visualization struct, so this can be conveniently
	// used in the visualization's constructor.
	return v
}
