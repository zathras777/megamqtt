package main

import (
	"fmt"
	"strings"
)

// Switch Structure that contains details of a configured switch (GPIO channel)
type Switch struct {
	Pin  uint8
	Name string

	uid     string
	board   *Board
	status  bool
	bitMask uint8
}

func newSwitch(name string, pin uint8, board *Board) (*Switch, error) {
	if pin < 1 || pin > 8 {
		return nil, fmt.Errorf("Invalid pin %d for a switch. Must be between 1 and 8", pin)
	}
	sw := &Switch{Pin: pin, Name: name, board: board}
	sw.uid = strings.ReplaceAll(strings.ToLower(name), " ", "_")
	sw.bitMask = uint8(1 << (pin - 1))
	board.setGPIO(sw.bitMask, true)
	sw.updateStatus()
	return sw, nil
}

func (sw *Switch) stateTopic() string {
	return fmt.Sprintf("%s/switch/%s/state", myConfig.MQTT.Topic, sw.uid)
}

func (sw *Switch) stateValue() string {
	if sw.status {
		return "On"
	}
	return "Off"
}

func (sw *Switch) pubData() pubData {
	return pubData{sw.stateTopic(), sw.stateValue()}
}

func (sw *Switch) updateStatus() bool {
	if status := sw.board.checkGpio(sw.bitMask); status != sw.status {
		sw.status = status
		return true
	}
	return false
}

func (sw *Switch) hassYaml() string {
	return fmt.Sprintf("  - platform: mqtt\n    name: %s\n    state_topic: \"%s\"\n    unique_id: %s\n\n",
		sw.Name,
		sw.stateTopic(),
		strings.ToUpper(sw.uid))
}
