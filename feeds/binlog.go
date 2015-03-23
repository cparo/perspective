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
	"encoding/binary"
	"fmt"
	"github.com/cparo/perspective"
	"image/png"
	"io"
	"log"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

// DumpEventData reads a binary-log formatted event-data dump and writes out a
// listing of the data in the event records which match the specified filtering
// criteria. These values are written as all int32 values for the sake of making
// the output easier to consume with such things as a JavaScript Typed Array
// parser (which lacks native support for such concepts as c-style structs).
func DumpEventData(
	events *[]perspective.EventData,
	tA int32,
	tΩ int32,
	typeFilter int,
	regionFilter int,
	statusFilter int,
	out io.Writer) {

	for i, _ := range *events {
		e := (*perspective.EventData)(unsafe.Pointer(&(*events)[i]))
		if eventFilter(e, tA, tΩ, typeFilter, regionFilter, statusFilter) {
			binary.Write(out, binary.LittleEndian, int32(e.ID))
			binary.Write(out, binary.LittleEndian, int32(e.Start))
			binary.Write(out, binary.LittleEndian, int32(e.Run))
			binary.Write(out, binary.LittleEndian, int32(e.Type))
			binary.Write(out, binary.LittleEndian, int32(e.Status))
			binary.Write(out, binary.LittleEndian, int32(e.Region))
			binary.Write(out, binary.LittleEndian, int32(e.Progress))
		}
	}
}

// GeneratePNGFromBinLog reads a binary-log formatted event-data dump and
// renders a visualization as a PNG file using the specified visualization
// generator and input-filtering parameters.
func GeneratePNGFromBinLog(
	events *[]perspective.EventData,
	tA int32,
	tΩ int32,
	typeFilter int,
	regionFilter int,
	statusFilter int,
	v perspective.Visualizer,
	out io.Writer) {

	// Passing event data by reference instead of passing it by value cuts about
	// 12-15% off of run time in repeated before/after tests with the scatter
	// visualization through the HTTP API.
	for i, _ := range *events {
		e := (*perspective.EventData)(unsafe.Pointer(&(*events)[i]))
		if eventFilter(e, tA, tΩ, typeFilter, regionFilter, statusFilter) {
			v.Record(e)
		}
	}

	png.Encode(out, v.Render())
}

// GetSuccessRate reads a binary-log formatted event-data dump and writes out
// the rate of successful event completions relative to all event completions
// within the specified time range and event type filter criteria, encoded as
// a string percentage value of up to five places (like "99.997%").
func GetSuccessRate(
	events *[]perspective.EventData,
	tA int32,
	tΩ int32,
	typeFilter int,
	regionFilter int,
	out io.Writer) {

	var (
		pass  = 0
		total = 0
	)
	for i, _ := range *events {
		e := (*perspective.EventData)(unsafe.Pointer(&(*events)[i]))
		if eventFilter(e, tA, tΩ, typeFilter, regionFilter, 4) {
			pass++
		}
		if eventFilter(e, tA, tΩ, typeFilter, regionFilter, 6) {
			total++
		}
	}
	if total > 0 {
		fmt.Fprintf(out, "%.3f%%", 100*float64(pass)/float64(total))
	} else {
		fmt.Fprint(out, "NaN%")
	}
}

func MapBinLogFile(path string, lookback int64) *[]perspective.EventData {

	iFile, err := os.Open(path)
	if err != nil {
		log.Println("Failed to open input file for reading.")
		return nil
	}

	defer iFile.Close()

	iStat, err := iFile.Stat()
	if err != nil {
		log.Println("Failed to stat input file.")
		return nil
	}

	fileSize := iStat.Size()

	var start, length int64
	if lookback > 0 && lookback < fileSize {
		start = fileSize - lookback
		// Round down start position to fall on an even page boundary so the
		// mmap will succeed:
		start = start - start % int64(syscall.Getpagesize())
		length = fileSize - start
	} else {
		start, length = 0, fileSize
	}

	binLog, err := syscall.Mmap(
		int(iFile.Fd()),
		start,
		int(length),
		syscall.PROT_READ,
		syscall.MAP_PRIVATE)
	if err != nil {
		log.Println("Failed to mmap input file.")
		return nil
	}

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

func UnmapBinLogFile(eventData *[]perspective.EventData) error {

	mapping := (*[]byte)(unsafe.Pointer(eventData))

	header := (*reflect.SliceHeader)(unsafe.Pointer(eventData))
	header.Len *= int(unsafe.Sizeof(perspective.EventData{}))
	header.Cap *= int(unsafe.Sizeof(perspective.EventData{}))

	return syscall.Munmap(*mapping)
}
