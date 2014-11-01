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

type rollingStack struct {
	w  int             // Width of the visualization
	h  int             // Height of the visualization
	tA float64         // Lower limit of time range to be visualized
	tτ float64         // Length of time range to be visualized
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
		float64(maxTime - minTime),
		make(map[int16][]int),
		make([]float64, width)}
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *rollingStack) Record(e *EventData) {
	for int(e.Status)+1 > len(v.n) {
		v.n[int16(len(v.n))] = make([]int, v.w)
	}
	w := float64(v.w)
	s := float64(e.Start)
	x := int(math.Min(w-1, w*(s-v.tA)/v.tτ))
	v.n[e.Status][x]++
	v.σ[x]++
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *rollingStack) Render() image.Image {

	// Initialize our image canvas.
	vis := initializeVisualization(v.w, v.h)

	// Draw the rolling stack, giving each failure type a different color and
	// scaling the height of the visualization to normalize for the overall
	// event density at each x-position such that each column advancing over the
	// x-axis is a representation of the stacked rate of each error category
	// relative to the total number of events started within the time window
	// represented by that position on the x-axis.
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
