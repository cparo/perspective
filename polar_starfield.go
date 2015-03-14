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

// Note that floating-point pre-rendering canvases have a two-pixel bleed on all
// edges to allow for simple use of the bloom effect's convolution kernel.
type polarStarfield struct {
	w     int       // Width of the visualization
	h     int       // Height of the visualization
	s     []float64 // Channel for successful events
	f     []float64 // Channel for failed events
	a     []float64 // Channel for active events
	tA    float64   // Lower limit of time range to be visualized
	tτ    float64   // Length of time range to be visualized
	p0    float64   // Temporal period phase offset value
	pτ    float64   // The periodic interval length
	yLog2 float64   // Number of pixels over which elapsed times double
	cΔ    float64   // Increment for color channel value increases
	bg    int       // Background gray level
	ϕΔ    float64   // Angular value, in radians, of a step in time
}

// NewPolarStarfield returns a polar floating-point scatter-visualization
// generator.
func NewPolarStarfield(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int,
	phasePoint int,
	period int,
	yLog2 float64,
	colorSteps float64) Visualizer {

	// Ensure we have a positive, non-zero period length. If we don't (for
	// instance, if none was specified by the end user and we were given a
	// negative value to signify this), then take the length of time covered by
	// this visualization as the overall period length.
	if period <= 0 {
		period = maxTime - minTime
	}

	// Note the calculation for the temporal phase offset value, which is used
	// to normalize the phase-offset time to the the corresponding same-angle
	// point in time just before the logical start of a period (it will always
	// be a value >= 0) for the sake of simplifying positional calculations
	// when plotting individual event data points.

	return (&polarStarfield{
		width,
		height,
		make([]float64, (width+4)*(height+4)),
		make([]float64, (width+4)*(height+4)),
		make([]float64, (width+4)*(height+4)),
		float64(minTime),
		float64(maxTime - minTime),
		float64(phasePoint%period - period),
		float64(period),
		float64(yLog2),
		saturated / colorSteps,
		bg,
		2 * math.Pi / float64(period)})
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *polarStarfield) Record(e *EventData) {

	// Angular position (for event start time).
	ϕ := math.Pi / 2 - v.ϕΔ * math.Mod(float64(e.Start) - v.p0, v.pτ)

	// Apply a bit of random "noise" (on a Gaussian distribution, with a
	// standard deviation of 0.5), to the time scale to avoid Moire patterns
	// and quantization artifacts which could distract from real patterns or
	// create a false sense of consistency in the run times of short-lived
	// events.
	t := float64(e.Run) + rng.NormFloat64() / 2

	// Distance from center of visualization (for event run time).
	r := v.yLog2 * math.Log2(t)

	w, h := v.w, v.h

	// Translate to Cartesian coordinates (with the quirk of the upside-down
	// y-axis common in computer images).
	xP := int(r*math.Cos(ϕ)) + w/2
	yP := h/2 - int(r*math.Sin(ϕ))

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
				frame[y*w+x] += pointConvolutionKernel[iK]
				iK++
			}
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *polarStarfield) Render() image.Image {

	// Create a normal image canvas to render to.
	w, h, cΔ := v.w, v.h, v.cΔ
	vis := initializeVisualization(w, h, v.bg)

	// Draw crosshairs.
	drawXGridLine(vis, v.w/2)
	drawYGridLine(vis, v.h/2)

	// TODO: Draw circles on ylog2 intervals

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
