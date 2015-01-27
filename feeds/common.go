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
	"log"
)

func eventFilter(
	event *perspective.EventData,
	minTime int32,
	maxTime int32,
	typeFilter int,
	statusFilter int) bool {
	if minTime < event.Start && maxTime > event.Start {
		if typeFilter < 0 || int(event.Type) == typeFilter {
			if event.Status == 0 && 4&statusFilter != 0 {
				return true // Done
			}
			if event.Status > 0 && 2&statusFilter != 0 {
				return true // Failed
			}
			if event.Status < 0 && 1&statusFilter != 0 {
				return true // Running
			}
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
