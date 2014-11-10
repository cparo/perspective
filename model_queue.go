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

package perspective

type modelQueue struct {
	q []int32 // Values are completion times
}

type ModelQueue interface {
	Push(int32)
	Step(int32) int
}

// NewModelQueue returns a model queue implementation which can be used for
// simulations/replays of queue behavior given a stream of inputs in the order
// they would have been added to the queue, each inserted as an indication of
// the time it will be removed from the queue.
func NewModelQueue() ModelQueue {
	return &modelQueue{make([]int32, 0, 4096)}
}

// Push takes an expiration time value indicating that an item has been added to
// the queue which will expire out of it at the specified time.
func (this *modelQueue) Push(e int32) {
	// First, insert at the tail...
	this.q = append(this.q, e)
	// Now bubble the newly-inserted event up to the appropriate position so the
	// queue's check/remove/report process can rely on ordering within the queue
	// (this ordering actually behaves more stack than a queue in its internal
	// implementation, since the soonest-to-expire events are put at the tail to
	// minimize churn in element positions)...
	i := len(this.q) - 1
	for i > 0 && this.q[i] > this.q[i-1] {
		eʹ := this.q[i]
		this.q[i] = this.q[i-1]
		this.q[i-1] = eʹ
		i--
	}
}

// Step takes a time value, expires anything with an expiration date before that
// value from the queue, and returns the number of elements left in the queue
// following this expiration process.
func (this *modelQueue) Step(t int32) int {
	i := len(this.q) - 1
	for i >= 0 && this.q[i] <= t {
		this.q = this.q[:i]
		i--
	}
	return len(this.q)
}
