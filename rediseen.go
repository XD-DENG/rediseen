package main

import (
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

func savePID(pid int) {

	file, err := os.Create(pidFile)
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk
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
				fmt.Println(fmt.Sprintf("Already running or %s file exist.", pidFile))
				os.Exit(1)
			}

			cmd := exec.Command(os.Args[0], args...)
			err = cmd.Start()
			if err != nil {
				fmt.Println("[ERROR] " + err.Error())
				return
			}
			log.Println("[INFO] Running in daemon. PID:", cmd.Process.Pid)
			savePID(cmd.Process.Pid)
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
		if _, err := os.Stat(pidFile); err == nil {
			data, err := ioutil.ReadFile(pidFile)
			if err != nil {
				fmt.Println("Not running")
				os.Exit(1)
			}
			ProcessID, err := strconv.Atoi(string(data))

			if err != nil {
				fmt.Println("Unable to read and parse process id found in ", pidFile)
				os.Exit(1)
			}

			process, err := os.FindProcess(ProcessID)

			if err != nil {
				fmt.Printf("Unable to find process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			}
			// remove PID file
			os.Remove(pidFile)

			fmt.Printf("Killing process ID [%v] now.\n", ProcessID)
			// kill process and exit immediately
			err = process.Kill()

			if err != nil {
				fmt.Printf("Unable to kill process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			} else {
				fmt.Printf("Killed process ID [%v]\n", ProcessID)
				os.Exit(0)
			}

		} else {
			fmt.Println("Not running.")
			os.Exit(1)
		}
	case "help":
		fmt.Println(strHelpDoc)
	case "version":
		fmt.Println(rediseenVersion)
	default:
		fmt.Println(strUsage)
	}
}
