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

func convertCommaSeparatedToBinary(iPath string, oPath string) {
	feeds.ConvertCSVToBinary(
		iPath,
		oPath,
		minTime,
		maxTime,
		typeFilter,
		errorClassConf)
}

func generateErrorStackVisualization(iPath string, oPath string) {
	v := perspective.NewErrorStack(width, height)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateHistogramVisualization(iPath string, oPath string) {
	v := perspective.NewHistogram(width, height, yLog2)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateRollingStackVisualization(iPath string, oPath string) {
	v := perspective.NewRollingStack(width, height, minTime, maxTime)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateScatterVisualization(iPath string, oPath string) {
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

func generateSortedWaveVisualization(iPath string, oPath string) {
	v := perspective.NewSortedWave(width, height, minTime, maxTime)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateStatusStackVisualization(iPath string, oPath string) {
	v := perspective.NewStatusStack(width, height)
	feeds.GeneratePNGFromBinLog(iPath, oPath, minTime, maxTime, typeFilter, v)
}

func generateSweepVisualization(iPath string, oPath string) {
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

func generateWaveVisualization(iPath string, oPath string) {
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
