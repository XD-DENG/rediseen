package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func Test_stopDaemon_no_pid_file(t *testing.T) {
	err := stopDaemon("/tmp/non-existing")

	compareAndShout(t, "Not running", err.Error())
}

func Test_stopDaemon_invalid_pid(t *testing.T) {

	testPidFileLocation := "/tmp/temptesting.pid"
	defer os.Remove(testPidFileLocation)

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

func Test_stopDaemon_normal(t *testing.T) {

	testPidFileLocation := "/tmp/temptesting.pid"
	defer os.Remove(testPidFileLocation)

	cmd := exec.Command("sleep", "30")
	cmd.Start()
	savePID(cmd.Process.Pid, testPidFileLocation)

	err := stopDaemon(testPidFileLocation)

	if err != nil {
		t.Error("Expecting nil \ngot\n", err)
	}

}
