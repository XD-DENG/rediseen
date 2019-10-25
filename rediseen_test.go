package main

import (
	"log"
	"os"
	"testing"
)

func Test_stopDaemon_no_pid_file(t *testing.T) {
	err := stopDaemon("/tmp/non-existing")

	compareAndShout(t, "Not running", err.Error())
}

func Test_stopDaemon_invalid_pid(t *testing.T) {

	testPidFileLocation := "/tmp/temptesting.pid"

	file, err := os.Create(testPidFileLocation)
	if err != nil {
		log.Printf("Unable to create PID file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString("invalid PID")
	if err != nil {
		log.Printf("Unable to create PID file : %v\n", err)
		os.Exit(1)
	}

	file.Sync()

	err = stopDaemon(testPidFileLocation)

	compareAndShout(t, "Invalid PID found in "+testPidFileLocation, err.Error())
}
