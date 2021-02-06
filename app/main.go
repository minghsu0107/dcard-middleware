package main

import (
	"log"
)

func main() {
	container := BuildContainer()
	err := container.Invoke(func(server *Server) {
		server.Run()
	})
	if err != nil {
		log.Fatal(err)
	}
}
