package main

import (
	"gobox/app"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) == 4 && os.Args[1] == "run" && os.Args[2] == "-c" {
		app := app.NewApp(os.Args[3])
		app.Run()
		return
	}
	log.Printf("usage: %s run -c <config file>", os.Args[0])
}
