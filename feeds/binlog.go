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
	"image/png"
	"github.com/cparo/perspective"
	"syscall"
	"unsafe"
)

// This struct is written to pack neatly into a 64-byte line while still
// accommodating any data we will realistically be pulling out of our reference
// database in the next couple of decades. This may not matter much for
// performance, but it is pretty convenient for reading a hex dump of the
// resulting binary event log format.
type eventData struct {
	ID     int32 // Event ID as recorded in reference data source.
	Start  int32 // In seconds since the beginning of the Unix epoch.
	Run    int32 // Event run time, in seconds.
	Type   int16 // Event type ID as recorded in reference data source.
	Status int16 // Zero indicates success, non-zero indicates failure.
}

// GeneratePNGFromBinLog reads a binary-log formatted event-data dump and
// renders a visualization as a PNG file using the specified visualization
// generator and input-filtering parameters.
func GeneratePNGFromBinLog(
	iPath string,
	oPath string,
	tA int,
	tΩ int,
	typeFilter int,
	v perspective.Visualizer) {

	iFile, oFile := openFiles(iPath, oPath)

	iStat, err := iFile.Stat()
	exitOnError(err, "Failed to stat input file.")

	iSize := int(iStat.Size())

	binLog, err := syscall.Mmap(
		int(iFile.Fd()),
		0,
		iSize,
		syscall.PROT_READ,
		syscall.MAP_PRIVATE)
	exitOnError(err, "Failed to mmap input file.")

	// Using this mmap-and-cast method of parsing the input log instead of the
	// more idiomatic use of Go's bufio and encoding/binary packages for reading
	// the input log into eventData structs yields a sixfold improvement in run
	// time and CPU cost in testing against a 45-MiB log of reference event
	// data. When removing the actual rendering of graph data in a test run with
	// each log-file reading implementation, the measured performance gain is
	// 42-fold. The difference is 80-fold if we also remove the encoding of the
	// blank image canvas to a png file. Which should help to illustrate the
	// absurd cost of avoiding an "unsafe" method for reading a file which would
	// be considered perfectly valid in traditional systems development.
	events := *(*[]eventData)(unsafe.Pointer(&binLog))
	n := iSize / int(unsafe.Sizeof(eventData{}))
	for i := 0; i < n; i++ {
		e := events[i]
		if eventFilter(int(e.Start), int(e.Type), tA, tΩ, typeFilter) {
			v.Record(perspective.EventDataPoint{e.Start, e.Run, e.Status})
		}
	}

	png.Encode(oFile, v.Render())
}
