package feeds

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/cparo/perspective"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// PrintEventLog pretty-prints binary event log data.
func PrintEventLog(
	iPath string,
	minTime int32,
	maxTime int32,
	typeFilter int16,
	errorClassConf string) {

	filterString := "[blank]"
	errorFilters := []string{filterString}
	if errorClassConf != "" {
		cFile, err := os.Open(errorClassConf)
		panicOnError(err, "Failed to open error-reason filter config file.")
		confReader := csv.NewReader(bufio.NewReader(cFile))
		// Filter conf file is designed to look nicely tabular in plain text,
		// so it has a pipe field delimiter and extra white space.
		confReader.Comma = '|'
		for {
			fields, err := confReader.Read()
			if atEOF(err, "Error encountered consuming filter config.") {
				break
			}
			// NOTE: We assume a filter configuartion format which has pipe
			//       separated fields and keeps a human-friendly description of
			//       the error case in the second field.
			if len(fields) < 2 {
				panic("Incorrect field count in filter config.")
			}
			filterString = strings.TrimSpace(fields[1])
			errorFilters = append(errorFilters, filterString)
		}
		cFile.Close()
	}
	errorFilters = append(errorFilters, "[other]")

	iFile, err := os.Open(iPath)
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

	for i, _ := range *events {
		e := (*perspective.EventData)(unsafe.Pointer(&(*events)[i]))
		if eventFilter(e, minTime, maxTime, typeFilter) {
			fmt.Printf("Event ID: %d\n", e.ID)
			fmt.Printf(
				"  Start Time: %s\n",
				time.Unix(int64(e.Start), int64(0)).String())
			fmt.Printf("  Run Time:   %d\n", e.Run)
			fmt.Printf("  Type ID:    %d\n", e.Type)
			if e.Status > 0 {
				fmt.Printf("  Status:     %s\n\n", errorFilters[e.Status-1])
			} else if e.Status < 0 {
				fmt.Printf("  Status:     [in progress]")
			}
		}
	}
}
