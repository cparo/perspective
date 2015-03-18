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
	"image/draw"
	"unsafe"
)

const (
	grid      = 45    // Gray level for grid lines
	opaque    = 255   // Alpha component of an opaque color value
	saturated = 255   // Saturated 8-bit color value
	maxC16    = 65535 // Maximum color value returned from image.RGBA.At()
)

// Bloom-effect convolution kernel (Gaussian distribution with a standard
// deviation of 0.5, denormalized so center pixel will match expectation from
// selected color intensity value). At an intensity level of one, individual
// non-stacked data points should also be more visible with this renderer vs.
// the standard scatter graph renderer because of the blurred-circle effect
// extending beyond the data point's central pixel.
var pointConvolutionKernel = [25]float64{
	0.000004, 0.000455, 0.001978, 0.000455, 0.000004,
	0.000455, 0.053093, 0.230420, 0.053093, 0.000455,
	0.001978, 0.230420, 1.000000, 0.230420, 0.001978,
	0.000455, 0.053093, 0.230420, 0.053093, 0.000455,
	0.000004, 0.000455, 0.001978, 0.000455, 0.000004,
}

// Struct to represent data to submit to the visualization generators, and to be
// used for the binary log format.
type EventData struct {
	ID       int32 // Event identifier.
	Start    int32 // In seconds since the beginning of the Unix epoch.
	Run      int32 // Event run time, in seconds.
	Type     uint8 // Event type indication.
	Status   int8  // 0 for success, >0 for failure, <0 for in-progress.
	Region   uint8 // Region identifier, 0 if undefined.
	Progress uint8 // Event progress percentage.
}

// Abstract interface for visualization generators.
type Visualizer interface {
	Record(*EventData)
	Render() image.Image
}

// Utility function to draw a vertical grid line at the specified x position.
func drawXGridLine(vis *image.RGBA, x int) {
	c := color.RGBA{grid, grid, grid, opaque}
	h := vis.Bounds().Max.Y
	for y := 0; y < h; y++ {
		vis.Set(x, y, c)
	}
}

// Utility function to draw a horizontal grid line as the specified y position.
func drawYGridLine(vis *image.RGBA, y int) {
	c := color.RGBA{grid, grid, grid, opaque}
	w := vis.Bounds().Max.X
	for x := 0; x < w; x++ {
		vis.Set(x, y, c)
	}
}

// Utility function get getting a shade of red to represent a class of failures
// in a stack representing multiple failure types.
func getErrorStackColor(layer int, layers int) color.RGBA {
	v := float64(layer) * 255 / float64(layers)
	return color.RGBA{
		uint8(127 + v/2),
		uint8(11 + v*2/3),
		uint8(11 + v*2/3),
		opaque}
}

// Utility function to return a pointer to a pixel in an RGBA image, which can
// be used to shave a little time (about 10% as measured over repeated "before"
// vs. "after" tests - which isn't huge, but does help substantially with
// improving the UX potential for using these visualizations in an interactive
// context) off of rendering our visualizations relative to the more obvious
// method of using the image's At() and Set() methods. Obviously, when using
// this function, the resulting pointer should be read/updated and then allowed
// to fall out of scope instead of being reused or carelessly passed around.
func getRGBA(i *image.RGBA, x int, y int) *color.RGBA {
	if (image.Point{x, y}).In(i.Rect) {
		o := i.PixOffset(x, y)
		return (*color.RGBA)(unsafe.Pointer(&i.Pix[o]))
	}
	// If the specified coordinates are out-of-bounds, return a pointer to a
	// new "empty" pixel - effectively allowing the image to quietly ignore any
	// attempts to write to an out-of-bounds location. If attempting to write
	// data outside of the image bounds is something which should be considered
	// an error in context, that check can be done before calling this function.
	// Here we'll allow it since there are situations where a rendering tool
	// would be substantially simplified by allowing it to safely draw "outside
	// the lines".
	return &color.RGBA{}
}

// Utility function for setting up a visualization canvas.
func initializeVisualization(width int, height int, bg int) *image.RGBA {
	vis := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{uint8(bg), uint8(bg), uint8(bg), opaque}
	draw.Draw(vis, vis.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
	return vis
}
