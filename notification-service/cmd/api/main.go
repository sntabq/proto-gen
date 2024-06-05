package main

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"log/slog"
	"notification-service/config"
	"notification-service/internal/data/dto"
	"notification-service/mailer"
	"notification-service/sl"
	"os"
)

const (
	envLocal = "local" // локальный запуск. Используем удобный для консоли TextHandler и уровень логирования Debug (будем выводить все сообщения).
	envDev   = "dev"   // удаленный dev-сервер. Уровень логирования тот же, но формат вывода — JSON, удобный для систем сбора логов вроде Kibana или Grafana Loki.
	envProd  = "prod"  // продакшен. Повышаем уровень логирования до Info, чтобы не выводить дебаг-логи в проде. То есть мы будем получать сообщения только с уровнем Info или Error.
)

func main() {
	cfg := config.LoadConfig()
	logger := setupLogger(cfg.Env)
	logger.Info("config setup correct")
	m := mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender)
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"shop", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	log.Printf("queue name")
	log.Printf(q.Name)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			body := d.Body
			logger.Debug("received message! ", body)
			data := make(map[string]json.RawMessage)
			err := json.Unmarshal(body, &data)
			if err != nil {
				logger.Error("failed to unmarshal body: ", sl.Err(err))
				return
			}
			var orderDTO dto.OrderDTO
			err = json.Unmarshal(data["order_info"], &orderDTO)
			if err != nil {
				logger.Error("failed to unmarshal order info", sl.Err(err))
				return
			}
			logger.Debug(fmt.Sprintf("unmarshalled orderDTO %v", orderDTO))

			var userDTO dto.UserDTO
			err = json.Unmarshal(data["user_info"], &userDTO)
			if err != nil {
				logger.Error("failed to unmarshal user info", sl.Err(err))
				return
			}
			logger.Debug(fmt.Sprintf("unmarshalled userDTO %v", userDTO))
			messageData := map[string]any{
				"itemName":  orderDTO.Item.Name,
				"username":  userDTO.Username,
				"itemImage": orderDTO.Item.ImageURL,
			}

			logger.Debug(fmt.Sprintf("message data %v", messageData))
			err = m.Send(userDTO.Email, "user_welcome.tmpl", messageData, logger)
			if err != nil {
				logger.Error("failed to send mail ", sl.Err(err))
				return
			}
		}
	}()

	logger.Info("[*] Waiting for messages")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := sl.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
