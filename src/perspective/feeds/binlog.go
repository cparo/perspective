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
	"bufio"
	"encoding/binary"
	"image/png"
	"perspective"
)

// This struct is written to pack neatly into a 64-byte line while still
// accommodating any data we will realistically be pulling out of out event
// database in the next couple of decades. This may not matter much for
// performance, but it is pretty convenient for reading a hex dump of the
// resulting binary event log format.
type eventData struct {
	EventID   int32 // Event ID as recorded in reference data source.
	StartTime int32 // In seconds since the beginning of the Unix epoch.
	RunTime   int32 // Event run time, in seconds.
	EventType int16 // Event type ID as recorded in reference data source.
	Status    int16 // Zero indicates success, non-zero indicates failure.
}

// GeneratePNGFromBinLog reads a binary-log formatted event-data dump and
// renders a visualization as a PNG file using the specified visualization
// generator and input-filtering parameters.
func GeneratePNGFromBinLog(
	iPath string,
	oPath string,
	minTime int,
	maxTime int,
	typeFilter int,
	v perspective.Visualizer) {

	iFile, oFile := openFiles(iPath, oPath)

	binReader := bufio.NewReader(iFile)

	for {

		var event eventData
		err := binary.Read(binReader, binary.LittleEndian, &event)

		if atEOF(err, "Error reading event data from binary log.") {
			break
		}

		if eventFilter(
			int(event.StartTime),
			int(event.EventType),
			minTime,
			maxTime,
			typeFilter) {
			v.Record(
				perspective.EventDataPoint{
					event.StartTime,
					event.RunTime,
					event.Status})
		}
	}

	png.Encode(oFile, v.Render())
}
