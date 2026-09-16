package main

import (
	"log"

	"github.com/nitin1chandani/ticketmaster/app"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		log.Fatal(err)
	}
	defer container.Close()
	log.Fatal(container.App.Listen(":" + container.Config.Port))
}
