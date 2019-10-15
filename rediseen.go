package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8000"

func main() {

	if len(os.Args) != 2 {
		fmt.Println(strLogo)
		fmt.Println(strHeader)
		fmt.Println(strUsage)
		os.Exit(0)
	}

	var command = os.Args[1]

	switch command {
	case "start":

		fmt.Println(strLogo)
		fmt.Println(strHeader)

		err := configCheck()
		if err != nil {
			fmt.Println("[ERROR] " + err.Error())
			return
		}

		http.HandleFunc("/", service)

		port := os.Getenv("REDISEEN_PORT")
		if port == "" {
			port = defaultPort
		}
		log.Printf("Running with port %s", port)
		serve := http.ListenAndServe(":"+port, nil)
		if serve != nil {
			panic(serve)
		}
	case "help":
		fmt.Println(strLogo)
		fmt.Println(strHelpDoc)
	case "version":
		fmt.Println(rediseenVersion)
	default:
		fmt.Println(strLogo)
		fmt.Println(strHeader)
		fmt.Println(strUsage)
	}
}
