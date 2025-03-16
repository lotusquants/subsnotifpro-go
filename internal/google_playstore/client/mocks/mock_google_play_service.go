package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/api/androidpublisher/v3"
)

// ✅ MockGooglePlayService mocks androidpublisher.Service
type MockGooglePlayService struct {
	mock.Mock
}

// ✅ Mock Get Subscription Purchase
func (m *MockGooglePlayService) Get(packageName, purchaseToken string) *androidpublisher.PurchasesSubscriptionsv2GetCall {
	args := m.Called(packageName, purchaseToken)
	call := new(androidpublisher.PurchasesSubscriptionsv2GetCall)
	if _, ok := args.Get(0).(*androidpublisher.SubscriptionPurchaseV2); ok {
		call = call.Context(context.Background())
		_ = call // Ignore error in mock
	}
	return call
}

// ✅ Mock List Subscriptions
func (m *MockGooglePlayService) List(packageName string) *androidpublisher.MonetizationSubscriptionsListCall {
	args := m.Called(packageName)
	call := new(androidpublisher.MonetizationSubscriptionsListCall)
	if _, ok := args.Get(0).(*androidpublisher.ListSubscriptionsResponse); ok {
		call = call.Context(context.Background())
		_ = call // Ignore error in mock
	}
	return call
}

// ✅ Mock Get Subscription Details
func (m *MockGooglePlayService) GetSubscriptionDetails(packageName, productID string) *androidpublisher.MonetizationSubscriptionsGetCall {
	args := m.Called(packageName, productID)
	call := new(androidpublisher.MonetizationSubscriptionsGetCall)
	if _, ok := args.Get(0).(*androidpublisher.Subscription); ok {
		call = call.Context(context.Background())
		_ = call // Ignore error in mock
	}
	return call
}
