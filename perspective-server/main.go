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
	"fmt"
	"github.com/cparo/perspective"
	"github.com/cparo/perspective/feeds"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const dataPath = "/var/opt/perspective/"

// Mapping of action names to handler functions:
var handlers = make(map[string]func(http.ResponseWriter, *options))

// Mapping of data-source paths to loaded data:
var sources = make(map[string]*[]perspective.EventData)

// Options and arguments:
type options struct {
	typeFilter int     // Event type to filter for, if non-negative.
	tA         int     // Lower limit of time range to be visualized.
	tΩ         int     // Upper limit of time range to be visualized.
	xGrid      int     // Number of horizontal grid divisions.
	yLog2      float64 // Number of pixels over which elapsed times double.
	w          int     // Visualization width, in pixels.
	h          int     // Visualization height, in pixels.
	colors     int     // The number of color steps before saturation.
	feed       string  // Input feed name.
}

func init() {

	handlers["vis-error-stack"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewErrorStack(r.w, r.h), out, r)
	}

	handlers["vis-histogram"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewHistogram(r.w, r.h, r.yLog2), out, r)
	}

	handlers["vis-ribbon"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewRibbon(r.w, r.h, r.tA, r.tΩ), out, r)
	}

	handlers["vis-rolling-stack"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewRollingStack(r.w, r.h, r.tA, r.tΩ), out, r)
	}

	handlers["vis-scatter"] = func(out http.ResponseWriter, r *options) {
		visualize(
			perspective.NewScatter(
				r.w, r.h, r.tA, r.tΩ, r.yLog2, r.colors, r.xGrid),
			out,
			r)
	}

	handlers["vis-status-stack"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewStatusStack(r.w, r.h), out, r)
	}

	handlers["vis-sweep"] = func(out http.ResponseWriter, r *options) {
		visualize(
			perspective.NewSweep(
				r.w, r.h, r.tA, r.tΩ, r.yLog2, r.colors, r.xGrid),
			out,
			r)
	}

	handlers["vis-wave"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewWave(r.w, r.h, r.tA, r.tΩ), out, r)
	}

	handlers["vis-wave-sorted"] = func(out http.ResponseWriter, r *options) {
		visualize(perspective.NewSortedWave(r.w, r.h, r.tA, r.tΩ), out, r)
	}
}

func intOpt(values url.Values, name string, defaultValue int) int {
	strValue := values.Get(name)
	if strValue == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		logMalformedOption(name, strValue)
		return defaultValue
	}
	return intValue
}

func hasUnitSuffix(value string, unit string) (trimmed string, match bool) {
	if strings.HasSuffix(value, unit) {
		return strings.TrimSuffix(value, unit), true
	}
	return value, false
}

func timeOpt(values url.Values, name string, defaultValue int) int {
	strValue := values.Get(name)
	// If no value is specified, fall back to default value...
	if strValue == "" {
		return defaultValue
	}
	// Attempt parsing the time as a human-friendly timestamp...
	// Format is: "YYYY-MM-DD HH:MM:SS Z"
	timeValue, err := time.Parse("2006-01-02 15:04:05 MST", strValue)
	if err == nil {
		return int(timeValue.Unix())
	}
	// Check for leading "-" as indication of time offset backward from NOW()
	// rather than forward from the beginning of the Unix epoch.
	if strings.HasPrefix(strValue, "-") {
		// Trim any trailing "s" from the option-value string, as this is either
		// an unimportant plural indicator (which can be safely discarded in
		// parsing) or an indicator of the seconds unit of time (which can also
		// be safely discarded in parsing, as seconds are the default unit).
		// Note that this would not allow for fractional-second units - but that
		// is okay as the internal assumption is that time is represented in a
		// granularity of seconds and has no smaller fractional unit. A consumer
		// of this graphing system could of course submit time stamps whose
		// integer values represent something else (ms, µs, ns, samples, etc.),
		// record events with these units and display them with a window in
		// these units. This off-label usage would just be a bit of a hack in
		// that internally this system would assume that it is working with
		// second-level epoch times - and might do unexpected things if given a
		// unit-suffixed or human-friendly time value for the display range.
		strValue = strings.TrimSuffix(strValue, "s")
		// Detect and handle any time-unit-specifying suffix...
		var unitSeconds int
		if trimmed, match := hasUnitSuffix(strValue, "year"); match {
			strValue, unitSeconds = trimmed, 31536000 // Assume 365-day year
		} else if trimmed, match := hasUnitSuffix(strValue, "month"); match {
			strValue, unitSeconds = trimmed, 2678400 // Assume 31-day month
		} else if trimmed, match := hasUnitSuffix(strValue, "week"); match {
			strValue, unitSeconds = trimmed, 604800
		} else if trimmed, match := hasUnitSuffix(strValue, "day"); match {
			strValue, unitSeconds = trimmed, 86400
		} else if trimmed, match := hasUnitSuffix(strValue, "h"); match {
			strValue, unitSeconds = trimmed, 3600
		} else if trimmed, match := hasUnitSuffix(strValue, "m"); match {
			strValue, unitSeconds = trimmed, 60
		} else {
			unitSeconds = 1 // Default to a one-second unit
		}
		// Process what is left of the option value, which should be parsable as
		// an integer at this point.
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			logMalformedOption(name, strValue)
			return defaultValue
		}
		return int(time.Now().Unix()) + intValue*unitSeconds
	}
	// Attempt parsing the time value as Unix epoch time in seconds...
	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		logMalformedOption(name, strValue)
		return defaultValue
	}
	return intValue
}

