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

type errorStack struct {
	w  int           // Width of the visualization
	h  int           // Height of the visualization
	bg int           // Background grey level
	n  map[int16]int // Event counts by exit status code
	σ  float64       // Total count of failed events
}

// NewErrorStack returns an error-stack-visualization generator.
func NewErrorStack(width int, height int, bg int) Visualizer {
	return &errorStack{width, height, bg, make(map[int16]int), 0}
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *errorStack) Record(e *EventData) {
	// For this visualization, we only care about failed events.
	if e.Status > 0 {
		for int(e.Status)+1 > len(v.n) {
			v.n[int16(len(v.n))] = 0
		}
		v.n[e.Status]++
		v.σ++
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *errorStack) Render() image.Image {

	// Initialize our image canvas.
	vis := initializeVisualization(v.w, v.h, v.bg)

	// Draw the stack, giving each failure type a different color and scaling
	// the overall stack to fill the image canvas such that each failure case
	// occupies space proportionate to its relative frequency amongst the
	// failure cases recorded.
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
