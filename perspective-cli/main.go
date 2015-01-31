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

package main

import (
	"flag"
	"github.com/cparo/perspective"
	"github.com/cparo/perspective/feeds"
	"log"
	"os"
	"time"
)

// Mapping of action names to handler functions:
var handlers = make(map[string]func())

// Command-line options and arguments:
var (
	errorClassConf string  // Optional conf file for error classification.
	typeFilter     int     // Event type to filter for, if non-negative.
	regionFilter   int     // Region to filter for, if non-negative.
	statusFilter   int     // Least significant bits: {done, failed, running}.
	tA             int     // Lower limit of time range to be visualized.
	tΩ             int     // Upper limit of time range to be visualized.
	xGrid          int     // Number of horizontal grid divisions.
	yLog2          float64 // Number of pixels over which elapsed times double.
	w              int     // Visualization width, in pixels.
	h              int     // Visualization height, in pixels.
	bg             int     // Graph background color.
	colors         int     // The number of color steps before saturation.
	action         string  // Indication of action to be taken.
	iPath          string  // Filesystem path for input.
	oPath          string  // Filesystem path for output.
)

func init() {

	handlers["csv-convert"] = func() {
		feeds.ConvertCSVToBinary(
			iPath,
			oPath,
			int32(tA),
			int32(tΩ),
			typeFilter,
			regionFilter,
			statusFilter,
			errorClassConf)
	}

	handlers["print-event-log"] = func() {
		feeds.PrintEventLog(
			iPath,
			int32(tA),
			int32(tΩ),
			typeFilter,
			regionFilter,
			statusFilter,
			errorClassConf)
	}

	handlers["vis-error-stack"] = func() {
		visualize(perspective.NewErrorStack(w, h, bg))
	}

	handlers["vis-histogram"] = func() {
		visualize(perspective.NewHistogram(w, h, bg, yLog2))
	}

	handlers["vis-ribbon"] = func() {
		visualize(perspective.NewRibbon(w, h, bg, tA, tΩ))
	}

	handlers["vis-rolling-stack"] = func() {
		visualize(perspective.NewRollingStack(w, h, bg, tA, tΩ))
	}

	handlers["vis-scatter"] = func() {
		visualize(
			perspective.NewScatter(w, h, bg, tA, tΩ, yLog2, colors, xGrid))
	}

	handlers["vis-status-stack"] = func() {
		visualize(perspective.NewStatusStack(w, h, bg))
	}

	handlers["vis-sweep"] = func() {
		visualize(perspective.NewSweep(w, h, bg, tA, tΩ, yLog2, colors, xGrid))
	}

	handlers["vis-wave"] = func() {
		visualize(perspective.NewWave(w, h, bg, tA, tΩ))
	}

	handlers["vis-wave-sorted"] = func() {
		visualize(perspective.NewSortedWave(w, h, bg, tA, tΩ))
	}
}

func main() {

	flag.StringVar(
		&errorClassConf,
		"error-reason-filter",
		"",
		"Error reason filter congfiguration.")

	flag.IntVar(
		&typeFilter,
		"event-type-id",
		-1,
		"Event type ID to filter for.")

	flag.IntVar(
		&regionFilter,
		"region-id",
		-1,
		"Event region ID to filter for.")

	flag.IntVar(
		&statusFilter,
		"status-filter",
		-1,
		"Bitmask for event statuses; LSB are {done,failed,running}.")

	flag.IntVar(
		&tA,
		"min-time",
		0,
		"Least recent time to show, expressed as seconds in Unix epoch time.")

	flag.IntVar(
		&tΩ,
		"max-time",
		int(time.Now().Unix()),
		"Most recent time to show, expressed as seconds in Unix epoch time.")

	flag.IntVar(
		&xGrid,
		"x-grid",
		0,
		"Number of divisions to be separated with vertical grid lines.")

	flag.Float64Var(
		&yLog2,
		"run-time-scale",
		16,
		"Pixels along y-axis for every doubling in seconds of run time.")

	flag.IntVar(
		&w,
		"width",
		256,
		"Width of the rendered graph, in pixels.")

	flag.IntVar(
		&h,
		"height",
		128,
		"Height of the rendered graph, in pixels.")

	flag.IntVar(
		&bg,
		"bg",
		32,
		"Background gray level.")

	flag.IntVar(
		&colors,
		"color-steps",
		1,
		"Number of color steps to use in rendering before clipping.")

	flag.Parse()

	if flag.NArg() != 3 {
		log.Fatalln("Incorrect argument count.")
	}

	action = flag.Arg(0)
	iPath = flag.Arg(1)
	oPath = flag.Arg(2)

	if handler, exists := handlers[action]; exists {
		handler()
	} else {
		log.Fatalln("Unrecognized action.")
	}
}

func visualize(v perspective.Visualizer) {

	out, err := os.Create(oPath)
	if err != nil {
		log.Println("Failed to open output file for writing.")
		log.Fatalln(err)
	}

	eventData := feeds.MapBinLogFile(iPath)
	if eventData == nil {
		log.Fatalln("Failed to parse data feed.")
	}

	feeds.GeneratePNGFromBinLog(
		eventData,
		int32(tA),
		int32(tΩ),
		typeFilter,
		regionFilter,
		statusFilter,
		v,
		out)
}
