package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	lights      map[string]*light
	switches    map[string]*Switch
	cmdChannel  chan command
	pubChannel  chan pubData
	rpiActivity chan bool
	doneChannel chan bool
	mqttClient  mqtt.Client
)

func main() {
	var configFn string
	var hassCfg bool

	flag.StringVar(&configFn, "cfg", "configuration.yaml", "Configuration file to parse")
	flag.BoolVar(&hassCfg, "hass", false, "Generate HASS configuration YAML")
	flag.Parse()

	if err := readConfiguration(configFn); err != nil {
		fmt.Printf("Unable to get configuration: %s\nExiting...\n", err)
		os.Exit(0)
	}

	syslogger, err := syslog.New(syslog.LOG_INFO, "megamqtt")
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(syslogger)

	lights = make(map[string]*light)
	switches = make(map[string]*Switch)
	cmdChannel = make(chan command, 1)
	pubChannel = make(chan pubData, 16)
	rpiActivity = make(chan bool, 1)
	doneChannel = make(chan bool, 1)
	quitChannel := make(chan bool, 1)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	configureLights()
	configureSwitches()

	if hassCfg {
		dumpYamlData()
		os.Exit(0)
	}

	if err = setupRPi(); err != nil {
		fmt.Printf("Unable to setup RaspberryPi: %s\n", err)
		os.Exit(0)
	}
	if err = connectBroker(); err != nil {
		log.Printf("Unable to connect to the MQTT Broker: %s", err)
		os.Exit(0)
	}

	defer mqttClient.Unsubscribe(myConfig.MQTT.Topic + "/light/+/set")
	defer mqttClient.Disconnect(0)

	go func() {
		for {
			select {
			case cmd := <-cmdChannel:
				cmd.light.process(cmd.cmd, cmd.value)
			case pub := <-pubChannel:
				publishState(mqttClient, pub)
			case <-rpiActivity:
				checkSwitches()
			case <-doneChannel:
				quitChannel <- true
				break
			case <-sigs:
				quitChannel <- true
				break
			}
		}
	}()
	<-quitChannel
        log.Print("Exiting...")
}

func publishState(client mqtt.Client, pub pubData) {
        token := client.Publish(pub.topic, byte(0), true, pub.value)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Unable to publish message '%s' to topic '%s': %s", pub.value, pub.topic, token.Error())
	}
}

func checkSwitches() {
	//	log.Printf("checkSwitches()")
	for _, sw := range switches {
		if sw.updateStatus() {
			log.Printf("Change of state for %s [now %s]", sw.Name, sw.stateValue())
			pubChannel <- sw.pubData()
		}
	}
	//	showSwitches()
}

func dumpYamlData() {
	fmt.Println("switch:")
	for _, lt := range lights {
		fmt.Printf(lt.hassYaml())
	}
	fmt.Println("sensor:")
	for _, sw := range switches {
		fmt.Printf(sw.hassYaml())
	}
}
