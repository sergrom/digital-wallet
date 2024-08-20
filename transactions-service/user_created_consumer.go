package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

type ucConsumerGroupHandler struct{}

func (ucConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ucConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h ucConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var ucMsg UserCreatedMsg
		err := json.Unmarshal(msg.Value, &ucMsg)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = dbConn.Query(`
				insert into "user" ("user_id", "balance", "created_at") values ($1, 0, $2)
				on conflict do nothing`,
			ucMsg.UserID, ucMsg.CreatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

func consumeUserCreated() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup([]string{kafkaAddr}, kafkaConsumerGroupID, config)
	if err != nil {
		log.Printf("error creating consumer group client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	for {
		err := consumerGroup.Consume(ctx, []string{kafkaUserCreatedTopic}, ucConsumerGroupHandler{})
		if err != nil {
			log.Printf("Error from consumer: %v\n", err)
		}
	}
}
