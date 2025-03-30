package models

// VoidedPurchaseProductType represents what type of product was voided
type VoidedPurchaseProductType int

const (
	ProductTypeSubscription VoidedPurchaseProductType = 1
	ProductTypeOneTime      VoidedPurchaseProductType = 2
)

func (t VoidedPurchaseProductType) String() string {
	switch t {
	case ProductTypeSubscription:
		return "PRODUCT_TYPE_SUBSCRIPTION"
	case ProductTypeOneTime:
		return "PRODUCT_TYPE_ONE_TIME"
	default:
		return "UNKNOWN"
	}
}

// VoidedPurchaseRefundType represents full or partial refund
type VoidedPurchaseRefundType int

const (
	RefundTypeFull    VoidedPurchaseRefundType = 1
	RefundTypePartial VoidedPurchaseRefundType = 2
)

func (t VoidedPurchaseRefundType) String() string {
	switch t {
	case RefundTypeFull:
		return "Full refund"
	case RefundTypePartial:
		return "Partial refund"
	default:
		return "Unknown refund type"
	}
}

// VoidedPurchaseNotification contains voided purchase details
type VoidedPurchaseNotification struct {
	PurchaseToken string                    `json:"purchaseToken,omitempty" gorm:"default:null"`
	OrderID       string                    `json:"orderId,omitempty" gorm:"default:null"`
	ProductType   VoidedPurchaseProductType `json:"productType,omitempty"`
	RefundType    VoidedPurchaseRefundType  `json:"refundType,omitempty"`
}
