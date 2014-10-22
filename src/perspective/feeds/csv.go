package feeds

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func ConvertCSVToBinary(
	iPath string,
	oPath string,
	minTime int,
	maxTime int,
	typeFilter int,
	errorReasonFilterConf string) {

	// Initial filter is to match for the lack of an error reason string, as
	// signified by an empty or all-whitespace string. This is implied even if
	// we aren't given a configuration file to ensure that we minimally produce
	// output which differentiates errors given with reasons from errors for
	// which no explanation was provided.
	filterString := "^\\s*$"
	filter, err := regexp.Compile(filterString)
	if err != nil {
		log.Printf("Failed to compile regex '%s'.\n", filterString)
		log.Println(err)
		os.Exit(1)
	}
	errorFilters := []*regexp.Regexp{filter}
	if errorReasonFilterConf != "" {
		cFile, err := os.Open(errorReasonFilterConf)
		if err != nil {
			log.Println("Failed to open error-reason filter config file.")
			log.Println(err)
			os.Exit(1)
		}
		confReader := csv.NewReader(bufio.NewReader(cFile))
		// Filter conf file is designed to look nicely tabular in plain text,
		// so it has a pipe field delimiter and extra white space.
		confReader.Comma = '|'
		for {
			fields, err := confReader.Read()
			if atEOF(err, "Error encountered consuming filter config.") {
				break
			}
			// NOTE: We ignore and fields beyond the first here. They can be
			//       parsed out elsewhere for purposes like correlating
			//       human-friendly textual descriptions with the numeric codes
			//       we assign to our output. Ignoring and such additional info
			//       here makes for one less thing that would have to be updated
			//       if we change our minds about what should be provided along
			//       with a list of regex filters in the error-reason filter
			//       config.
			if len(fields) < 1 {
				log.Println("Incorrect field count in filter config.")
				os.Exit(1)
			}
			filterString = strings.TrimSpace(fields[0])
			filter, err = regexp.Compile(filterString)
			if err != nil {
				log.Println("Failed to compile regex '%s'.\n", filterString)
				log.Println(err)
				os.Exit(1)
			}
			errorFilters = append(errorFilters, filter)
		}
		cFile.Close()
	}

	iFile, oFile := openFiles(iPath, oPath)

	csvReader := csv.NewReader(bufio.NewReader(iFile))
	binWriter := bufio.NewWriter(oFile)

	var (
		eventData  eventData
		fieldValue int64
	)

	for {

		fields, err := csvReader.Read()
		if atEOF(err, "Error encountered consuming CSV input.") {
			break
		}
		// INPUT FIELDS:
		// 0) event_id
		// 1) event_type_id
		// 2) event_start_time (in seconds since UNIX epoch)
		// 3) event_run_time (in seconds)
		// 4) exit_status (success if 0, else failure)
		// 5) errror_reason (text field)
		if len(fields) != 6 {
			log.Println("Incorrect field count in CSV input.")
			os.Exit(1)
		}

		fieldValue, err = strconv.ParseInt(fields[1], 10, 16)
		exitOnError(err, "Error encountered parsing event type.")
		eventData.EventType = int16(fieldValue)

		fieldValue, err = strconv.ParseInt(fields[2], 10, 32)
		exitOnError(err, "Error encountered parsing event start time.")
		eventData.StartTime = int32(fieldValue)

		if eventFilter(
			int(eventData.StartTime),
			int(eventData.EventType),
			minTime,
			maxTime,
			typeFilter) {

			fieldValue, err = strconv.ParseInt(fields[0], 10, 32)
			exitOnError(err, "Error encountered parsing event ID.")
			eventData.EventID = int32(fieldValue)

			fieldValue, err = strconv.ParseInt(fields[3], 10, 32)
			exitOnError(err, "Error encountered parsing event run time.")
			eventData.RunTime = int32(fieldValue)

			fieldValue, err = strconv.ParseInt(fields[4], 10, 16)
			exitOnError(err, "Error encountered parsing event status.")
			if fieldValue == 0 {
				eventData.Status = 0
			} else {
				eventData.Status = getErrorCode(fields[5], errorFilters)
			}

			exitOnError(
				binary.Write(binWriter, binary.LittleEndian, eventData),
				"Error writing event data to binary log.")
		}
	}

	exitOnError(
		binWriter.Flush(),
		"Error flushing data to binary log.")
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

func getErrorCode(errorReason string, errorFilters []*regexp.Regexp) int16 {
	var i int
	for i = 0; i < len(errorFilters); i++ {
		if errorFilters[i].MatchString(errorReason) {
			return int16(i + 1)
		}
	}
	// Implied "other" case, which will return a value one past the last value
	// which should be associated with a filter, indicating that no filters
	// matched the errorReason we were given. Note that the error codes start at
	// 1, not 0, so in the example case of our having four error reason filters
	// (including one for a blank error reason), this will be code 5, not 4.
	return int16(i + 1)
}
