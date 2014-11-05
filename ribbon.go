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

type ribbon struct {
	w    int     // Width of the visualization
	h    int     // Height of the visualization
	tA   float64 // Lower limit of time range to be visualized
	tτ   float64 // Length of time range to be visualized
	pass []int   // Successful events by x-axis position
	fail []int   // Failed events by x-axis position
	open []int   // In-progress events by x-axis position
	pMax int     // Maximum number of successful events in any x position
	fMax int     // Maximum number of failed events in any x position
	oMax int     // Maximum number of in-progress events in any x position
}

// NewRibbon returns a ribbon-visualization generator.
func NewRibbon(width int, height int, minTime int, maxTime int) Visualizer {

	// Max counts are initialized to 1 instead of 0 to avoid division-by-zero
	// issues with feeds which do not contain examples of all three of
	// completed, failed, and in-progress events.
	return &ribbon{
		width,
		height,
		float64(minTime),
		float64(maxTime - minTime),
		make([]int, width),
		make([]int, width),
		make([]int, width),
		1,
		1,
		1}
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *ribbon) Record(e *EventData) {

	w := float64(v.w)
	s := float64(e.Start)
	x := int(math.Min(w-1, w*(s-v.tA)/v.tτ))
	if e.Status == 0 {
		if v.pass[x]++; v.pass[x] > v.pMax {
			v.pMax++
		}
	} else if e.Status > 0 {
		if v.fail[x]++; v.fail[x] > v.fMax {
			v.fMax++
		}
	} else {
		if v.open[x]++; v.open[x] > v.oMax {
			v.oMax++
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *ribbon) Render() image.Image {

	// Initialize our canvas.
	vis := initializeVisualization(v.w, v.h)

	// Draw the ribbon. The ribbon's display is a simple representation of the
	// number of passed and failed events which occurred within each pixel's
	// represented time range on the x-axis - with successes showing as full
	// strength at the top and fading toward the bottom while failures show as
	// strength at the bottom and fade toward the top (this being done to make
	// it easy to see both values within a constrained space, while maintaining
	// a unified and attractive presentation).
	h := float64(v.h)
	fMax := float64(v.fMax)
	pMax := float64(v.pMax)
	oMax := float64(v.oMax)
	for x := 0; x < v.w; x++ {
		r := saturated * float64(v.fail[x]) / fMax
		b := saturated * float64(v.pass[x]) / pMax
		w := bg + (saturated-bg)*float64(v.open[x])/oMax
		for y := 0; y < v.h; y++ {
			c := getRGBA(vis, x, y)
			c.R = uint8(math.Min(saturated, w+r*float64(y)/h))
			c.G = uint8(w)
			c.B = uint8(math.Min(saturated, w+b*(1-(float64(y)/h))))
		}
	}

	return vis
}
