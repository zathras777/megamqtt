package main

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func connectBroker() error {
	mqOpts := mqtt.NewClientOptions().SetClientID("megamqtt")
	mqOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", myConfig.MQTT.Host, myConfig.MQTT.Port))
	mqOpts.OnConnect = connectCb
	mqOpts.SetDefaultPublishHandler(publishCb)

	mqttClient = mqtt.NewClient(mqOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

var connectCb mqtt.OnConnectHandler = func(c mqtt.Client) {
	if token := c.Subscribe(myConfig.MQTT.Topic+"/light/+/set", 1, nil); token.Wait() && token.Error() != nil {
		log.Printf("Unable to subscribe to required topics: %s", token.Error())
		doneChannel <- true
	}
	log.Print("MQTT connected and subscribed OK")
}

var publishCb mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	parts := strings.Split(msg.Topic(), "/")
	lt, ck := lights[parts[2]]
	if !ck {
		log.Printf("Message received for unknown light, %s. Ignoring...", parts[2])
		return
	}
	cmdChannel <- command{lt, parts[3], string(msg.Payload())}
}
