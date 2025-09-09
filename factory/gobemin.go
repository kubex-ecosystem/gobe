// Package factory fornece funções para criar instâncias do GoBE e seus componentes.
package factory

import (
	"fmt"
	"log"
	"time"

	t "github.com/kubex-ecosystem/gdbase/types"
	gb "github.com/kubex-ecosystem/gobe"
	s "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
	"github.com/streadway/amqp"
)

type GoBE interface {
	ci.IGoBE
}

type DBConfig = t.DBConfig

var (
	dbConfig *DBConfig
)

func NewGoBE(name, port, bind, logFile, configFile string, isConfidential bool, logger l.Logger, debug, releaseMode bool) (ci.IGoBE, error) {
	err := initRabbitMQ()
	if err != nil {
		return nil, err
	}
	goBe, err := gb.NewGoBE(name, port, bind, logFile, configFile, isConfidential, logger, debug, releaseMode)
	if err != nil {
		return nil, err
	}
	dbService, err := GetDatabaseService(goBe)
	if err != nil {
		return nil, err
	}
	if dbService != nil {
		dbConfig = dbService.GetConfig()
	}

	return goBe, nil
}

var rabbitMQConn *amqp.Connection

func initRabbitMQ() error {
	var err error
	url := getRabbitMQURL()
	if url != "" {
		rabbitMQConn, err = amqp.Dial(url)
		if err != nil {
			log.Printf("Erro ao conectar ao RabbitMQ: %s", err)
			return err
		}
		if rabbitMQConn == nil {
			return fmt.Errorf("RabbitMQ connection is not initialized")
		}
		log.Println("Conexão com RabbitMQ estabelecida com sucesso.")
	}
	return nil
}

func getRabbitMQURL() string {
	if dbConfig != nil {
		if dbConfig.Messagery != nil {
			if dbConfig.Messagery.RabbitMQ != nil {
				return fmt.Sprintf("amqp://%s:%s@%s:%d/",
					dbConfig.Messagery.RabbitMQ.Username,
					dbConfig.Messagery.RabbitMQ.Password,
					dbConfig.Messagery.RabbitMQ.Host,
					dbConfig.Messagery.RabbitMQ.Port,
				)
			}
		}
	}
	return ""
}

func closeRabbitMQ() {
	if rabbitMQConn != nil {
		rabbitMQConn.Close()
		log.Println("Conexão com RabbitMQ encerrada.")
	}
}

func ConsumeMessages(queueName string) {
	url := getRabbitMQURL()
	if url == "" {
		log.Printf("RabbitMQ URL is not configured")
		return
	}
	conn, err := amqp.Dial(url)
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
	url := getRabbitMQURL()
	if url == "" {
		log.Printf("RabbitMQ URL is not configured")
		return fmt.Errorf("RabbitMQ URL is not configured")
	}
	conn, err := amqp.Dial(url)
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
