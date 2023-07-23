package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kube-orchestra/maestro/internal/db"
)

type Connection struct {
	Client          mqtt.Client
	ResourceChannel chan db.ResourceMessage
}

func NewConnection() *Connection {
	client := NewClient()
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	resourceChan := make(chan db.ResourceMessage)
	return &Connection{
		Client:          client,
		ResourceChannel: resourceChan,
	}
}

func (c *Connection) StartSender() {
	go func() {
		for msg := range c.ResourceChannel {
			topic := fmt.Sprintf("v1/%s/%s/content", msg.ConsumerId, msg.Id)
			msgJson, _ := json.Marshal(msg)
			token := c.Client.Publish(topic, 1, false, msgJson)
			token.Wait()
		}
	}()
}

func (c *Connection) StartStatusReceiver() {
	c.Client.Subscribe("v1/+/+/status", 1, messagePubHandler)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topicComponents := strings.Split(msg.Topic(), "/")

	err := db.SetStatusResource(topicComponents[2], msg.Payload())
	if err != nil {
		panic(err)
	}
}

func NewClient() mqtt.Client {
	// mqtt.ERROR = log.New(os.Stdout, "E: ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "C: ", 0)
	// mqtt.WARN = log.New(os.Stdout, "W: ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "D: ", 0)

	var broker = "localhost"
	var port = 31320
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("maestro")
	opts.SetUsername("admin")
	opts.SetPassword("password")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	return client
}
