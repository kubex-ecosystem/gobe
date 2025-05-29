package factory

import (
	"fmt"
	"log"
	"time"

	l "github.com/faelmori/logz"
	gb "github.com/rafa-mori/gobe"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	"github.com/streadway/amqp"
)

type GoBE interface {
	ci.IGoBE
}

func NewGoBE(name, port, bind, logFile, configFile string, isConfidential bool, logger l.Logger, debug bool) (ci.IGoBE, error) {
	return gb.NewGoBE(name, port, bind, logFile, configFile, isConfidential, logger, debug)
}

var rabbitMQConn *amqp.Connection

func initRabbitMQ() error {
	var err error
	rabbitMQConn, err = amqp.Dial(getRabbitMQURL())
	if err != nil {
		log.Printf("Erro ao conectar ao RabbitMQ: %s", err)
		return err
	}
	log.Println("Conexão com RabbitMQ estabelecida com sucesso.")
	return nil
}

func getRabbitMQURL() string {
	return "amqp://guest:guest@localhost:5672/"
}

func closeRabbitMQ() {
	if rabbitMQConn != nil {
		rabbitMQConn.Close()
		log.Println("Conexão com RabbitMQ encerrada.")
	}
}

func ConsumeMessages(queueName string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("Erro ao conectar ao RabbitMQ: %s", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Erro ao abrir um canal: %s", err)
		return
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Erro ao registrar um consumidor: %s", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Mensagem recebida: %s", d.Body)
			// Processar a mensagem aqui
		}
	}()

	log.Printf("Aguardando mensagens na fila %s. Para sair pressione CTRL+C", queueName)
	<-forever
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			log.Printf("Tentativa %d falhou: %s", i+1, err)
			time.Sleep(sleep)
			continue
		}
		return nil
	}
	return fmt.Errorf("todas as tentativas falharam")
}

func PublishMessageWithRetry(queueName string, message string) error {
	return retry(3, 2*time.Second, func() error {
		return PublishMessage(queueName, message)
	})
}

func PublishMessage(queueName, message string) error {
	conn, err := amqp.Dial(getRabbitMQURL())
	if err != nil {
		log.Printf("Erro ao conectar ao RabbitMQ: %s", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Erro ao abrir um canal: %s", err)
		return err
	}
	defer ch.Close()

	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Printf("Erro ao publicar mensagem: %s", err)
		return err
	}

	log.Printf("Mensagem publicada na fila %s: %s", queueName, message)
	return nil
}
