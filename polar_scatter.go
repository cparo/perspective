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

type polar_scatter struct {
	w     int         // Width of the visualization
	h     int         // Height of the visualization
	vis   *image.RGBA // Visualization canvas
	tA    float64     // Lower limit of time range to be visualized
	tτ    float64     // Length of time range to be visualized
	p0    float64     // Temporal period phase offset value
	pτ    float64     // The periodic interval length
	yLog2 float64     // Number of pixels over which elapsed times double
	cΔ    float64     // Increment for color channel value increases
	ϕΔ    float64     // Angular value, in radians, of a step in time
}

// NewPolarScatter returns a polar scatter-visualization generator.
func NewPolarScatter(
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
	return (&polar_scatter{
		width,
		height,
		initializeVisualization(width, height, bg),
		float64(minTime),
		float64(maxTime - minTime),
		float64(phasePoint%period - period),
		float64(period),
		float64(yLog2),
		saturated / colorSteps,
		2 * math.Pi / float64(period)}).drawGrid()
}

// Record accepts an EventData pointer and plots it onto the visualization.
func (v *polar_scatter) Record(e *EventData) {

	// Angular position (for event start time).
	ϕ := math.Pi / 2 - v.ϕΔ * math.Mod(float64(e.Start) - v.p0, v.pτ)

	// Distance from center of visualization (for event run time).
	r := v.yLog2 * math.Log2(float64(e.Run))

	// Translate to Cartesian coordinates (with the quirk of the upside-down
	// y-axis common in computer images). A bit of random "noise" is added to
	// avoid distracting Moire patterns as an artefact of the translation.
	x := int(r * math.Cos(ϕ) + 4*rng.Float64() - 2) + v.w / 2
	y := v.h / 2 - int(r * math.Sin(ϕ) + 4*rng.Float64() - 2)

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
		// In-progress events are shown as green points, which are kept from
		// desaturating similar to failures to help distinguish high-density
		// clusters of in-progress events from clusters of successfully
		// completed events.
		c.G = uint8(math.Min(saturated, float64(c.G)+v.cΔ))
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *polar_scatter) Render() image.Image {
	return v.vis
}

// Draw crosshair grid on the visualization to clearly show center point and
// quartile angular positions relative to the period start.
func (v *polar_scatter) drawGrid() *polar_scatter {

	// Draw crosshairs.
	drawXGridLine(v.vis, v.w/2)
	drawYGridLine(v.vis, v.h/2)

	// TODO: Draw circles on ylog2 intervals

	// Return the polar_scatter visualization struct, so this can be
	// conveniently used in the visualization's constructor.
	return v
}
