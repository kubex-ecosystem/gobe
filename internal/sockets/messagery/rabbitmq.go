// Package messagery provides AMQP messaging functionality.
package messagery

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"

	t "github.com/kubex-ecosystem/gdbase/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"

	crtSvc "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
)

type DBConfig = t.DBConfig

type AMQP struct {
	URL   string
	Conn  *amqp.Connection
	Chan  *amqp.Channel
	ready atomic.Bool
}

func (a *AMQP) Connect(ctx context.Context, url string, logf func(string, ...any)) error {
	a.URL = url
	backoff := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second, 30 * time.Second}
	var last error
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		conn, err := amqp.Dial(url)
		if err == nil {
			ch, err := conn.Channel()
			if err != nil {
				_ = conn.Close()
				last = err
			}
			if err == nil {
				a.Conn, a.Chan = conn, ch
				if err := a.declareTopology(); err != nil {
					_ = ch.Close()
					_ = conn.Close()
					last = err
				}
				a.ready.Store(true)
				logf("amqp conectado e pronto")
				go a.watch(ctx, logf) // reconectar em caso de close
				return nil
			}
		} else {
			last = err
		}
		d := backoff[min(i, len(backoff)-1)]
		logf("amqp falhou: %v; retry em %s", last, d)
		time.Sleep(d)
	}
}

func (a *AMQP) watch(ctx context.Context, logf func(string, ...any)) {
	errs := a.Conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-errs:
			a.ready.Store(false)
			if e != nil {
				logf("amqp closed: %v", e)
			}
			_ = a.Chan.Close()
			_ = a.Conn.Close()
			// tenta reconectar
			_ = a.Connect(ctx, a.URL, logf)
			return
		}
	}
}

func (a *AMQP) declareTopology() error {
	// exchanges/queues/bindings idempotentes
	// ex:
	// return a.Chan.ExchangeDeclare("events", "topic", true, false, false, false, nil)
	return nil
}

func (a *AMQP) PublishReliable(exchange, key string, body []byte) error {
	if !a.ready.Load() {
		return errors.New("amqp not ready")
	}
	if err := a.Chan.Confirm(false); err != nil {
		return err
	}
	confirms := a.Chan.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := a.Chan.Publish(exchange, key, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	}); err != nil {
		return err
	}
	c := <-confirms
	if !c.Ack {
		return errors.New("publish nack")
	}
	return nil
}

func GetRabbitMQURL(dbConfig *DBConfig) string {
	var host = ""
	var port = ""
	var username = ""
	var password = ""
	if dbConfig.Messagery.RabbitMQ.Host != "" {
		host = dbConfig.Messagery.RabbitMQ.Host
	} else {
		host = "localhost"
	}
	if dbConfig.Messagery.RabbitMQ.Port != "" {
		strPort, ok := dbConfig.Messagery.RabbitMQ.Port.(string)
		if ok {
			port = strPort
		} else {
			gl.Log("error", "RabbitMQ port is not a string")
			port = "5672"
		}
	} else {
		port = "5672"
	}
	if dbConfig.Messagery.RabbitMQ.Username != "" {
		username = dbConfig.Messagery.RabbitMQ.Username
	} else {
		username = "gobe"
	}
	if dbConfig.Messagery.RabbitMQ.Password != "" {
		password = dbConfig.Messagery.RabbitMQ.Password
	} else {
		rabbitPassKey, rabbitPassErr := crtSvc.GetOrGenPasswordKeyringPass("rabbitmq")
		if rabbitPassErr != nil {
			gl.Log("error", "Skipping RabbitMQ setup due to error generating password")
			gl.Log("debug", fmt.Sprintf("Error generating key: %v", rabbitPassErr))
			goto postRabbit
		}
		password = string(rabbitPassKey)
	}

	if host != "" && port != "" && username != "" && password != "" {
		return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", username, password, host, port, "gobe")
	}
postRabbit:
	return ""
}
