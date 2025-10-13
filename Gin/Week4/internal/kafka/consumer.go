package kafka

import (
	"context"
	"encoding/json"

	"news-aggregator/internal/models"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader: reader,
	}
}

func (c *Consumer) ConsumeNews(ctx context.Context, handler func(article models.NewsArticle) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				logrus.WithError(err).Error("Failed to read message from Kafka")
				continue
			}

			var article models.NewsArticle
			if err := json.Unmarshal(message.Value, &article); err != nil {
				logrus.WithError(err).Error("Failed to unmarshal news article")
				continue
			}

			if err := handler(article); err != nil {
				logrus.WithError(err).Error("Failed to process news article")
				continue
			}

			logrus.WithField("article_id", article.ID).Info("Successfully processed news article")
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
