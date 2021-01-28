# megamqtt

This is a small daemon to provide an MQTT interface for a Sequent MegaIO board. It's simple and works for me.

## Configuration File

The configuration is YAML. Valid sections are
- megaio which contains details for connectiong to the MegaIO Board(s)
- mqtt which contains information about the broker to use
- switches which contains a list of switch inputs. Presently each switch is connected to a GPIO input.
- lights which details lights connected to relays.

The app accepts the -cfg flag to determine configuration file location.

## Sample Configuration

    megaio:
      pin: 4

    mqtt:
      host: "127.0.0.1"
      port: 1883
      topic: "megamqtt"

    switches:
      -
        pin: 1
        board: 0
        name: "Switch 1"
      -
        pin: 2
        board: 0
        name: "Switch 2"

    lights:
      -
        relay: 8
        board: 0
        name: First Light

## Home Automation Configuration
My primary use for the daemon is to allow control of lights and switches from a Home Assistant install, so to make life easier the -hass flag can be passed that will print the configuration yaml needed.

    $ ./megamqtt -hass
    switch:
      - platform: mqtt
        name: First Light
        state_topic: "megamqtt/switch/first_light/state"
        command_topic: "megamqtt/light/first_light/set"
        payload_on: "On"
        payload_off: "Off"
        unique_id: FIRST_LIGHT
    
    sensor:
      - platform: mqtt
        name: Switch 1
        state_topic: "megamqtt/switch/switch_1/state"
        unique_id: SWITCH_1
    
      - platform: mqtt
        name: Switch 2
        state_topic: "megamqtt/switch/switch_2/state"
        unique_id: SWITCH_2

## Development
I have a few minor additions and changes to make, but it works for my setup. Bugs, corrections and improvements welcome :-)
