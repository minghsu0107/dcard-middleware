package main

import (
	"fmt"
	"log"
)

func main() {
	container := BuildContainer()
	err := container.Invoke(func(r LimiterRepository) {
		if err := r.SetVisitCount("12", 4); err != nil {
			log.Fatal(err)
		}
		var err error
		var a *Record
		var b bool
		a, b, err = r.GetVisitCount("12")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v %v\n", a, b)
	})
	if err != nil {
		log.Fatal(err)
	}
}
