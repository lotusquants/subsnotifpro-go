package queue

import (
	"log"

	"subsnotifpro-go/internal/constants"

	"github.com/streadway/amqp"
)

// InitializeRabbitMQ sets up the necessary exchanges and queues
func InitializeRabbitMQ(ch *amqp.Channel) {
	log.Println("⚙️ Initializing RabbitMQ...")

	// ✅ Declare Dead Letter Exchange (DLX)
	err := ch.ExchangeDeclare(
		constants.RTDNDLX, "direct", true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare Dead Letter Exchange: %v", err)
	}

	// ✅ Declare Dead Letter Queue (DLQ)
	_, err = ch.QueueDeclare(
		constants.RTDNDLQ, true, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare Dead Letter Queue: %v", err)
	}

	// ✅ Bind DLQ to DLX
	err = ch.QueueBind(constants.RTDNDLQ, constants.RTDNDLQ, constants.RTDNDLX, false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to bind Dead Letter Queue: %v", err)
	}

	// ✅ Declare the main RTDN queue
	_, err = ch.QueueDeclare(
		constants.RTDNQueue, true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    constants.RTDNDLX,
			"x-dead-letter-routing-key": constants.RTDNDLQ,
		},
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare RTDN Queue: %v", err)
	}

	log.Println("✅ RabbitMQ setup complete: RTDN Queue, DLX, and DLQ are ready")
}
