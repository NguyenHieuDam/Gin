package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(broker, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		topic:  topic,
	}
}

func (p *Producer) PublishNews(ctx context.Context, news interface{}) error {
	message, err := json.Marshal(news)
	if err != nil {
		return fmt.Errorf("failed to marshal news: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("news"),
		Value: message,
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to publish news to Kafka")
		return err
	}

	logrus.Info("Successfully published news to Kafka")
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
