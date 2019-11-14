package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Test_savePID(t *testing.T) {
	testPidFileLocation := "/xxx/yyy.pid"
	err := savePID(100, testPidFileLocation)
	compareAndShout(t, "unable to create PID file", strings.Split(err.Error(), ":")[0])
}

func Test_stopDaemon_no_pid_file(t *testing.T) {
	err := stopDaemon("/tmp/non-existing")

	compareAndShout(t, "no running service found", err.Error())
}

func Test_stopDaemon_invalid_pid(t *testing.T) {

	testPidFileLocation := "/tmp/temptesting.pid"
	defer os.Remove(testPidFileLocation)

	file, err := os.Create(testPidFileLocation)
	if err != nil {
		log.Printf("unable to create PID file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString("invalid PID")
	if err != nil {
		log.Printf("unable to create PID file : %v\n", err)
		os.Exit(1)
	}

	file.Sync()

	err = stopDaemon(testPidFileLocation)

	compareAndShout(t, "invalid PID found in "+testPidFileLocation, err.Error())
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

	process, _ := os.FindProcess(cmd.Process.Pid)
	status, _ := process.Wait()
	compareAndShout(t, "signal: killed", status.String())
}

func Test_Main(t *testing.T) {
	// First element "" is a placeholder for executable
	//ref: https://stackoverflow.com/a/48674736

	originalValue := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6400")
	defer os.Setenv("REDISEEN_REDIS_URI", originalValue)

	// command "rediseen"
	os.Args = []string{""}
	main()

	// commands "rediseen version", "rediseen help", "rediseen stop", "rediseen wrong_command"
	for _, command := range []string{"version", "help", "stop", "wrong_command"} {
		os.Args = []string{"", command}
		main()
	}
}

func Test_Main_invalid_config(t *testing.T) {
	// First element "" is a placeholder for executable
	//ref: https://stackoverflow.com/a/48674736

	// command "rediseen start" with invalid configuration
	originalValue := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "invalid_url")
	defer os.Setenv("REDISEEN_REDIS_URI", originalValue)

	os.Args = []string{"", "start"}
	main()
}

func Test_Main_start_command(t *testing.T) {
	originalValue := os.Getenv("REDISEEN_REDIS_URI")
	os.Setenv("REDISEEN_REDIS_URI", "redis://:@localhost:6400")
	defer os.Setenv("REDISEEN_REDIS_URI", originalValue)

	os.Args = []string{"", "-d", "start"}
	main()
}
