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

type countLines struct {
	w         int       // Width of the visualization
	h         int       // Height of the visualization
	tA        float64   // Lower limit of time range to be visualized
	tτ        float64   // Length of time range to be visualized
	s         []float64 // Counts of successful events by x-axis position
	f         []float64 // Counts of failed events by x-axis position
	resonance float64   // Inverse of geometric decay for moving-window
	window    int       // Moving-window width
	xGrid     int       // Number of vertical grid divisions
	bg        int       // Background grey level
}

// NewCountLines returns an line-graph event-count-visualization generator.
func NewCountLines(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int,
	resonance float64,
	xGrid int) Visualizer {

	// Select a window which is appropriate for the selected resonance
	window := 0;
	n := 1.0;
	for  window < width && n > 0.02 {
		n = n * resonance
		window++
	}

	return &countLines{
		width,
		height,
		float64(minTime),
		float64(maxTime - minTime),
		make([]float64, width),
		make([]float64, width),
		resonance,
		window, //width / 42,
		xGrid,
		bg}
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *countLines) Record(e *EventData) {

	resonance := v.resonance
	window := v.window

	// Position on the x-axis corresponds to the event's start time. Event run
	// times are not taken into account in this visualization. A margin is left
	// on each edge for smoothing purposes.
	x := int(float64(v.w) * (float64(e.Start) - v.tA) / v.tτ)

	// Ignore active events
	if e.Status < 0 {
		return
	}

	var frame []float64
	if e.Status == 0 {
		frame = v.s
	} else {
		frame = v.f
	}

	// Line is smothed with a bi-directional variation on an exponential moving
	// average (which is implemented as a windowed FIR here for efficiency
	// purposes).
	frame[x] = frame[x] + 1
	leftWindow := int(math.Min(float64(window), float64(x)))
	for i, n := 1, 1.0; i < leftWindow; i++ {
		n = n * resonance
		frame[x-i] = frame[x-i] + n
	}
	rightWindow := int(math.Min(float64(window), float64(v.w-x-1)))
	for i, n := 1, 1.0; i < rightWindow; i++ {
		n = n * resonance
		frame[x+i] = frame[x+i] + n
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *countLines) Render() image.Image {

	// Stroke width (for visibility and calligraphic effect)
	stroke := v.h / 32

	// Initialize our image canvas and grid.
	vis := initializeVisualization(v.w, v.h, v.bg)
	v.drawGrid(vis)

	// Find the highest point of the chart to normalize the height of the lines.
	maxCount := float64(0)
	for x := 0; x < v.w; x++ {
		maxCount = math.Max(maxCount, v.s[x])
		maxCount = math.Max(maxCount, v.f[x])
	}
	scale := float64(v.h) / (maxCount)

	// Draw the masts, with successes stacked atop failures.
	var yMin, yMax int
	for x := 1; x < v.w-1; x++ {

		// Odd extra "yMin" logic here is to make sure steep sections of the
		// line plot are drawn as a connected line rather than as a series of
		// disjoint dashes.
		sP := int(math.Ceil(v.s[x-1] * scale))
		sC := int(math.Ceil(v.s[x+0] * scale))
		sN := int(math.Ceil(v.s[x+1] * scale))
		yMin = intMinOfThree(sC - stroke, sP, sN)
		yMax = sC
		for y := yMin; y < yMax; y++ {
			cS := getRGBA(vis, x, v.h-y)
			cS.R += 24
			cS.G += 24
			cS.B += 128
		}

		fP := int(math.Ceil(v.f[x-1] * scale))
		fC := int(math.Ceil(v.f[x+0] * scale))
		fN := int(math.Ceil(v.f[x+1] * scale))
		yMin = intMinOfThree(fC - stroke, fP, fN)
		yMax = fC
		for y := yMin; y < yMax; y++ {
			cF := getRGBA(vis, x, v.h-y)
			cF.R += 128
			cF.G += 24
			cF.B += 24
		}
	}

	return vis
}

func (v *countLines) drawGrid(vis *image.RGBA) {

	// Render hatching to indicate dropoff at the end of the plot due to the
	// smoothing window.
	for y := 0; y < v.h; y++ {
		for x := 0; x < v.window; x++ {
			if (x+y)%8 < 3 || (x-y)%8 == 0 {
				cS := getRGBA(vis, x, y)
				cS.R += 18
				cS.G += 18
				cS.B += 18
			}
		}
		for x := v.w - v.window; x < v.w; x++ {
			if (x+y)%8 < 3 || (x-y)%8 == 0 {
				cS := getRGBA(vis, x, y)
				cS.R += 18
				cS.G += 18
				cS.B += 18
			}
		}
	}

	// Draw vertical grid lines, if vertical divisions were specified.
	if v.xGrid > 0 {
		for i := 1; i < v.xGrid; i++ {
			drawXGridLine(vis, i*v.w/v.xGrid)
		}
	}
}
