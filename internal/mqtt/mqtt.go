package mqtt

import (
	"context"
	"fmt"
	"github.com/kube-orchestra/maestro/internal/db"
	"k8s.io/klog/v2"
	"open-cluster-management.io/work/pkg/clients/mqclient"
	"open-cluster-management.io/work/pkg/clients/workclient"
	"open-cluster-management.io/work/pkg/clients/workclient/protocal/mqtt"
	"os"
)

const (
	mqttClientID       = "MQTT_CLIENT_ID"
	mqttBrokerURL      = "MQTT_BROKER_URL"
	mqttBrokerUsername = "MQTT_BROKER_USERNAME"
	mqttBrokerPassword = "MQTT_BROKER_PASSWORD"
)

type Connection struct {
	mqClient        *mqclient.MessageQueueClient[*db.Resource]
	ResourceChannel chan db.Resource
}

func NewConnection(ctx context.Context) *Connection {
	opts, err := NewClientOpts()
	if err != nil {
		panic(err)
	}

	subClient, pubClient, err := opts.GetClients(ctx, "")

	mqClient := mqclient.NewMessageQueueClient[*db.Resource](
		"maestro",
		subClient,
		pubClient,
		&ResourceLister{},
		&ResourceStatusHashGetter{},
	)
	mqClient.WithEventDataHandler(workclient.ManifestGVR, &Encoder{}, &Decoder{})

	resourceChan := make(chan db.Resource)
	return &Connection{
		mqClient:        mqClient,
		ResourceChannel: resourceChan,
	}
}

func (c *Connection) StartSender(ctx context.Context) {
	go func() {
		for msg := range c.ResourceChannel {
			eventType := mqclient.CloudEventType{
				GroupVersionResource: workclient.ManifestGVR,
				SubResource:          mqclient.SubResourceSpec,
				Action:               mqclient.CreateRequestAction,
			}
			// assume consumer ID here is the cluster ID
			ctxGetter := mqtt.NewMQTTContextGetter()
			pubCtx := ctxGetter.Context(ctx, msg.ConsumerId, eventType)
			err := c.mqClient.Publish(pubCtx, eventType, &msg)
			if err != nil {
				klog.Errorf("failed to publish message with err %v", err)
			}
		}
	}()
}

func (c *Connection) StartStatusReceiver(ctx context.Context) {
	err := c.mqClient.Subscribe(ctx)
	if err != nil {
		klog.Errorf("failed to subscribe")
		return
	}

	go func() {
		for evt := range c.mqClient.SubscriptionResultChan() {
			err := db.SetStatusResource(evt.Object.Id, evt.Object.Status.ContentStatus)
			if err != nil {
				panic(err)
			}
		}
	}()
}

func NewClientOpts() (*mqtt.MQTTClientOptions, error) {
	clientID := os.Getenv(mqttClientID)
	if len(clientID) == 0 {
		return nil, fmt.Errorf("%s must be set", mqttClientID)
	}

	brokerURL := os.Getenv(mqttBrokerURL)
	if len(brokerURL) == 0 {
		return nil, fmt.Errorf("%s must be set", mqttBrokerURL)
	}

	brokerUsername := os.Getenv(mqttBrokerUsername)
	if len(brokerUsername) == 0 {
		return nil, fmt.Errorf("%s must be set", mqttBrokerUsername)
	}

	brokerPassword := os.Getenv(mqttBrokerPassword)
	if len(brokerPassword) == 0 {
		return nil, fmt.Errorf("%s must be set", mqttBrokerPassword)
	}

	opts := mqtt.NewMQTTClientOptions()
	opts.BrokerHost = brokerURL
	opts.Username = brokerUsername
	opts.Password = brokerPassword

	return opts, nil
}
