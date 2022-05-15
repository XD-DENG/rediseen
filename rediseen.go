package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

func getPidFileDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(userHomeDir, ".rediseen")
}

func getPidFilePath() string {
	return path.Join(getPidFileDir(), "rediseen.pid")
}

func savePID(pid int, fileForPid string) error {
	pidFileDir := getPidFileDir()
	if _, err := os.Stat(pidFileDir); os.IsNotExist(err) {
		os.Mkdir(pidFileDir, 0766)
	}

	f, err := os.Create(fileForPid)
	if err != nil {
		return fmt.Errorf("unable to create PID file: %s", err.Error())
	}
	defer f.Close()

	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("unable to write to PID file : %s", err.Error())
	}

	f.Sync()
	return nil
}

func stopDaemon(fileForPid string) error {
	if rawPid, err := ioutil.ReadFile(fileForPid); err == nil {
		pid, err := strconv.Atoi(string(rawPid))
		if err != nil {
			return fmt.Errorf("invalid PID found in %s", fileForPid)
		}

		err = os.Remove(fileForPid)
		if err != nil {
			return fmt.Errorf("unable to remove PID file (error: %s)", err.Error())
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("unable to find PID %d (error: %s)", pid, err.Error())
		}

		err = process.Kill()
		if err != nil {
			return fmt.Errorf("unable to kill process %d (error: %s)", pid, err.Error())
		}
	} else {
		return errors.New("no running service found")
	}
	return nil
}

func main() {
	var daemonMode bool
	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Rediseen service",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(strHeader)
			log.Println("[INFO] Daemon mode:", daemonMode)

			var s service
			var pidFile = getPidFilePath()

			err := s.loadConfigFromEnv()
			if err != nil {
				fmt.Println("[ERROR] " + err.Error())
				return
			}

			log.Printf("[INFO] Serving at %s", s.bindAddress)

			if daemonMode {
				// check if daemon is already running
				if _, err := os.Stat(pidFile); err == nil {
					fmt.Println(fmt.Sprintf("[ERROR] Resideen is already running, or file %s exists", pidFile) +
						" (delete the file only if you are sure there is no running Rediseen instance).")
					os.Exit(1)
				}

				cmd := exec.Command(os.Args[0], "start")
				err = cmd.Start()
				if err != nil {
					fmt.Println("[ERROR] " + err.Error())
					return
				}
				log.Println("[INFO] Running in daemon. PID:", cmd.Process.Pid)
				err = savePID(cmd.Process.Pid, pidFile)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				os.Exit(0)
			}

			http.Handle("/", &s)

			serve := http.ListenAndServe(s.bindAddress, nil)
			if serve != nil {
				log.Println("[ERROR] Failed to launch. Details: ", serve.Error())
			}
		},
	}

	cmdStart.Flags().BoolVarP(&daemonMode, "daemon-mode", "d", false, "run in background")

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop the service (running in the background)",
		Run: func(cmd *cobra.Command, args []string) {
			var pidFile = getPidFilePath()
			err := stopDaemon(pidFile)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Service running in daemon is stopped.")
			}
		},
	}

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display the version of Rediseen",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Rediseen Version:\t%s\nGo Runtime Version:\t%s\n", rediseenVersion, runtime.Version())
		},
	}

	var cmdConfigDoc = &cobra.Command{
		Use:   "configdoc",
		Short: "Display the full configuration help documentation",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(strHeader)
			fmt.Println(strHelpDoc)
		},
	}

	var rootCmd = &cobra.Command{Use: "rediseen"}
	rootCmd.AddCommand(cmdStart, cmdStop, cmdVersion, cmdConfigDoc)
	rootCmd.Execute()
}
