// internal/google_playstore/rtdn/worker.go
package rtdn

import (
	"context"
	"log"
	"math/rand"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/rtdn/models"
	"subsnotifpro-go/internal/google_playstore/rtdn/repository"
	"subsnotifpro-go/internal/google_playstore/rtdn/service"
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
	defer ticker.Stop() // ✅ Stop ticker when function exits

	log.Println("🚀 Started pending event processing...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🚦 [CTX] Shutdown signal received. Stopping event processing.")
			return // ✅ Ensure function exits immediately

		case <-ticker.C:
			if ctx.Err() != nil {
				log.Println("🚦 [CTX] Context cancelled. Exiting event processor.")
				return
			}

			log.Println("📢 Checking for pending events...")

			// ✅ Fetch pending events in batches
			events, err := w.repo.GetPendingEvents(ctx, batchSize)
			if err != nil {
				log.Printf("❌ Error fetching pending events: %v", err)
				continue
			}

			// ✅ Process events safely
			w.processEvents(ctx, events)
		}
	}
}

// processEvents handles individual event processing with retry logic
func (w *Worker) processEvents(ctx context.Context, events []models.GooglePlayWebhookEvent) {
	for _, event := range events {
		select {
		case <-ctx.Done():
			log.Println("🚦 Stopping event processing mid-loop due to shutdown signal.")
			return // ✅ Ensure graceful shutdown

		default:
			// ✅ Move to DLQ if max retries reached
			if event.RetryCount >= constants.MaxRetries {
				log.Printf("⚠️ Max retries reached for event %s. Moving to DLQ...", event.ID)
				if err := w.retryMoveToDLQ(ctx, event, 3); err != nil { // ✅ Retry DLQ move before failing
					log.Printf("❌ Critical: Failed to move event %s to DLQ after multiple retries: %v", event.ID, err)
				}
				continue
			}

			// ✅ Process the event
			err := w.service.ProcessWebhookEvent(event)
			if err != nil {
				log.Printf("❌ Error processing event %s. Retrying...", event.ID)
				exponentialBackoffWithJitter(event.RetryCount)

				if err := w.repo.IncrementRetryCount(ctx, event.ID); err != nil {
					log.Printf("⚠️ Failed to increment retry count for event %s: %v", event.ID, err)
				}
			} else {
				if err := w.repo.UpdateWebhookStatus(ctx, event.ID, "completed"); err != nil {
					log.Printf("⚠️ Failed to update status for event %s: %v", event.ID, err)
				}
			}
		}
	}
}

// retryMoveToDLQ attempts to move an event to the Dead Letter Queue (DLQ) with retries
func (w *Worker) retryMoveToDLQ(ctx context.Context, event models.GooglePlayWebhookEvent, retries int) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		err := w.repo.MoveToDeadLetterQueue(ctx, event.ID)
		if err == nil {
			return nil // ✅ Success
		}
		lastErr = err
		log.Printf("⚠️ Retry %d/%d: Failed to move event %s to DLQ. Retrying...", i+1, retries, event.ID)
		time.Sleep(2 * time.Second) // ✅ Add small delay between retries
	}
	return lastErr // ❌ Return last failure
}

// exponentialBackoffWithJitter applies an exponential backoff before retrying
func exponentialBackoffWithJitter(retryCount int) {
	baseDelay := 1 << retryCount // 2^retryCount (1s, 2s, 4s, 8s...)
	jitter := rand.Intn(1000)    // Random jitter up to 1000ms
	delay := time.Duration(baseDelay*1000+jitter) * time.Millisecond

	log.Printf("🔄 Applying retry delay of %v (Retry #%d)", delay, retryCount+1)
	time.Sleep(delay) // Apply the delay before retrying
}
