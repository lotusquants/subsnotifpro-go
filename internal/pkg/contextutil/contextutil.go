// internal/pkg/contextutil/contextutil.go
package contextutil

import (
	"context"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type contextKey string

const (
	// TxKey is the key used to store *gorm.DB transaction in context
	TxKey          contextKey = "gormTx"
	DeliveryCtxKey contextKey = "amqp-delivery"
	RetryCountKey  contextKey = "retry-count"
	requestIDKey   contextKey = "request_id"
)

func ContextWithDelivery(ctx context.Context, msg *amqp.Delivery) context.Context {
	return context.WithValue(ctx, DeliveryCtxKey, msg)
}

func DeliveryFromContext(ctx context.Context) (*amqp.Delivery, bool) {
	msg, ok := ctx.Value(DeliveryCtxKey).(*amqp.Delivery)
	return msg, ok
}

// WithTx adds a GORM transaction to context
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

// TxFromContext retrieves a GORM transaction from context
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	return tx, ok
}

// MustTx panics if no transaction is found in context
func MustTx(ctx context.Context) *gorm.DB {
	tx, ok := TxFromContext(ctx)
	if !ok {
		panic("no transaction found in context")
	}
	return tx
}
