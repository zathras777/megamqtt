package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type yamlConfig struct {
	Megaio struct {
		Pin int
	}
	MQTT struct {
		Host  string `yaml:"host"`
		Port  int    `yaml:"port"`
		Topic string `yaml:"topic"`
	}
	Switches []struct {
		Pin   uint8
		Board uint8
		Name  string
	}
	Lights []struct {
		Relay uint8
		Board uint8
		Name  string
	}
}

type command struct {
	light *light
	cmd   string
	value string
}

type pubData struct {
	topic string
	value string
}

var myConfig yamlConfig

func readConfiguration(cfgFile string) error {
	yamlFile, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlFile, &myConfig)
}

func configureLights() {
	for _, ll := range myConfig.Lights {
		board, err := getBoard(ll.Board)
		if err != nil {
			log.Printf("Unable to setup %s: %s", ll.Name, err)
			continue
		}
		lt, err := newLight(ll.Name, ll.Relay, board)
		if err != nil {
			log.Printf("Error creating Light object from %v: %s", ll, err)
			continue
		}
		lights[lt.uid] = lt
		pubChannel <- lt.pubData()
	}
}

func configureSwitches() {
	for _, ll := range myConfig.Switches {
		board, err := getBoard(ll.Board)
		if err != nil {
			log.Printf("Unable to setup %s: %s", ll.Name, err)
			continue
		}
		lt, err := newSwitch(ll.Name, ll.Pin, board)
		if err != nil {
			log.Printf("Error creating Switch object from %v: %s", ll, err)
			continue
		}
		switches[lt.uid] = lt
		pubChannel <- lt.pubData()
	}
}
