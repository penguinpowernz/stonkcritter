package sinks

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/penguinpowernz/stonkcritter/models"
)

// MQTT will create a new sink that can be used to send the full disclosure object
// as JSON over the given MQTT URL (e.g. localhost:1883) and creds (e.g. user:pass
// or empty for unauthed), to the given topic.
func MQTT(url, creds, topic string) (Sink, error) {
	if !strings.Contains(url, ":") {
		url += ":1883"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", url))
	opts.SetClientID("stonkcritter")

	if strings.Contains(creds, ":") {
		bits := strings.Split(creds, ":")
		user := bits[0]
		pass := bits[1]
		opts.SetUsername(user)
		opts.SetPassword(pass)
	}

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	return MQTTWithOptions(opts, topic)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
}

// MQTTWithOptions will create a sync with the provided options.  It will send to the given topic.
func MQTTWithOptions(opts *mqtt.ClientOptions, topic string) (Sink, error) {
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return func(d models.Disclosure) error {
		t := client.Publish(topic, 0, false, d.Bytes())
		<-t.Done()
		err := t.Error()
		logerr(err, "mqtt", "while publishing")
		return err
	}, nil
}
