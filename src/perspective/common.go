package perspective

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	bg     = 33    // Gray level for visualization backgrounds
	grid   = 45    // Gray level for grid lines
	opaque = 255   // Alpha component of an opaque color value
	maxC16 = 65535 // Maximum color value returned from image.RGBA.At()
)

// Stripped-down struct for the data to be submitted to the actual visualization
// generators, after filtering of the binary-formatted input data.
type EventDataPoint struct {
	Start  int32 // In seconds since the beginning of the Unix epoch.
	Run    int32 // Event run time, in seconds.
	Status int16 // Zero indicates success, non-zero indicates failure.
}

// Abstract interface for visualization generators.
type Visualizer interface {
	Record(EventDataPoint)
	Render() *image.RGBA
}

// Utility function for converting 32-bit color components (with an unsigned
// 16-bit range, as is returned by image.RGBA.At()) into a standard 8-bit RGBA
// color value usable for image.RGBA.Set().
func c8(r16 uint32, g16 uint32, b16 uint32) color.RGBA {
	return color.RGBA{
		uint8(r16 >> 8),
		uint8(g16 >> 8),
		uint8(b16 >> 8),
		opaque}
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

// Utility function for setting up a visualization canvas.
func initializeVisualization(width int, height int) *image.RGBA {
	vis := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{bg, bg, bg, opaque}
	draw.Draw(vis, vis.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
	return vis
}

// Utility function for returning the larger of two integers.
func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

// Utility function for pixel-by-pixel drawing.
func plot(vis *image.RGBA, x int, y int, r16 uint32, g16 uint32, b16 uint32) {
	vis.Set(x, y, c8(r16, g16, b16))
}
