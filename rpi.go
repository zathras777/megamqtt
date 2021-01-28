package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

var doOnce sync.Once

func setupRPi() error {
	if myConfig.Megaio.Pin < 0 || myConfig.Megaio.Pin > 40 {
		return fmt.Errorf("Invalid pin specified, %d. Must be between 0 and 40", myConfig.Megaio.Pin)
	}

	doOnce.Do(func() {
		host.Init()
	})

	pinString := fmt.Sprintf("GPIO%d", myConfig.Megaio.Pin)
	log.Printf("RaspberryPi: Setting up pin %d", myConfig.Megaio.Pin)
	rPin := gpioreg.ByName(pinString)
	if rPin == nil {
		return fmt.Errorf("Unable to listen on RaspberryPi pin #%d", myConfig.Megaio.Pin)
	}

	err := rPin.In(gpio.PullDown, gpio.BothEdges)
	if err != nil {
		return fmt.Errorf("Unable to set pin #%d to be an Input. %s", myConfig.Megaio.Pin, err)
	}
	log.Printf("Configured Pin #%d for input. Current value %v\n", myConfig.Megaio.Pin, rPin.Read())

	go func() {
		for {
			if rPin.WaitForEdge(-1) {
				rpiActivity <- true
				time.Sleep(700 * time.Millisecond)
			} else {
				log.Print("WaitForEdge() returned false")
			}
		}
	}()

	return nil
}
