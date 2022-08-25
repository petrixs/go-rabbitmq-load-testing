package main

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
	"strconv"
	"time"
)

func main() {

	rabbitMQHost := flag.String("h", "localhost", "rabbitMQ host, by default is localhost")
	rabbitMQUser := flag.String("u", "rabbitmq", "rabbitMQ host, by default is rabbitmq")
	rabbitMQPass := flag.String("p", "rabbitmq", "rabbitMQ password, by default rabbitmq")
	rabbitMQExchange := flag.String("e", "", "rabbitMQ exchange name, by default empty")
	rabbitMQRoutingKey := flag.String("r", "", "rabbitMQ routing key, by default empty")
	rabbitMQVhost := flag.String("v", "", "rabbitMQ vhost, by default empty")

	totalMessages := flag.Int("t", 1, "Total messages to send, by default is 1")
	messagesPerSecond := flag.Int("s", 1, "Maximum requests per second, by Default is 1")
	messageToSend := flag.String("m", "{}", "Message to send, by default is {}")
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	rabbitMQHostURI := "amqp://" + *rabbitMQUser + ":" + *rabbitMQPass + "@" + *rabbitMQHost
	if *rabbitMQVhost != "" {
		rabbitMQHostURI = rabbitMQHostURI + "/" + *rabbitMQVhost
	}

	log.Info().Str("host", rabbitMQHostURI).Send()

	publisher, err := rabbitmq.NewPublisher(rabbitMQHostURI, rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging)

	if err != nil {
		log.Fatal().AnErr("error", err).Send()
	}

	defer func() {
		err := publisher.Close()
		if err != nil {
			log.Fatal().AnErr("error", err).Send()
		}
	}()

	sendMessagesToRabbitMQ(publisher, rabbitMQExchange, rabbitMQRoutingKey, messageToSend, totalMessages, messagesPerSecond)

}

func sendMessagesToRabbitMQ(publisher *rabbitmq.Publisher, rabbitMQExchange *string, rabbitMQRoutingKey *string, messageToSend *string, totalMessages *int, messagesPerSecond *int) {

	iterator := 0
	startTime := time.Now()

	returns := publisher.NotifyPublish()

	for iterator <= *totalMessages {

		err := publisher.Publish(
			[]byte(*messageToSend),
			[]string{*rabbitMQRoutingKey},
			rabbitmq.WithPublishOptionsExchange(*rabbitMQExchange),
		)

		if err != nil {
			log.Fatal().AnErr("error", err).Send()
		}

		iterator += 1

		if iterator%(*messagesPerSecond) == 0 {

			processed := 0

			for r := range returns {

				processed += 1
				log.Info().Str("processed", strconv.Itoa(processed)).Str("deliveryTag", strconv.FormatUint(r.DeliveryTag, 10)).Send()

				if (processed % (*messagesPerSecond)) == 0 {
					break
				}
			}

			log.Debug().Str("action", "Stop go func").Send()

			startTime = throttleSpeed(startTime)
		}
	}
}

func throttleSpeed(startTime time.Time) time.Time {

	time.Sleep(time.Second - time.Since(startTime))

	return time.Now()
}

