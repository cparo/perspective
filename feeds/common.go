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
	"log"
)

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

func panicOnError(err error, message string) {
	if err != nil {
		log.Println(err)
		panic(message)
	}
}