func f64Opt(values url.Values, name string, defaultValue float64) float64 {
	strValue := values.Get(name)
	if strValue == "" {
		return defaultValue
	}
	f64Value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		logMalformedOption(name, strValue)
		return defaultValue
	}
	return f64Value
}

func strOpt(values url.Values, name string, defaultValue string) string {
	strValue := values.Get(name)
	if strValue == "" {
		return defaultValue
	}
	return strValue
}

func logFileLoad(path string) {
	log.Printf("Loading data from file: \"%s\"\n", path)
}

func logMalformedOption(name string, value string) {
	log.Printf(
		"Malformed option: %s = \"%s\", falling back to default.\n",
		name,
		value)
}

func main() {
	http.HandleFunc("/", responder)
	http.ListenAndServe(":8080", nil)
}

func responder(response http.ResponseWriter, request *http.Request) {

	// Parse options, using the same defaults as are used by the CLI interface
	// where options are missing or malformed:
	values := request.URL.Query()
	options := &options{
		intOpt(values, "event-type", -1),
		timeOpt(values, "min-time", 0),
		timeOpt(values, "max-time", int(time.Now().Unix())),
		intOpt(values, "x-grid", 0),
		f64Opt(values, "run-time-scale", 16),
		intOpt(values, "width", 256),
		intOpt(values, "height", 256),
		intOpt(values, "color-steps", 1),
		strOpt(values, "feed", "")}

	action := request.URL.Path[1:]
	if handler, exists := handlers[action]; exists {
		handler(response, options)
	} else {
		msg := fmt.Sprintf("Unrecognized action: %s", action)
		log.Println(msg)
		http.Error(response, msg, 501)
	}
}

func visualize(
	v perspective.Visualizer,
	out http.ResponseWriter,
	r *options) {

	// Load the event data if it is not already loaded.
	// TODO: Some re-work will be needed here to do this in a thread-safe
	//       manner before allowing this server to concurrently handle multiple
	//       requests. Essentially, we will want to make a worker thread which
	//       handles loading of and access to these mapped-file pointers - and
	//       then have our HTTP request handlers send in requests for mapped
	//       pointers by path which will then be asynchronously returned by the
	//       worker (which can either process these requests sequentially or
	//       lock on a per-path basis - in either case safeguarding against
	//       race conditions). Practically speaking this is probably a non-issue
	//       given the narrow time windows involved and invariant nature of the
	//       logs behind these maps once generated - but there is no built-in
	//       provision for safe concurrent manipulation of Go's maps themselves,
	//       so proper defensive practice would be to make this unlikely issue
	//       an impossible one.
	path := dataPath + r.feed + ".dat"
	if _, loaded := sources[path]; !loaded {
		logFileLoad(path)
		defer func() {
			if recovery := recover(); recovery != nil {
				log.Printf(
					"Recovering from internal server error: \"%s\"\n", recovery)
				http.Error(
					out,
					fmt.Sprintf("Internal Server Error: %s", recovery),
					500)
			}
		}()
		sources[path] = feeds.MapBinLogFile(path)
	}

	// This check needs to be repeated as it is possible that we just recovered
	// from a failure to load a new input source.
	if _, loaded := sources[path]; loaded {
		feeds.GeneratePNGFromBinLog(
			sources[path],
			int32(r.tA),
			int32(r.tΩ),
			int16(r.typeFilter),
			v,
			out)
	}
}
