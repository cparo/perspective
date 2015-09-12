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
	"math/rand"
)

type medianLines struct {
	w         int       // Width of the visualization
	h         int       // Height of the visualization
	s         []float64 // Channel for successful events
	f         []float64 // Channel for failed events
	a         []float64 // Channel for active events
	n         []float64 // Array for count of events on each x-coordinate slice
	resonance float64   // Inverse of geometric decay for moving window
	tA        float64   // Lower limit of time range to be visualized
	tτ        float64   // Length of time range to be visualized
	yLog2     float64   // Number of pixels over which elapsed times double
	xGrid     int       // Number of vertical grid divisions
	bg        int       // Background gray level
}

// NewMedianLines returns a weighted-median-line visualization generator.
func NewMedianLines(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int,
	yLog2 float64,
	resonance float64,
	xGrid int) Visualizer {

	return (&medianLines{
		width,
		height,
		make([]float64, (width)*(height)),
		make([]float64, (width)*(height)),
		make([]float64, (width)*(height)),
		make([]float64, (width)),
		resonance,
		float64(minTime),
		float64(maxTime - minTime),
		float64(yLog2),
		xGrid,
		bg})
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *medianLines) Record(e *EventData) {

	x := int(float64(v.w) * (float64(e.Start) - v.tA) / v.tτ)
	y := v.h - int(v.yLog2*math.Log2(float64(e.Run)))

	w, h := v.w, v.h

	// Apply resonance-scaled noise as a pre-smoothing measure.
	x += int(rand.NormFloat64() * v.resonance * float64(w) / 128)

	// Only look at successfully-completed events
	var frame []float64
	if e.Status == 0 {
		frame = v.s
		if x >= 0 && x < w && y >= 0 && y < h {
			frame[y*w+x]++
			v.n[x]++
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *medianLines) Render() image.Image {

	// Create a normal image canvas to render to.
	w, h := v.w, v.h
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

	// Find maximum event density
	nMax := 0.0
	for x := 0; x < w; x++ {
		if v.n[x] > nMax {
			nMax = v.n[x]
		}
	}

	// Find (unsmoothed) median/percentile lines.
	p05 := make([]float64, w)
	p25 := make([]float64, w)
	p50 := make([]float64, w)
	p75 := make([]float64, w)
	p95 := make([]float64, w)
	s := v.s
	for x := 0; x < w; x++ {
		// Get position of 5% point
		n05 := v.n[x] / 20
		y, i := 0, float64(0)
		for y = 0; y < h; y++ {
			i += s[y*w+x]
			if i >= n05 {
				break
			}
		}
		p05[x] = float64(y)
		// Get position of 25% point
		n25 := v.n[x] / 4
		y, i = 0, float64(0)
		for y = 0; y < h; y++ {
			i += s[y*w+x]
			if i >= n25 {
				break
			}
		}
		p25[x] = float64(y)
		// Get position of 50% point
		n50 := v.n[x] / 2
		y, i = 0, float64(0)
		for y = 0; y < h; y++ {
			i += s[y*w+x]
			if i >= n50 {
				break
			}
		}
		p50[x] = float64(y)
		// Get position of 75% point
		n75 := 3 * v.n[x] / 4
		y, i = 0, float64(0)
		for y = 0; y < h; y++ {
			i += s[y*w+x]
			if i >= n75 {
				break
			}
		}
		p75[x] = float64(y)
		// Get position of 95% point
		n95 := 19 * v.n[x] / 20
		y, i = 0, float64(0)
		for y = 0; y < h; y++ {
			i += s[y*w+x]
			if i >= n95 {
				break
			}
		}
		p95[x] = float64(y)
	}

	// Find window for smoothing filter.
	window := 0
	for n := 1.0; window < w && n > 0.02; window++ {
		n = n * v.resonance
	}

	// Render (smoothed) median/percentile lines.
	for x := 0; x < w; x++ {
		// Ignore x-coordinates with no data.
		if v.n[x] > 0 {
			leftWindow := int(math.Min(float64(window), float64(x)))
			rightWindow := int(math.Min(float64(window), float64(v.w-x-1)))
			smoothedP05 := p05[x]
			smoothedP25 := p25[x]
			smoothedP50 := p50[x]
			smoothedP75 := p75[x]
			smoothedP95 := p95[x]
			divisor := 1.0
			for i, n := 1, 1.0; i < leftWindow; i++ {
				if v.n[x-i] > 0 {
					n = n * v.resonance
					smoothedP05 += n * p05[x-i]
					smoothedP25 += n * p25[x-i]
					smoothedP50 += n * p50[x-i]
					smoothedP75 += n * p75[x-i]
					smoothedP95 += n * p95[x-i]
					divisor += n
				}
			}
			for i, n := 1, 1.0; i < rightWindow; i++ {
				if v.n[x+i] > 0 {
					n = n * v.resonance
					smoothedP05 += n * p05[x+i]
					smoothedP25 += n * p25[x+i]
					smoothedP50 += n * p50[x+i]
					smoothedP75 += n * p75[x+i]
					smoothedP95 += n * p95[x+i]
					divisor += n
				}
			}
			multiplier := v.n[x]
			for i, n := 1, 1.0; i < leftWindow; i++ {
				n = n * v.resonance
				multiplier += n * v.n[x-i]
			}
			for i, n := 1, 1.0; i < rightWindow; i++ {
				n = n * v.resonance
				multiplier += n * v.n[x+i]
			}
			multiplier = multiplier / nMax / divisor
			yMin := int(smoothedP05 / divisor)
			yMax := int(smoothedP95 / divisor)
			for y := yMin; y <= yMax; y++ {
				c := getRGBA(vis, x, y)
				c.R += uint8(32 * multiplier)
				c.G += uint8(32 * multiplier)
				c.B += uint8(64 * multiplier)
			}
			yMin = int(smoothedP25 / divisor)
			yMax = int(smoothedP75 / divisor)
			for y := yMin; y <= yMax; y++ {
				c := getRGBA(vis, x, y)
				c.R += uint8(64 * multiplier)
				c.G += uint8(64 * multiplier)
				c.B += uint8(128 * multiplier)
			}
			yMin = int(smoothedP50/divisor - 1)
			yMax = int(smoothedP50/divisor + 1)
			for y := yMin; y <= yMax; y++ {
				c := getRGBA(vis, x, y)
				c.R = uint8(math.Min(float64(c.R)+96*multiplier, saturated))
				c.G = uint8(math.Min(float64(c.G)+96*multiplier, saturated))
				c.B = uint8(math.Min(float64(c.B)+192*multiplier, saturated))
			}
		}
	}

	return vis
}
