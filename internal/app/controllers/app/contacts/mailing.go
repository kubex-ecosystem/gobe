package contacts

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/smtp"
	"strings"
	"time"

	ci "github.com/rafa-mori/gobe/internal/contracts/interfaces"
	t "github.com/rafa-mori/gobe/internal/contracts/types"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
)

func enqueueEmail(cc *ContactController, emailQueue chan t.ContactForm) {
	if cc.properties == nil {
		gl.Log("error", "Properties not set in contact controller")
		return
	}
	formT, ok := cc.properties["contactForm"]
	if !ok {
		gl.Log("error", "Invalid contact form type")
		return
	}
	form, ok := formT.(*t.Property[t.ContactForm])
	if !ok {
		gl.Log("error", "Invalid contact form type")
		return
	}
	go func(form t.ContactForm) {
		emailQueue <- form
	}(form.GetValue())
}
func processQueue(cc *ContactController, attempts int, emailQueue chan t.ContactForm) {
	for form := range emailQueue {
		go func(f t.ContactForm) {
			if err := sendEmailWithRetry(cc, f, attempts); err != nil {
				gl.Log("error", "Failed to send email after 3 attempts:", err.Error())
			}
		}(form)
	}
}

func getSMTPConfig(env ci.IEnvironment) SMTPConfig {
	host := env.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com" // valor padrão para Gmail
	}
	port := env.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	user := env.Getenv("EMAIL_USR")
	pass := env.Getenv("EMAIL_PWD")
	return SMTPConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
	}
}

func sendEmail(cc *ContactController, form t.ContactForm) error {
	if cc.properties == nil {
		gl.Log("error", "Properties not set in contact controller")
		return errors.New("properties not set in contact controller")
	}

	env, ok := cc.properties["environment"]
	if !ok {
		gl.Log("error", "Environment not set in properties")
		return errors.New("environment not set in properties")
	}
	envT, ok := env.(*t.Property[ci.IEnvironment])
	if !ok {
		gl.Log("error", "Invalid environment type")
		return errors.New("invalid environment type")
	}
	envF := envT.GetValue()

	// Obtém as configurações SMTP parametrizadas
	smtpConfig := getSMTPConfig(envF)
	if smtpConfig.User == "" || smtpConfig.Pass == "" {
		gl.Log("error", "Email user or password not set in environment variables")
		gl.Log("notice", fmt.Sprintf("User: %s", smtpConfig.User))
		gl.Log("notice", fmt.Sprintf("Password: %s", smtpConfig.Pass))
		return errors.New("email user or password not set in environment variables")
	}

	// Montagem dos detalhes do email:
	from := smtpConfig.User
	// Em um cenário real, o destinatário pode ser um campo dinâmico,
	// mas neste exemplo, vamos continuar enviando para o próprio remetente.
	to := []string{smtpConfig.User}

	// Corpo do email
	subject := "PROFILE PAGE - New contact form submission"
	body := fmt.Sprintf("Name: %s\nEmail: %s\nMessage: %s", form.Name, form.Email, form.Message)
	// Cabeçalho: observe o uso de \r\n para compatibilidade com SMTP
	msg := []byte("Subject: " + subject + "\r\n" +
		"From: " + from + "\r\n" +
		"To: " + strings.Join(to, ",") + "\r\n\r\n" + body)

	gl.Log("info", fmt.Sprintf("Sending email contact from %s to %s", form.Email, smtpConfig.User))

	// Autenticação SMTP Padrão:
	auth := smtp.PlainAuth("", smtpConfig.User, smtpConfig.Pass, smtpConfig.Host)

	// Configuração inicial utilizando SendMail (que utiliza STARTTLS automaticamente para a maioria dos servidores na porta 587)
	address := smtpConfig.Host + ":" + smtpConfig.Port
	err := smtp.SendMail(address, auth, from, to, msg)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to send email via %s: %v", smtpConfig.Host, err.Error()))
		return err
	}

	gl.Log("success", "Email sent successfully")
	return nil
}

func sendEmailWithTimeout(cc *ContactController, form t.ContactForm) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Timeout definido
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- sendEmail(cc, form)
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			gl.Log("error", fmt.Sprintf("Timeout error: %v", ctx.Err().Error()))
			return errors.New("error: " + ctx.Err().Error())
		}
	case err := <-errChan:
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error sending email: %v", err.Error()))
			return err // Falha ao enviar
		}
	}

	gl.Log("success", "Email sent successfully within timeout")
	return nil // Sucesso no envio
}

func sendEmailWithRetry(cc *ContactController, form t.ContactForm, attempts int) error {
	var err error
	for attemptsCounter := 0; attemptsCounter < attempts; attemptsCounter++ {
		err = sendEmailWithTimeout(cc, form)
		if err == nil {
			gl.Log("success", fmt.Sprintf("Email sent successfully after %d attempt(s)", attemptsCounter+1))
			return nil // Sucesso
		}
		// Implementa uma estratégia de retry exponencial:
		randomDelay := time.Duration(math.Pow(2, float64(attemptsCounter))) * time.Second
		time.Sleep(randomDelay)
	}
	return fmt.Errorf("failed to send email after %d attempts: %v", attempts, err)
}
