package main

import (
	"flag"
	"log"
	"os"
	"perspective"
	"perspective/feeds"
	"time"
)

// Variables for command-line option flags.
var (
	errorClassConf string
	typeFilter     int
	minTime        int
	maxTime        int
	xGrid          int
	yLog2          float64
	width          int
	height         int
	colorSteps     int
)

// Variables for fixed-position command-line arguments.
var (
	action string // Indication of what type of visualization to generate.
	iPath  string // Filesystem path for input
	oPath  string // Filesystem path for output
)

func convertCommaSeparatedToBinary() {
	feeds.ConvertCSVToBinary(
		iPath,
		oPath,
		minTime,
		maxTime,
		typeFilter,
		errorClassConf)
}

func generateErrorStackVisualization() {
	v := perspective.NewErrorStack(width, height)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateHistogramVisualization() {
	v := perspective.NewHistogram(width, height, yLog2)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateRollingStackVisualization() {
	v := perspective.NewRollingStack(width, height, minTime, maxTime)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateScatterVisualization() {
	v := perspective.NewScatter(
		width,
		height,
		minTime,
		maxTime,
		yLog2,
		colorSteps,
		xGrid)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateSortedWaveVisualization() {
	v := perspective.NewSortedWave(width, height, minTime, maxTime)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateStatusStackVisualization() {
	v := perspective.NewStatusStack(width, height)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateSweepVisualization() {
	v := perspective.NewSweep(
		width,
		height,
		minTime,
		maxTime,
		yLog2,
		colorSteps,
		xGrid)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateWaveVisualization() {
	v := perspective.NewWave(width, height, minTime, maxTime)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func main() {

	// Default to no error reason coding. Applies only to the conversion of CSV
	// files to binary log format. TODO: Document filter conf format.
	flag.StringVar(
		&errorClassConf,
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

	action = flag.Arg(0)
	iPath = flag.Arg(1)
	oPath = flag.Arg(2)

	if action == "csv-convert" {
		convertCommaSeparatedToBinary()
	} else if action == "vis-error-stack" {
		generateErrorStackVisualization()
	} else if action == "vis-histogram" {
		generateHistogramVisualization()
	} else if action == "vis-rolling-stack" {
		generateRollingStackVisualization()
	} else if action == "vis-scatter" {
		generateScatterVisualization()
	} else if action == "vis-status-stack" {
		generateStatusStackVisualization()
	} else if action == "vis-sweep" {
		generateSweepVisualization()
	} else if action == "vis-wave" {
		generateWaveVisualization()
	} else if action == "vis-wave-sorted" {
		generateSortedWaveVisualization()
	} else {
		log.Println("Unrecognized action.")
		os.Exit(1)
	}
}
