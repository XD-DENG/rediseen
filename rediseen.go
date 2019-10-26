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
	"path"
	"strconv"
)

var daemon = flag.Bool("d", false, "Run in daemon mode")
var pidFile = flag.String("pidfile", path.Join(os.TempDir(), "rediseen.pid"), "where PID is stored for daemon mode")

func savePID(pid int, fileForPid string) error {

	file, err := os.Create(fileForPid)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to create PID file: %s", err.Error()))
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to write to PID file : %s", err.Error()))
	}

	file.Sync() // flush to disk
	return nil
}

func stopDaemon(fileForPid string) error {
	if rawPid, err := ioutil.ReadFile(fileForPid); err == nil {
		pid, err := strconv.Atoi(string(rawPid))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid PID found in %s", fileForPid))
		}

		os.Remove(fileForPid)

		process, err := os.FindProcess(pid)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to find PID %d (error: %s)", pid, err.Error()))
		}

		err = process.Kill()
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to kill process %d (error: %s)", pid, err.Error()))
		} else {
			return nil
		}
	} else {
		return errors.New("no running service found")
	}
}

func main() {
	fmt.Println(strHeader)

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println(strUsage)
		return
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
			if _, err := os.Stat(*pidFile); err == nil {
				fmt.Println(fmt.Sprintf("[ERROR] Already running or file %s exist.", *pidFile))
				os.Exit(1)
			}

			cmd := exec.Command(os.Args[0], args...)
			err = cmd.Start()
			if err != nil {
				fmt.Println("[ERROR] " + err.Error())
				return
			}
			log.Println("[INFO] Running in daemon. PID:", cmd.Process.Pid)
			err = savePID(cmd.Process.Pid, *pidFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
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
		err := stopDaemon(*pidFile)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Service running in daemon is stopped.")
		}
	case "help":
		fmt.Println(strHelpDoc)
	case "version":
		fmt.Println(rediseenVersion)
	default:
		fmt.Println(strUsage)
	}
}
