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

package feeds

import (
	"github.com/cparo/perspective"
	"image/png"
	"io"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

// GeneratePNGFromBinLog reads a binary-log formatted event-data dump and
// renders a visualization as a PNG file using the specified visualization
// generator and input-filtering parameters.
func GeneratePNGFromBinLog(
	events *[]perspective.EventData,
	tA int32,
	tΩ int32,
	typeFilter int16,
	v perspective.Visualizer,
	out io.Writer) {

	// Passing event data by reference instead of passing it by value cuts about
	// 12-15% off of run time in repeated before/after tests with the scatter
	// visualization through the HTTP API.
	for i, _ := range(*events) {
		e := (*perspective.EventData)(unsafe.Pointer(&(*events)[i]))
		if eventFilter(e.Start, e.Type, tA, tΩ, typeFilter) {
			v.Record(e)
		}
	}

	png.Encode(out, v.Render())
}

func MapBinLogFile(path string) *[]perspective.EventData {

	iFile, err := os.Open(path)
	panicOnError(err, "Failed to open input file for reading.")
	defer iFile.Close()

	iStat, err := iFile.Stat()
	panicOnError(err, "Failed to stat input file.")

	fileSize := int(iStat.Size())

	binLog, err := syscall.Mmap(
		int(iFile.Fd()),
		0,
		fileSize,
		syscall.PROT_READ,
		syscall.MAP_PRIVATE)
	panicOnError(err, "Failed to mmap input file.")

	// Using this mmap-and-cast method of parsing the input log instead of the
	// more idiomatic use of Go's bufio and encoding/binary packages for reading
	// the input log into EventData structs yields a sixfold improvement in run
	// time and CPU cost in testing against a 45-MiB log of reference event
	// data. When removing the actual rendering of graph data in a test run with
	// each log-file reading implementation, the measured performance gain is
	// 42-fold. The difference is 80-fold if we also remove the encoding of the
	// blank image canvas to a png file. Which should help to illustrate the
	// absurd cost of avoiding an "unsafe" method for reading a file which would
	// be considered perfectly valid in traditional systems development.
	events := (*[]perspective.EventData)(unsafe.Pointer(&binLog))

	// Correct the length and capacity of the events slice now that we have
	// re-cast its type, so anything using that slice will know what to iterate
	// over without running past the end.
	header := (*reflect.SliceHeader)(unsafe.Pointer(events))
	header.Len /= int(unsafe.Sizeof(perspective.EventData{}))
	header.Cap /= int(unsafe.Sizeof(perspective.EventData{}))

	return events
}
