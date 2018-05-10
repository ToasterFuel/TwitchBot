package main

import (
	"fmt"
	"net/http"
	"os"
)

var config configInformation

func main() {
	var err error
	config, err = readConfigInformation()
	if err != nil {
		fmt.Println("An error happened. ", err)
	}

	if len(os.Args) == 1 {
		fmt.Println("Yo, give some arguments")
		return
	}

	if os.Args[1] == "server" {
		http.HandleFunc("/callbacks/", handleCallback)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("Could not start server", err)
		}
	} else if os.Args[1] == "callbacks" {
		subscribeToFollowerNotifications("randomId1", 14400)
		subscribeToScreamUpDownNotifications("randomId2", 14400)
	} else {
		fmt.Println("Nothing to do with command: ", os.Args[1])
	}
}
