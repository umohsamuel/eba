package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/umohsamuel/eba/pkg/database"
	"github.com/umohsamuel/eba/pkg/logger"
	"github.com/umohsamuel/eba/services/notification/internal/adapters/db/repository"
	"github.com/umohsamuel/eba/services/notification/internal/adapters/email"
	"github.com/umohsamuel/eba/services/notification/internal/adapters/kafka"
	"github.com/umohsamuel/eba/services/notification/internal/adapters/templates"
	"github.com/umohsamuel/eba/services/notification/internal/services/notification"
	"github.com/umohsamuel/eba/services/notification/pkg/env"
)

func main() {
	envCfg, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	slog.SetDefault(logger.New("notification", envCfg.AppEnv, envCfg.LogLevel))

	if err := run(envCfg); err != nil {
		slog.Error("notification service stopped", "error", err)
		os.Exit(1)
	}
}

func run(envCfg *env.Env) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbSrv := database.New(database.Config{
		Host:     envCfg.DB.DB_HOST,
		Port:     envCfg.DB.DB_PORT,
		User:     envCfg.DB.DB_USER,
		Password: envCfg.DB.DB_PASSWORD,
		Database: envCfg.DB.DB_DATABASE,
		Schema:   envCfg.DB.DB_SCHEMA,
		SSLMode:  envCfg.DB.DB_SSLMODE,
	})
	defer dbSrv.Close()

	renderer, err := templates.New(envCfg.AppName)
	if err != nil {
		return err
	}

	sender, err := email.NewSMTPSender(email.SMTPConfig{
		Host:      envCfg.SMTP.SMTP_HOST,
		Port:      envCfg.SMTP.SMTP_PORT,
		Username:  envCfg.SMTP.SMTP_USERNAME,
		Password:  envCfg.SMTP.SMTP_PASSWORD,
		From:      envCfg.SMTP.SMTP_FROM,
		TLSPolicy: envCfg.SMTP.SMTP_TLS_POLICY,
	})
	if err != nil {
		return err
	}

	svc := notification.New(notification.Config{
		Repository: repository.NewNotificationRepository(dbSrv.DB()),
		Renderer:   renderer,
		Email:      sender,
	})

	brokers := envCfg.Kafka.KAFKA_BROKERS

	if err := kafka.EnsureTopics(brokers, kafka.AllTopics, envCfg.Kafka.KAFKA_TOPIC_PARTITIONS, envCfg.Kafka.KAFKA_TOPIC_REPLICATION); err != nil {
		return err
	}

	producer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		return err
	}
	defer producer.Close()

	router := kafka.NewRetryRouter(producer, kafka.RetryTopics, kafka.TopicDLQ)
	handle := kafka.HandleNotificationRequested(svc)

	type stage struct {
		group string
		topic string
		delay time.Duration
	}

	stages := []stage{{group: kafka.GroupMain, topic: kafka.TopicRequested}}
	for i, topic := range kafka.RetryTopics {
		stages = append(stages, stage{group: kafka.RetryGroups[i], topic: topic, delay: envCfg.Kafka.KAFKA_RETRY_DELAYS[i]})
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(stages))

	for _, st := range stages {
		consumer, err := kafka.NewKafkaConsumer(brokers, st.group)
		if err != nil {
			return err
		}

		handler := kafka.NewConsumerGroupHandler(handle, router, st.delay)

		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("consuming", "topic", st.topic, "group", st.group, "delay", st.delay)
			if err := consumer.Run(ctx, []string{st.topic}, handler); err != nil {
				errs <- err
				stop()
			}
		}()
	}

	slog.Info("notification service started")
	<-ctx.Done()
	wg.Wait()
	close(errs)

	for err := range errs {
		return err
	}

	slog.Info("notification service stopped")
	return nil
}
