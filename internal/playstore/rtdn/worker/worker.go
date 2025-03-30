// internal/google_playstore/rtdn/worker.go
package rtdn

import (
	"context"
	"log"
	"math/rand"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	"subsnotifpro-go/internal/playstore/rtdn/service"

	"gorm.io/gorm"
)

// Worker struct to hold repository and service
type Worker struct {
	repo    repository.RTDNRepository
	service service.RTDNService
}

// NewWorker initializes the Worker with repository and service dependencies
func NewWorker(repo repository.RTDNRepository, svc service.RTDNService) *Worker {
	return &Worker{
		repo:    repo,
		service: svc,
	}
}

// ProcessPendingEvents processes pending webhook events with retry logic
func (w *Worker) ProcessPendingEvents(ctx context.Context, batchSize int) {
	ticker := time.NewTicker(time.Second * constants.EventProcessingInterval)
	defer ticker.Stop()

	log.Println("🚀 Started pending event processing...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🚦 [CTX] Shutdown signal received. Stopping event processing.")
			return

		case <-ticker.C:
			if ctx.Err() != nil {
				log.Println("🚦 [CTX] Context cancelled. Exiting event processor.")
				return
			}

			log.Println("📢 Checking for pending events...")

			// 🧩 Fetch pending events in batches
			events, err := w.repo.GetPendingEvents(ctx, nil, batchSize)
			if err != nil {
				log.Printf("❌ Error fetching pending events: %v", err)
				continue
			}

			for _, event := range events {
				// 🧩 Process each event in its own transaction
				err := w.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
					if event.RetryCount >= constants.MaxRetries {
						log.Printf("⚠️ Max retries reached for event %s. Moving to DLQ...", event.ID)
						return w.repo.MoveToDeadLetterQueue(ctx, tx, event.ID)
					}

					// Process using service (you can implement `ProcessWebhookEventWithTx`)
					err := w.service.ProcessWebhookEvent(event)
					if err != nil {
						log.Printf("❌ Error processing event %s: %v", event.ID, err)
						exponentialBackoffWithJitter(event.RetryCount)

						// Retry count increment
						if err := w.repo.IncrementRetryCount(ctx, tx, event.ID); err != nil {
							log.Printf("⚠️ Failed to increment retry count for event %s: %v", event.ID, err)
						}
						return err
					}

					// Mark completed
					if err := w.repo.UpdateWebhookStatus(ctx, tx, event.ID, "completed"); err != nil {
						log.Printf("⚠️ Failed to update status for event %s: %v", event.ID, err)
						return err
					}

					return nil
				})

				if err != nil {
					log.Printf("❌ Transaction failed for event %s: %v", event.ID, err)
				}
			}
		}
	}
}

// processEvents handles individual event processing with retry logic
func (w *Worker) processEvents(ctx context.Context, events []models.GooglePlayWebhookEvent) {
	for _, event := range events {
		select {
		case <-ctx.Done():
			log.Println("🚦 Stopping event processing mid-loop due to shutdown signal.")
			return

		default:
			err := w.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
				if event.RetryCount >= constants.MaxRetries {
					log.Printf("⚠️ Max retries reached for event %s. Moving to DLQ...", event.ID)
					return w.repo.MoveToDeadLetterQueue(ctx, tx, event.ID)
				}

				// 🔁 Process event via RTDN service
				if err := w.service.ProcessWebhookEvent(event); err != nil {
					log.Printf("❌ Error processing event %s: %v", event.ID, err)
					exponentialBackoffWithJitter(event.RetryCount)

					if err := w.repo.IncrementRetryCount(ctx, tx, event.ID); err != nil {
						log.Printf("⚠️ Failed to increment retry count for event %s: %v", event.ID, err)
					}
					return err
				}

				// ✅ Mark as completed
				if err := w.repo.UpdateWebhookStatus(ctx, tx, event.ID, "completed"); err != nil {
					log.Printf("⚠️ Failed to update status for event %s: %v", event.ID, err)
					return err
				}
				return nil
			})

			if err != nil {
				log.Printf("❌ Transaction failed for event %s: %v", event.ID, err)
			}
		}
	}
}

// retryMoveToDLQ attempts to move an event to the Dead Letter Queue (DLQ) with retries
func (w *Worker) retryMoveToDLQ(ctx context.Context, event models.GooglePlayWebhookEvent, retries int) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		err := w.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
			return w.repo.MoveToDeadLetterQueue(ctx, tx, event.ID)
		})
		if err == nil {
			return nil // ✅ Success
		}
		lastErr = err
		log.Printf("⚠️ Retry %d/%d: Failed to move event %s to DLQ. Retrying...", i+1, retries, event.ID)
		time.Sleep(2 * time.Second)
	}
	return lastErr
}

// exponentialBackoffWithJitter applies an exponential backoff before retrying
func exponentialBackoffWithJitter(retryCount int) {
	baseDelay := 1 << retryCount // 2^retryCount (1s, 2s, 4s, 8s...)
	jitter := rand.Intn(1000)    // Random jitter up to 1000ms
	delay := time.Duration(baseDelay*1000+jitter) * time.Millisecond

	log.Printf("🔄 Applying retry delay of %v (Retry #%d)", delay, retryCount+1)
	time.Sleep(delay) // Apply the delay before retrying
}
