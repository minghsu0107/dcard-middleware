package main

import log "github.com/sirupsen/logrus"

func main() {
	server, err := InitializeServer()
	if err != nil {
		log.Fatal(err)
	}
	server.Run()
}
