// Perspective: Graphing library for quality control in event-driven systems

// Copyright (C) 2015 Christian Paro <christian.paro@gmail.com>,
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

type runTimeLine struct {
	w         int     // Width of the visualization
	h         int     // Height of the visualization
	tA        float64 // Lower limit of time range to be visualized
	tτ        float64 // Length of time range to be visualized
	yLog2     float64 // Number of pixels over which elapsed times double
	nS        []int   // Counts of successful events by x-axis position
	nF        []int   // Counts of failed events by x-axis position
	nA        []int   // Counts of active events by x-axis position
	t         []int   // Sums of run-times of events by x-position
	xGrid     int     // Number of vertical grid divisions
	bg        int     // Background grey level
}

// NewRunTimeLine returns an line-graph event-run-time-visualization generator.
func NewRunTimeLine(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int,
	yLog2 float64,
	xGrid int) Visualizer {

	return &runTimeLine{
		width,
		height,
		float64(minTime),
		float64(maxTime - minTime),
		yLog2,
		make([]int, width),
		make([]int, width),
		make([]int, width),
		make([]int, width),
		xGrid,
		bg}
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *runTimeLine) Record(e *EventData) {

	// Position on the x-axis corresponds to the event's start time.
	x := int(float64(v.w) * (float64(e.Start) - v.tA) / v.tτ)

	// Update count and aggregate run-time values for appropriate x-position.
	if e.Status == 0 {
		v.nS[x]++
	} else if e.Status > 0 {
		v.nF[x]++
	} else {
		v.nA[x]++
	}
	v.t[x] = v.t[x] + int(e.Run)
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *runTimeLine) Render() image.Image {

	// Stroke width (for visibility and calligraphic effect)
	stroke := v.h / 48

	// Initialize our image canvas and grid.
	vis := initializeVisualization(v.w, v.h, v.bg)
	v.drawGrid(vis)

	// Draw the lines.
	xLast, yLast := 0, 0;
	for x := 0; x < v.w; x++ {

		// We only calculate logs on source time values which exceed 1 in order
		// to put a floor value of zero on the output value.
		y := 0
		n := v.nS[x] + v.nF[x] + v.nA[x]
		this := float64(v.t[x])/math.Max(float64(n), 1)
		if this > 1 { y = int(v.yLog2*math.Log2(this)) }

		// Color line according to relative quantities of completed, failed, and
		// successful events recorded at during the time range corresponding to
		// this x-position.
		if n > 0 {

			// Flatline data from beginning of graph up to first data point, and
			// make line dotted until real data is available.
			xIncrement := 1
			if xLast == 0 {
				yLast = y
				xIncrement = 4
			}

			r := uint8(32 + 128 * v.nF[x] / n)
			g := uint8(32 + 128 * v.nA[x] / n)
			b := uint8(32 + 128 * v.nS[x] / n)

			for xPos := xLast; xPos < x; xPos += xIncrement {
				var yMin, yMax int
				yA := yLast + (y - yLast) * (xPos - xLast) / (x - xLast)
				yB := yLast + (y - yLast) * (xPos + 1 - xLast) / (x - xLast)
				if yLast < y {
					yMin, yMax = yA, yB
				} else {
					yMin, yMax = yB, yA
				}
				for yPos := yMin; yPos <= yMax + stroke; yPos++ {
					c := getRGBA(vis, xPos, v.h-yPos)
					c.R += r
					c.G += g
					c.B += b
				}
			}

			yLast = y
			xLast = x
		}
	}

	// Flatline data from last data point out to end of graph, and make line
	// dotted after real data has ceased to be available.
	n := v.nS[xLast] + v.nF[xLast] + v.nA[xLast]
	r := uint8(32 + 128 * v.nF[xLast] / n)
	g := uint8(32 + 128 * v.nA[xLast] / n)
	b := uint8(32 + 128 * v.nS[xLast] / n)
	for x := xLast; x < v.w; x += 4 {
		for yPos := yLast; yPos <= yLast + stroke; yPos++ {
			c := getRGBA(vis, x, v.h-yPos)
			c.R += r
			c.G += g
			c.B += b
		}
	}

	return vis
}

func (v *runTimeLine) drawGrid(vis *image.RGBA) {

	// Draw vertical grid lines, if vertical divisions were specified.
	if v.xGrid > 0 {
		for i := 1; i < v.xGrid; i++ {
			drawXGridLine(vis, i*v.w/v.xGrid)
		}
	}

	// Draw horizontal grid lines on each doubling of the run time in seconds.
	for y := float64(v.h); y > 0; y -= v.yLog2 {
		drawYGridLine(vis, int(y))
	}
}
