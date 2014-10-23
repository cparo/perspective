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
	"log"
	"os"
	"github.com/cparo/perspective"
	"github.com/cparo/perspective/feeds"
	"time"
)

var (
	errorClassConf string  // Optional conf file for error classification.
	typeFilter     int     // Event type to filter for, if non-negative.
	tA             int     // Lower limit of time range to be visualized.
	tΩ             int     // Upper limit of time range to be visualized.
	xGrid          int     // Number of horizontal grid divisions.
	yLog2          float64 // Number of pixels over which elapsed times double.
	w              int     // Visualization width, in pixels.
	h              int     // Visualization height, in pixels.
	colors         int     // The number of color steps before saturation.
	action         string  // Indication of action to be taken.
	iPath          string  // Filesystem path for input.
	oPath          string  // Filesystem path for output.
)

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
		&colors,
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
		convertCSV()
	} else if action == "vis-error-stack" {
		visualize(perspective.NewErrorStack(w, h))
	} else if action == "vis-histogram" {
		visualize(perspective.NewHistogram(w, h, yLog2))
	} else if action == "vis-rolling-stack" {
		visualize(perspective.NewRollingStack(w, h, tA, tΩ))
	} else if action == "vis-scatter" {
		visualize(perspective.NewScatter(w, h, tΩ, tA, yLog2, colors, xGrid))
	} else if action == "vis-status-stack" {
		visualize(perspective.NewStatusStack(w, h))
	} else if action == "vis-sweep" {
		visualize(perspective.NewSweep(w, h, tA, tΩ, yLog2, colors, xGrid))
	} else if action == "vis-wave" {
		visualize(perspective.NewWave(w, h, tA, tΩ))
	} else if action == "vis-wave-sorted" {
		visualize(perspective.NewSortedWave(w, h, tA, tΩ))
	} else {
		log.Println("Unrecognized action.")
		os.Exit(1)
	}
}

func convertCSV() {
	feeds.ConvertCSVToBinary(iPath, oPath, tA, tΩ, typeFilter, errorClassConf)
}

func visualize(v perspective.Visualizer) {
	feeds.GeneratePNGFromBinLog(iPath, oPath, tA, tΩ, typeFilter, v)
}
