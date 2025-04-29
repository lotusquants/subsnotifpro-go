package service

import (
	"context"
	"fmt"
	"strconv"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/streadway/amqp"
)

func (s *rtdnService) GetDLQSize(ctx context.Context) (int, error) {
	var size int
	err := s.rmqManager.WithAdminChannel(ctx, func(ch *amqp.Channel) error {
		queueInfo, err := ch.QueueInspect(s.cfg.RTDN.DLQ)
		if err != nil {
			return fmt.Errorf("failed to inspect DLQ: %w", err)
		}
		size = queueInfo.Messages
		return nil
	})
	return size, err
}

func (s *rtdnService) RetryMessages(ctx context.Context) error {
	return s.rmqManager.WithAdminChannel(ctx, func(ch *amqp.Channel) error {
		msgs, err := ch.Consume(
			s.cfg.RTDN.DLQ,
			"dlq-retry-worker",
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to consume from DLQ: %w", err)
		}

		go s.processDLQMessages(ctx, ch, msgs)
		return nil
	})
}

func (s *rtdnService) processDLQMessages(ctx context.Context, ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("recovered from panic in processDLQMessages: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return // channel closed
			}

			s.processSingleMessage(ctx, ch, msg)
		}
	}
}

func (s *rtdnService) processSingleMessage(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery) {
	messageID := msg.MessageId
	if messageID == "" {
		messageID = "unknown"
	}

	retryCount := getRetryCount(msg.Headers)
	logFields := map[string]interface{}{
		"message_id":  messageID,
		"retry_count": retryCount,
	}

	if retryCount > s.cfg.MaxRetries {
		logger.Log.WithFields(logFields).Warn("permanently discarding message after exceeding max retries")
		if err := msg.Ack(false); err != nil {
			logger.Log.WithError(err).Error("failed to ack message")
		}
		return
	}

	logger.Log.WithFields(logFields).Info("retrying message from DLQ")

	err := ch.Publish(
		"", // exchange
		s.cfg.RTDN.Queue,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
			Headers:     updateRetryCount(msg.Headers),
		},
	)

	if err != nil {
		logger.Log.WithError(err).Error("failed to republish message")
		time.Sleep(s.cfg.RetryDelay)
		if err := msg.Nack(false, true); err != nil { // requeue
			logger.Log.WithError(err).Error("failed to nack message")
		}
		return
	}

	if err := msg.Ack(false); err != nil {
		logger.Log.WithError(err).Error("failed to ack message")
	}
}

func getRetryCount(headers amqp.Table) int {
	if val, ok := headers["x-retry-count"]; ok {
		if count, err := strconv.Atoi(fmt.Sprintf("%v", val)); err == nil {
			return count
		}
	}
	return 0
}

func updateRetryCount(headers amqp.Table) amqp.Table {
	if headers == nil {
		headers = make(amqp.Table)
	}
	headers["x-retry-count"] = getRetryCount(headers) + 1
	return headers
}
