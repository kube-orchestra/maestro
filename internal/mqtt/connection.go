package mqtt

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/go-logr/logr"
)

type ConnectionOptions struct {
	BrokerURLs         []string
	Topic              string
	KeepAlive          time.Duration
	ClientID           string
	Username, Password string
	OnMessage          func(m *paho.Publish)
}

type Connection struct {
	log   logr.Logger
	opts  ConnectionOptions
	errCh chan error
	cm    *autopaho.ConnectionManager
}

func NewConnection(
	log logr.Logger, opts ConnectionOptions,
) *Connection {
	return &Connection{
		log:   log,
		opts:  opts,
		errCh: make(chan error),
	}
}

func (c *Connection) onConnectionUp(
	ctx context.Context, cm *autopaho.ConnectionManager,
	cack *paho.Connack,
) {
	c.log.Info("mqtt connection up")
	if _, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			c.opts.Topic: {QoS: 1},
		},
	}); err != nil {
		c.log.Error(err, "failed to subscribe")
		c.errCh <- err
		return
	}
	c.log.Info("subscribed")
}

func (c *Connection) onConnectError(err error) {
	c.log.Error(err, "connection error")
	c.errCh <- err
}

func (c *Connection) onClientError(err error) {
	c.log.Error(err, "client error")
	c.errCh <- err
}

func (c *Connection) onServerDisconnect(d *paho.Disconnect) {
	if d.Properties != nil {
		c.log.Info(
			"server disconnect",
			"reason", d.Properties.ReasonString)
		return
	}
	c.log.Info(
		"server disconnect",
		"reasonCode", d.ReasonCode)
	// TODO: What now?
}

func (c *Connection) onMessage(
	ctx context.Context, m *paho.Publish) {
	// TODO: Parse, inject into controller manager
	c.log.Info(fmt.Sprintf("received: %s %s", m.Topic, string(m.Payload)))
	c.opts.OnMessage(m)
}

func (c *Connection) Publish(ctx context.Context, m *paho.Publish) error {
	if c.cm == nil {
		return nil
	}
	_, err := c.cm.Publish(ctx, m)
	return err
}

func (c *Connection) Start(ctx context.Context) error {
	brokerURLs, err := parseBrokerURLs(c.opts.BrokerURLs)
	if err != nil {
		return fmt.Errorf("parse broker URLs: %w", err)
	}

	connCtx, connCancel := context.WithCancel(context.Background())
	defer connCancel()
	cfg := autopaho.ClientConfig{
		BrokerUrls: brokerURLs,
		KeepAlive:  uint16(c.opts.KeepAlive.Seconds()),
		OnConnectionUp: func(
			cm *autopaho.ConnectionManager,
			cack *paho.Connack,
		) {
			c.onConnectionUp(connCtx, cm, cack)
		},
		OnConnectError: c.onConnectError,
		ClientConfig: paho.ClientConfig{
			ClientID: c.opts.ClientID,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				c.onMessage(connCtx, m)
			}),
			OnClientError:      c.onClientError,
			OnServerDisconnect: c.onServerDisconnect,
		},
	}
	if len(c.opts.Password) > 0 && len(c.opts.Username) > 0 {
		cfg.SetUsernamePassword(
			c.opts.Username,
			[]byte(c.opts.Password))
	}

	cm, err := autopaho.NewConnection(connCtx, cfg)
	if err != nil {
		return fmt.Errorf("new MQTT connection: %w", err)
	}
	c.cm = cm

	select {
	case <-ctx.Done():
	case err := <-c.errCh:
		c.log.Error(err, "shutting down due to error")
	}
	c.log.Info("shutting down")
	defer c.log.Info("shutdown done")

	disconnectCtx, disconnectCancel := context.WithTimeout(
		connCtx, time.Second)
	defer disconnectCancel()
	_ = cm.Disconnect(disconnectCtx)
	return nil
}

func parseBrokerURLs(
	brokerURLs []string,
) ([]*url.URL, error) {
	var out []*url.URL
	for _, raw := range brokerURLs {
		bURL, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, bURL)
	}
	return out, nil
}
