// queue/admin.go
package queue

import (
	"context"

	"github.com/streadway/amqp"
)

// WithAdminChannel executes a function with a temporary admin channel
func (rm *RabbitMQManager) WithAdminChannel(ctx context.Context, fn func(ch *amqp.Channel) error) error {
	// Get channel from your existing connection pool
	ch, err := rm.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return fn(ch)
}
