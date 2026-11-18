package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"buenvivir/notifications/internal/config"
	"buenvivir/notifications/internal/services"
)

func main() {
	cfg := config.LoadConfig()
	if cfg.ResendAPIKey == "" {
		log.Fatal("Error: RESEND_API_KEY environment variable is not configured in .env or system environment")
	}

	emailService, err := services.NewDualResendEmailService(cfg.ResendAPIKey, cfg.ResendFromNotifications, cfg.ResendFromAlerts)
	if err != nil {
		log.Fatalf("Error initializing email service: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	msg := services.EmailMessage{
		To:      []string{"leonardomorabito02@gmail.com"},
		Subject: "TESTING SUBJECT",
		Html:    "<strong>TESTING</strong>",
	}

	id, err := emailService.SendEmail(ctx, msg)
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	fmt.Printf("Email sent successfully. Resend ID: %s\n", id)
}
