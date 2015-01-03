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
	"image/color"
	"math"
)

type wave struct {
	w   int          // Width of the visualization
	h   int          // Height of the visualization
	bg  int          // Background grey level.
	tA  float64      // Lower limit of time range to be visualized
	tΩ  float64      // Upper limit of time range to be visualized
	vis *image.RGBA  // Visualization canvas
	x   int          // Current x-position for recording events
	p   []*EventData // Passing event data points in current x-position
	f   []*EventData // Failing event data points in current x-position
}

// NewWave returns a wave-visualization generator.
func NewWave(
	width int,
	height int,
	bg int,
	minTime int,
	maxTime int) Visualizer {
	return &wave{
		width,
		height,
		bg,
		float64(minTime),
		float64(maxTime),
		initializeVisualization(width, height, bg),
		0,
		[]*EventData{},
		[]*EventData{}}
}

// Record accepts an EventData pointer and plots it onto the visualization.
//
// NOTE: Event input is expected to be received in chronological order. If
//       it is not received in chronological order, the graph will not be
//       rendered properly (with the severity of the issue being dependent
//       upon the degree of deviation between the input order and the ideal
//       chronologically-sorted input.
func (v *wave) Record(e *EventData) {
	pʹ := make([]*EventData, 0, len(v.p)+64)
	fʹ := make([]*EventData, 0, len(v.f)+64)
	for _, p := range v.p {
		if p.Start+p.Run > e.Start {
			pʹ = append(pʹ, p)
		}
	}
	for _, f := range v.f {
		if f.Start+f.Run > e.Start {
			fʹ = append(fʹ, f)
		}
	}
	v.p = pʹ
	v.f = fʹ
	inProgress := 0
	if e.Status == 0 {
		v.p = append(v.p, e)
	} else if e.Status > 0 {
		v.f = append(v.f, e)
	} else {
		inProgress++
	}
	t := float64(e.Start)
	xʹ := int(float64(v.w) * (t - v.tA) / (v.tΩ - v.tA))
	for xʹ > v.x {
		v.x++
		yP := 0
		yF := 0
		for i := 0; i < len(v.p); i++ {
			p := v.p[len(v.p)-i-1]
			Δ := saturated * float64(e.Start-p.Start) / float64(p.Run+1)
			c := color.RGBA{
				uint8(math.Min(saturated, float64(v.bg)+Δ/4)),
				uint8(math.Min(saturated, float64(v.bg)+Δ/4)),
				uint8(math.Min(saturated, float64(v.bg)+Δ)),
				opaque}
			yPʹ := yP + 1
			for ; yP < yPʹ; yP++ {
				v.vis.Set(v.x, v.h/2-yP-inProgress/2, c)
			}
		}
		for i := 0; i < len(v.f); i++ {
			f := v.f[len(v.f)-i-1]
			Δ := saturated * float64(e.Start-f.Start) / float64(f.Run+1)
			c := color.RGBA{
				uint8(math.Min(saturated, float64(v.bg)+Δ)),
				uint8(math.Min(saturated, float64(v.bg)+Δ/4)),
				uint8(math.Min(saturated, float64(v.bg)+Δ/4)),
				opaque}
			yFʹ := yF + 1
			for ; yF < yFʹ; yF++ {
				v.vis.Set(v.x, v.h/2+yF+inProgress/2, c)
			}
		}
	}
}

// Render returns the visualization constructed from all previously-recorded
// data points.
func (v *wave) Render() image.Image {
	return v.vis
}
