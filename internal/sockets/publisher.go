package zmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kubex-ecosystem/gobe/internal/config"

	zmq "github.com/pebbe/zmq4"
)

type Publisher struct {
	socket  *zmq.Socket
	config  config.ZMQConfig
	address string
}

func NewPublisher(config config.ZMQConfig) *Publisher {
	return &Publisher{
		config:  config,
		address: fmt.Sprintf("%s:%d", config.Address, config.Port),
	}
}

func (p *Publisher) Connect() error {
	socket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return fmt.Errorf("failed to create ZMQ socket: %w", err)
	}

	err = socket.Bind(p.address)
	if err != nil {
		socket.Close()
		return fmt.Errorf("failed to bind ZMQ socket to %s: %w", p.address, err)
	}

	p.socket = socket
	log.Printf("ZMQ Publisher connected to %s", p.address)
	return nil
}

func (p *Publisher) PublishMessage(topic string, data interface{}) error {
	if p.socket == nil {
		if err := p.Connect(); err != nil {
			return err
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	message := fmt.Sprintf("%s %s", topic, string(jsonData))
	_, err = p.socket.Send(message, 0)
	if err != nil {
		return fmt.Errorf("failed to send ZMQ message: %w", err)
	}

	log.Printf("Published ZMQ message: %s", topic)
	return nil
}

func (p *Publisher) Close() error {
	if p.socket != nil {
		return p.socket.Close()
	}
	return nil
}
