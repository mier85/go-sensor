// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package main

import (
	"time"

	instana "github.com/mier85/go-sensor"
)

func main() {
	go forever()
	select {}
}

func forever() {
	for {
		instana.SendServiceEvent("go-microservice-14c",
			"Go Microservice Online", "♪┏(°.°)┛┗(°.°)┓┗(°.°)┛┏(°.°)┓ ♪", instana.SeverityChange, 1*time.Second)
		time.Sleep(30000 * time.Millisecond)
	}
}
