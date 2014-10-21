package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"perspective"
	"regexp"
	"time"
)

// This struct is written to pack neatly into a 64-byte line while still
// accommodating any data we will realistically be pulling out of out event
// database in the next couple of decades. This may not matter much for
// performance, but it is pretty convenient for reading a hex dump of the
// resulting binary event log format.
type eventData struct {
	EventID   int32 // Event ID as recorded in Alpha.
	StartTime int32 // In seconds since the beginning of the Unix epoch.
	RunTime   int32 // Event run time, in seconds.
	EventType int16 // Event type ID as recorded in Alpha.
	Status    int16 // Zero indicates success, non-zero indicates failure.
}

// Variables for command-line option flags.
var (
	errorReasonFilterConf string
	typeFilter            int
	minTime               int
	maxTime               int
	xGrid                 int
	yLog2                 float64
	width                 int
	height                int
	colorSteps            int
)

func eventFilter(startTime int, eventType int) bool {
	if minTime < startTime && maxTime > startTime {
		if typeFilter < 0 || eventType == typeFilter {
			return true
		}
	}
	return false
}

func atEOF(err error, message string) bool {
	if err != nil {
		if err == io.EOF {
			return true
		}
		log.Println(message)
		log.Println(err)
		os.Exit(1)
	}
	return false
}

func exitOnError(err error, message string) {
	if err != nil {
		log.Println(message)
		log.Println(err)
		os.Exit(1)
	}
}

func findHitCount(value uint8, cV []uint8) int {
	for i := 0; i < len(cV); i++ {
		if value == cV[i] {
			return i
		}
	}
	return 0
}

func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func openFiles(iPath string, oPath string) (iFile *os.File, oFile *os.File) {

	var err error

	iFile, err = os.Open(iPath)
	exitOnError(err, "Failed to open input file for reading.")

	oFile, err = os.Create(oPath)
	exitOnError(err, "Failed to open output file for writing.")

	return iFile, oFile
}

func getErrorCode(errorReason string, errorFilters []*regexp.Regexp) int16 {
	var i int
	for i = 0; i < len(errorFilters); i++ {
		if errorFilters[i].MatchString(errorReason) {
			return int16(i + 1)
		}
	}
	// Implied "other" case, which will return a value one past the last value
	// which should be associated with a filter, indicating that no filters
	// matched the errorReason we were given. Note that the error codes start at
	// 1, not 0, so in the example case of our having four error reason filters
	// (including one for a blank error reason), this will be code 5, not 4.
	return int16(i + 1)
}

func getErrorStackColor(layer int, layers int) color.RGBA {
	v := float64(layer) * 255 / float64(layers)
	r8 := uint8(127 + v/2)
	g8 := uint8(11 + v*2/3)
	b8 := uint8(11 + v*2/3)
	return color.RGBA{r8, g8, b8, 255}
}

func initializeVisualization() *image.RGBA {

	vis := image.NewRGBA(image.Rect(0, 0, width, height))
	background := color.RGBA{33, 33, 33, 255}
	draw.Draw(vis, vis.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
	return vis
}

func generateErrorStackVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewErrorStack(width, height)

	for {

		var event eventData

		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func generateHistogramVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewHistogram(width, height, yLog2)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func generateRollingStackVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewRollingStack(width, height, minTime, maxTime)

	for {

		var event eventData

		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func generateScatterVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewScatter(
		width,
		height,
		minTime,
		maxTime,
		yLog2,
		colorSteps,
		xGrid)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func generateStatusStackVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewStatusStack(width, height)

	for {

		var event eventData

		err := binary.Read(binReader, binary.LittleEndian, &event)
		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func generateSweepVisualization(iPath string, oPath string) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	v := perspective.NewSweep(
		width,
		height,
		minTime,
		maxTime,
		yLog2,
		colorSteps,
		xGrid)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(int(event.StartTime), int(event.EventType)) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}

func main() {

	// Default to no error reason coding. Applies only to the conversion of CSV
	// files to binary log format. TODO: Document filter conf format.
	flag.StringVar(
		&errorReasonFilterConf,
		"error-reason-filter",
		"",
		"Error reason filter congfiguration.")

	// Default to showing all events.
	flag.IntVar(
		&typeFilter,
		"event-type-id",
		-1,
		"Event type ID to filter for.")

	// Default to the beginning of the epoch.
	flag.IntVar(
		&minTime,
		"min-time",
		0,
		"Least recent time to show, expressed as seconds in Unix epoch time.")

	// Default to now.
	flag.IntVar(
		&maxTime,
		"max-time",
		int(time.Now().Unix()),
		"Most recent time to show, expressed as seconds in Unix epoch time.")

	// Default to no grid. No effect on conversion of CSV files to binary log
	// format.
	flag.IntVar(
		&xGrid,
		"x-grid",
		0,
		"Number of divisions to be separated with vertical grid lines.")

	// Default to covering a power of two seconds run-time for every 16 pixels
	// on the y axis. No effect on conversion of CSV files to binary log format.
	flag.Float64Var(
		&yLog2,
		"run-time-scale",
		16,
		"Pixels along y-axis for every doubling in seconds of run time.")

	// Default to 256 pixels. May not apply for all graph types.
	flag.IntVar(
		&width,
		"width",
		256,
		"Width of the rendered graph, in pixels.")

	// Default to 128 pixels. May not apply for all graph types.
	flag.IntVar(
		&height,
		"height",
		128,
		"Height of the rendered graph, in pixels.")

	// Default to just one. May not apply for all graph types.
	flag.IntVar(
		&colorSteps,
		"color-steps",
		1,
		"Number of color steps to use in rendering before clipping.")

	flag.Parse()

	if flag.NArg() != 3 {
		log.Println("Incorrect argument count.")
		os.Exit(1)
	}

	var (
		action         = flag.Arg(0)
		inputFilePath  = flag.Arg(1)
		outputFilePath = flag.Arg(2)
	)

	if action == "csv-convert" {
		convertCommaSeparatedToBinary(inputFilePath, outputFilePath)
	} else if action == "vis-error-stack" {
		generateErrorStackVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-histogram" {
		generateHistogramVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-rolling-stack" {
		generateRollingStackVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-scatter" {
		generateScatterVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-status-stack" {
		generateStatusStackVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-sweep" {
		generateSweepVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-wave" {
		generateWaveVisualization(inputFilePath, outputFilePath)
	} else if action == "vis-wave-sorted" {
		generateSortedWaveVisualization(inputFilePath, outputFilePath)
	} else {
		log.Println("Unrecognized action.")
		os.Exit(1)
	}
}
