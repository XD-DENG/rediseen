package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	//"os"
)

func main() {
	fmt.Println(strHeader)

	var command = flag.Bool("d", false, "Run in daemon mode")
	flag.Parse()

	fmt.Println("Daemon mode:", *command)

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println(strUsage)
		os.Exit(0)
	}

	switch args[0] {
	case "start":
		err := configCheck()
		if err != nil {
			fmt.Println("[ERROR] " + err.Error())
			return
		}

		http.HandleFunc("/", service)

		addr := generateAddr()
		log.Printf("Serving at %s", addr)
		serve := http.ListenAndServe(addr, nil)
		if serve != nil {
			panic(serve)
		}
	case "help":
		fmt.Println(strHelpDoc)
	case "version":
		fmt.Println(rediseenVersion)
	default:
		fmt.Println(strUsage)
	}
}
