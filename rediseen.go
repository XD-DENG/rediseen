package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

var pidFile = "/tmp/rediseen.pid"

func savePID(pid int, fileForPid string) {

	file, err := os.Create(fileForPid)
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Printf("Unable to create PID file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk
}

func stopDaemon(fileForPid string) error {
	if _, err := os.Stat(fileForPid); err == nil {
		rawPid, err := ioutil.ReadFile(fileForPid)
		if err != nil {
			return errors.New("Not running")
		}
		pid, err := strconv.Atoi(string(rawPid))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid PID found in %s", fileForPid))
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to find PID [%v] (error: %v)\n", pid, err.Error()))
		}

		os.Remove(fileForPid)

		err = process.Kill()
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to kill process [%v] (error: %v)\n", pid, err.Error()))
		} else {
			return errors.New(fmt.Sprintf("Killed process [%v]\n", pid))
		}

	} else {
		return errors.New(fmt.Sprintf("Not running"))
	}
}

func main() {
	fmt.Println(strHeader)

	var daemon = flag.Bool("d", false, "Run in daemon mode")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println(strUsage)
		os.Exit(0)
	}

	switch args[0] {
	case "start":
		fmt.Println("Daemon mode:", *daemon)

		err := configCheck()
		if err != nil {
			fmt.Println("[ERROR] " + err.Error())
			return
		}

		if *daemon {
			// check if daemon already running.
			if _, err := os.Stat(pidFile); err == nil {
				fmt.Println(fmt.Sprintf("[ERROR] Already running or file %s exist.", pidFile))
				os.Exit(1)
			}

			cmd := exec.Command(os.Args[0], args...)
			err = cmd.Start()
			if err != nil {
				fmt.Println("[ERROR] " + err.Error())
				return
			}
			log.Println("[INFO] Running in daemon. PID:", cmd.Process.Pid)
			savePID(cmd.Process.Pid, pidFile)
			os.Exit(0)
		}

		http.HandleFunc("/", service)

		addr := generateAddr()
		log.Printf("[INFO] Serving at %s", addr)
		serve := http.ListenAndServe(addr, nil)
		if serve != nil {
			log.Println("[ERROR] Failed to launch. Details: ", serve.Error())
		}
	case "stop":
		err := stopDaemon(pidFile)
		if err != nil {
			fmt.Println(err.Error())
		}
	case "help":
		fmt.Println(strHelpDoc)
	case "version":
		fmt.Println(rediseenVersion)
	default:
		fmt.Println(strUsage)
	}
}
