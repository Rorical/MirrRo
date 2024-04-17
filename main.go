package main

import (
	"flag"
	"github.com/Rorical/MirrRo/service"
	"log"
)

func main() {
	// read flags
	var cfgPath string
	flag.StringVar(&cfgPath, "c", "config.json", "Config")
	flag.Parse()

	err := service.Listen(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
}
