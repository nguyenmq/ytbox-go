// Common init function for logging

package common

import (
	"io"
	"log"
	"os"
)

const (
	// base path of the log file
	baseLogPath string = "logs"
	// default file open flags
	defaultFlags int = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	// default file perissions
	defaultPerms os.FileMode = 0666
)

/*
 * Initialize the standard logging in a common way across all yt_box.
 *
 * prefix is used to define the name that will show up in the log. useStdOut
 * defines whether to print logging messages to both the log file and to
 * standard out. Remember to defer closing the returned file pointer from where
 * this function is called.
 */
func InitLogger(prefix string, useStdOut bool) *os.File {
	var logPrefix string = "[" + prefix + "] "
	var logPath string = baseLogPath + "/" + prefix + ".log"
	var fil *os.File
	var err error
	var writer io.Writer

	fil, err = os.OpenFile(logPath, defaultFlags, defaultPerms)
	if err != nil {
		log.Fatalf("failed to create real logging file: %v", err)
	}

	if useStdOut {
		writer = io.MultiWriter(fil, os.Stdout)
	} else {
		writer = fil
	}

	log.SetOutput(writer)
	log.SetPrefix(logPrefix)

	return fil
}
