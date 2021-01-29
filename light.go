package main

import (
	"fmt"
	"log"
	"strings"
)

type light struct {
	Name  string
	Relay uint8

	uid    string
	status bool
	board  *Board
}

func newLight(name string, relay uint8, board *Board) (*light, error) {
	if relay < 0 || relay > 8 {
		return nil, fmt.Errorf("Relay #%d is not available. Must be between 1 and 8", relay)
	}
	uid := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	return &light{Name: name, Relay: relay, uid: uid, board: board, status: board.getRelay(relay)}, nil
}

func (lt *light) stateTopic() string {
	return fmt.Sprintf("%s/light/%s/state", myConfig.MQTT.Topic, lt.uid)
}

func (lt *light) commandTopic() string {
	return fmt.Sprintf("%s/light/%s/set", myConfig.MQTT.Topic, lt.uid)
}

func (lt *light) stateValue() string {
	if lt.status {
		return "On"
	}
	return "Off"
}

func (lt *light) pubData() pubData {
	return pubData{"light/" + lt.uid + "/state", lt.stateValue()}
}

func (lt *light) process(cmd, val string) {
	if cmd != "set" {
		log.Printf("Unhandled command '%s' received. Ignoring...", cmd)
		return
	}
log.Printf("Light: %s cmd -> %s, current state %s", lt.Name, val, lt.stateValue())
	switch strings.ToLower(val) {
	case "on":
		lt.status = lt.board.setRelay(lt.Relay, true)
	case "off":
		lt.status = lt.board.setRelay(lt.Relay, false)
	default:
		log.Printf("Unknown command value received '%s'", val)
		return
	}
	pubChannel <- pubData{lt.stateTopic(), lt.stateValue()}
}

func (lt *light) hassYaml() string {
	return fmt.Sprintf("  - platform: mqtt\n    name: %s\n    state_topic: \"%s\"\n    command_topic: \"%s\"\n    payload_on: \"On\"\n    payload_off: \"Off\"\n    unique_id: %s\n\n",
		lt.Name,
		lt.stateTopic(),
		lt.commandTopic(),
		strings.ToUpper(lt.uid))
}
