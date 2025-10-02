// Package factory fornece funções para criar instâncias do GoBE e seus componentes.
package factory

import (
	"context"
	"fmt"
	"time"

	gb "github.com/kubex-ecosystem/gobe"
	s "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	msg "github.com/kubex-ecosystem/gobe/internal/sockets/messagery"
	amqp "github.com/rabbitmq/amqp091-go"
)

type GoBE interface {
	ci.IGoBE
}

var (
	dbConfig  s.DBConfig
	dbService s.DBService
)

func NewGoBE(args gl.InitArgs) (ci.IGoBE, error) {
	err := initRabbitMQ()
	if err != nil {
		return nil, err
	}
	goBe, err := gb.NewGoBE(args, gl.GetLogger("GoBE"))
	if err != nil {
		return nil, err
	}
	dbService, err = GetDatabaseService(goBe)
	if err != nil {
		return nil, err
	}
	if dbService != nil {
		cfg := dbService.GetConfig(context.Background())
		if cfg != nil {
			dbConfig = cfg.GetConfig(context.Background())
		} else {
			return nil, fmt.Errorf("Database config is nil")
		}
	}

	return goBe, nil
}

var rabbitMQConn *amqp.Connection

func initRabbitMQ() error {
	var err error
	url := msg.GetRabbitMQURL(dbService)
	if url != "" {
		rabbitMQConn, err = amqp.Dial(url)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Erro ao conectar ao RabbitMQ: %s", err))
			return err
		}
		if rabbitMQConn == nil {
			return fmt.Errorf("RabbitMQ connection is not initialized")
		}
		gl.Log("info", "Conexão com RabbitMQ estabelecida com sucesso.")
	}
	return nil
}

func ConsumeMessages(queueName string) {
	url := msg.GetRabbitMQURL(dbService)
	if url == "" {
		gl.Log("error", "RabbitMQ URL is not configured")
		return
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao conectar ao RabbitMQ: %s", err))
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao abrir um canal: %s", err))
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
		gl.Log("error", fmt.Sprintf("Erro ao registrar um consumidor: %s", err))
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			gl.Log("debug", fmt.Sprintf("Mensagem recebida: %s", d.Body))
			// Processar a mensagem aqui
		}
	}()

	gl.Log("debug", fmt.Sprintf("Aguardando mensagens na fila %s. Para sair pressione CTRL+C", queueName))
	<-forever
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			gl.Log("error", fmt.Sprintf("Tentativa %d falhou: %v", i+1, err))
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
	url := msg.GetRabbitMQURL(dbService)
	if url == "" {
		gl.Log("error", "RabbitMQ URL is not configured")
		return fmt.Errorf("RabbitMQ URL is not configured")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao conectar ao RabbitMQ: %s", err))
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao abrir um canal: %s", err))
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
		gl.Log("error", fmt.Sprintf("Erro ao publicar mensagem: %s", err))
		return err
	}

	gl.Log("info", fmt.Sprintf("Mensagem publicada na fila %s: %s", queueName, message))
	return nil
}

func GetDatabaseService(goBE ci.IGoBE) (s.DBService, error) {
	if goBE == nil {
		return nil, fmt.Errorf("GoBE instance is nil")
	}
	dbService := goBE.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil")
		return nil, nil
	}
	return dbService, nil
}
