package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func sendEmail(logs string) {

	err := godotenv.Load()
	if err != nil {
		// log.Fatalf irá imprimir o erro e sair do programa.
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	toRaw := os.Getenv("EMAIL_TO")
	subject := "Crawling Logs"
	body := logs

	// Validação básica para garantir que as variáveis foram carregadas
	if from == "" || password == "" || toRaw == "" {
		log.Fatal("Erro: Variáveis de ambiente EMAIL_FROM, EMAIL_PASSWORD ou EMAIL_TO não definidas no .env.")
	}

	// Converte a string de destinatários para um slice (se tiver múltiplos, separe por vírgula no .env)
	to := strings.Split(toRaw, ",")
	// TrimSpace em cada elemento para remover espaços em branco extras
	for i, email := range to {
		to[i] = strings.TrimSpace(email)
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Montando a mensagem
	msg := []byte(
		"From " + from + "\r\n" +
			"To: " + strings.Join(to, ", ") + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
			"\r\n" +
			body,
	)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Enviando o email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		fmt.Printf("Erro sending email: %v\n", err)
	}

	fmt.Println("Email sent with sucess!")
}
