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
	"io"
	"log"
	"os"
)

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

func eventFilter(
	startTime int,
	eventType int,
	minTime int,
	maxTime int,
	typeFilter int) bool {
	if minTime < startTime && maxTime > startTime {
		if typeFilter < 0 || eventType == typeFilter {
			return true
		}
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

func openFiles(iPath string, oPath string) (iFile *os.File, oFile *os.File) {

	var err error

	iFile, err = os.Open(iPath)
	exitOnError(err, "Failed to open input file for reading.")

	oFile, err = os.Create(oPath)
	exitOnError(err, "Failed to open output file for writing.")

	return iFile, oFile
}
