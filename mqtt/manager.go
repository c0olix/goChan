package mqtt

import (
	"fmt"
	"github.com/c0olix/goChan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"os"
)

type Manager struct {
	readyClient mqtt.Client
}

func NewManager(config goChan.ConfigInterface) (*Manager, error) {
	conf, ok := config.(ChannelConfig)
	if !ok {
		return nil, errors.New("unable to convert to mqtt channel config")
	}
	manager := Manager{}
	client, err := manager.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new mqtt client")
	}
	manager.readyClient = *client
	if token := manager.readyClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.Wrap(token.Error(), "unable to connect to broker")
	}
	return &manager, nil
}

type ChannelConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Qos      int
}

func (m Manager) CreateChannel(name string, config goChan.ConfigInterface) (goChan.ChannelInterface, error) {
	conf, ok := config.(ChannelConfig)
	if !ok {
		return nil, errors.New("unable to convert to mqtt channel config")
	}
	reader, err := m.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new mqtt client")
	}
	writer, err := m.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create new mqtt client")
	}

	channel := &Channel{
		name:   name,
		qos:    conf.Qos,
		reader: *reader,
		writer: *writer,
	}
	return channel, nil
}

func (m Manager) NewClient(config ChannelConfig) (*mqtt.Client, error) {
	opts := mqtt.NewClientOptions()

	server := fmt.Sprintf("tcp://%s:%d", config.Host, config.Port)
	opts.AddBroker(server)

	clientId, err := m.createMqttClientId()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create client id")
	}
	opts.SetClientID(clientId)

	if config.Username != "" {
		opts.SetUsername(config.Username)
	}
	if config.Password != "" {
		opts.SetUsername(config.Password)
	}
	client := mqtt.NewClient(opts)
	return &client, err
}

func (m Manager) createMqttClientId() (string, error) {
	hostname, _ := os.Hostname()
	aUUID := uuid.New()
	clientId := hostname + aUUID.String()
	return clientId, nil
}

func (m Manager) Ready() error {
	if !m.readyClient.IsConnected() {
		return errors.New("no connection to broker")
	}
	return nil
}
