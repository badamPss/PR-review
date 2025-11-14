package main

import (
	"flag"
	"pr-review/internal/app"
)

var (
	configPath = flag.String("config", "./config/local/config.yaml", "path to config file")
)

func main() {
	flag.Parse()

	appl := app.New(*configPath)
	appl.Run()
}
