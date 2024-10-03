package main

import (
	"SNS_Server/server"
	"runtime"
	"time"
)

func timeToGC() {
	for {
		runtime.GC()
		time.Sleep(15 * time.Second)
	}
}

func main() {
	go timeToGC()
	Server := server.NewServer("0.0.0.0", 5195)
	Server.Start()
}
