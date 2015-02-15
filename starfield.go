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

// Bloom-effect convolution kernel (Gaussian distribution with a standard
// deviation of 0.5, denormalized so center pixel will match expectation from
// selected color intensity value). At an intensity level of one, individual
// non-stacked data points should also be more visible with this renderer vs.
// the standard scatter graph renderer because of the blurred-circle effect
// extending beyond the data point's central pixel.
var kernel = [25]float64{
	0.000004, 0.000455, 0.001978, 0.000455, 0.000004,
	0.000455, 0.053093, 0.230420, 0.053093, 0.000455,
	0.001978, 0.230420, 1.000000, 0.230420, 0.001978,
	0.000455, 0.053093, 0.230420, 0.053093, 0.000455,
	0.000004, 0.000455, 0.001978, 0.000455, 0.000004,
}

// Note that floating-point pre-rendering canvases have a two-pixel bleed on all
// edges to allow for simple use of teh bloom effect's convolution kernel.
type starfield struct {
	w     int       // Width of the visualization
	h     int       // Height of the visualization
	s     []float64 // Channel for successful events
	f     []float64 // Channel for failed events
	a     []float64 // Channel for active events
	tA    float64   // Lower limit of time range to be visualized
	tτ    float64   // Length of time range to be visualized
	yLog2 float64   // Number of pixels over which elapsed times double
	cΔ    float64   // Increment for color channel value increases
	xGrid int       // Number of vertical grid divisions
	bg    int       // Background gray level
}

// NewStarfield returns a floating-point scatter-visualization generator.
func NewStarfield(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int,
	yLog2 float64,
	colorSteps float64,
	xGrid int) Visualizer {

	return (&starfield{
		width,
		height,
		make([]float64, (width+4)*(height+4)),
		make([]float64, (width+4)*(height+4)),
		make([]float64, (width+4)*(height+4)),
		float64(minTime),
		float64(maxTime - minTime),
		float64(yLog2),
		saturated / colorSteps,
		xGrid,
		bg})
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *starfield) Record(e *EventData) {

	xP := int(float64(v.w) * (float64(e.Start) - v.tA) / v.tτ)
	yP := v.h - int(v.yLog2*math.Log2(float64(e.Run)))

	w, h := v.w, v.h

	// Select appropriate canvas layer based on the event's status code.
	var frame []float64
	if e.Status == 0 {
		frame = v.s
	} else if e.Status > 0 {
		frame = v.f
	} else {
		frame = v.a
	}

	// Coordinates are translated to account for bleed on floating-point canvas
	// and convolution kernel.
	xMin, xMax := xP, xP + 5
	yMin, yMax := yP, yP + 5
	if xMin >= 0 && xMax < w+2 && yMin >= 0 && yMax < h+2 {
		// Convolved plot point fits entirely within canvas (including its bleed
		// zone), so we can safely draw it without falling off the edge.
		iK := 0
		for y := yMin; y < yMax; y++ {
			for x := xMin; x < xMax; x++ {
				frame[y*w+x] += kernel[iK]
				iK++
			}
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *starfield) Render() image.Image {

	// Create a normal image canvas to render to.
	w, h, cΔ := v.w, v.h, v.cΔ
	vis := initializeVisualization(w, h, v.bg)

	// Draw vertical grid lines, if vertical divisions were specified.
	if v.xGrid > 0 {
		for i := 1; i < v.xGrid; i++ {
			drawXGridLine(vis, i*w/v.xGrid)
		}
	}

	// Draw horizontal grid lines on each doubling of the run time in seconds.
	for y := float64(h); y > 0; y -= v.yLog2 {
		drawYGridLine(vis, int(y))
	}

	// Render point data to final image.
	s, f, a := v.s, v.f, v.a
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y+2)*w + x + 2
			if s[i] > 0 || f[i] > 0 || a[i] > 0 {
				c := getRGBA(vis, x, y)
				c.R = uint8(math.Min(saturated, float64(c.R)+(s[i]/4+f[i])*cΔ))
				c.G = uint8(math.Min(saturated, float64(c.G)+(s[i]/4+a[i])*cΔ))
				c.B = uint8(math.Min(saturated, float64(c.B)+(s[i])*cΔ))
			}
		}
	}

	return vis
}
