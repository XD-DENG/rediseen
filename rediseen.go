package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println(strHeader)

	if len(os.Args) != 2 {
		fmt.Println(strUsage)
		os.Exit(0)
	}

	switch os.Args[1] {
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
