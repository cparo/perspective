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
