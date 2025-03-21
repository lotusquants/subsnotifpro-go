package handler

import (
	"log"
	"net/http"
	"strconv"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/service"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionCatalogHandler handles subscription sync-related endpoints.
type SubscriptionCatalogHandler struct {
	service service.SubscriptionCatalogService
}

// NewSubscriptionCatalogHandler initializes the handler with the service dependency.
func NewSubscriptionCatalogHandler(svc service.SubscriptionCatalogService) *SubscriptionCatalogHandler {
	return &SubscriptionCatalogHandler{service: svc}
}

// RegisterPlaystoreSubscriptionCatalogRoutes registers all routes for managing subscription products, base plans, and offers.
func RegisterPlaystoreSubscriptionCatalogRoutes(r *gin.RouterGroup, handler *SubscriptionCatalogHandler) {
	r.POST("/sync-subscription-catalog", handler.SyncSubscriptionCatalogHandler)

	r.GET("/get-subscription-product-details", handler.GetSubscriptionProductDetailsHandler)
	r.GET("/check-subscription-product-exists", handler.CheckSubscriptionProductExistsHandler)
	r.GET("/list-all-subscription-products", handler.ListAllSubscriptionProductsHandler)

	r.GET("/get-subscription-baseplan-details", handler.GetBasePlanDetailsHandler)
	r.GET("/check-subscription-baseplan-active", handler.CheckBasePlanActiveHandler)
	r.GET("/check-subscription-baseplan-availability-in-region", handler.CheckBasePlanAvailabilityInRegionHandler)
	r.GET("/list-subscription-baseplan-names", handler.ListBasePlanNamesHandler)
	r.GET("/get-regional-baseplan-price", handler.GetRegionalBasePlanPriceHandler)
	r.GET("/get-other-regions-baseplan-price", handler.GetOtherRegionsBasePlanPriceHandler)

	r.GET("/get-subscription-offer-details", handler.GetSubscriptionOfferDetailsHandler)
	r.GET("/list-subscription-offer-names", handler.ListOfferNamesForBasePlanHandler)
	r.GET("/check-subscription-offer-active", handler.CheckSubscriptionOfferActiveHandler)
	r.GET("/check-subscription-offer-availability-in-region", handler.CheckSubscriptionOfferAvailabilityInRegionHandler)

	r.GET("/get-subscription-offer-phases", handler.GetOfferPhasesHandler)
	r.GET("/check-offer-phase-exists", handler.CheckSubscriptionOfferPhaseExistsHandler)
	r.GET("/get-current-offer-phase", handler.GetCurrentOfferPhaseHandler)
	r.GET("/get-regional-offer-phase-price", handler.GetRegionalOfferPhasePriceHandler)
}

// -------------------------
// ✅ HTTP Handler: Sync Subscription Catalog  (Runs in a Goroutine)
// -------------------------
func (h *SubscriptionCatalogHandler) SyncSubscriptionCatalogHandler(c *gin.Context) {
	// ✅ Bind JSON Body
	type SyncRequest struct {
		PackageName string `json:"package_name" binding:"required"`
	}

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_name is required"})
		return
	}
	packageName := req.PackageName

	// ✅ Run in a background Goroutine
	go func() {
		log.Printf("🚀 Running subscription sync in background for package: %s", req.PackageName)
		if err := h.service.SyncSubscriptionCatalog(req.PackageName); err != nil {
			log.Printf("❌ Subscription sync failed for package: %s - %v", req.PackageName, err)
		} else {
			log.Printf("✅ Subscription sync completed for package: %s", req.PackageName)
		}
	}()

	// ✅ Immediately respond with "Accepted"
	c.JSON(http.StatusAccepted, gin.H{"message": "Subscription sync started", "packageName": packageName})
}

// ✅ HTTP Handler for Listing Subscription Products
func (h *SubscriptionCatalogHandler) ListAllSubscriptionProductsHandler(c *gin.Context) {
	products, err := h.service.ListSubscriptionProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// ✅ API: Get Subscription Product
func (h *SubscriptionCatalogHandler) GetSubscriptionProductDetailsHandler(c *gin.Context) {
	// Parse query parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")

	// Validate input
	if packageName == "" || productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing either or both of the required parameters package_name and product_id"})
		return
	}

	// Call service layer
	product, err := h.service.GetSubscriptionProduct(packageName, productID)
	if err != nil {
		log.Printf("❌ Failed to get subscription product: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{"product": product})
}

// ✅ API: Check if Subscription Product Exists
func (h *SubscriptionCatalogHandler) CheckSubscriptionProductExistsHandler(c *gin.Context) {
	// Parse query parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")

	// Validate input
	if packageName == "" || productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing either or both of the required parameters package_name and product_id"})
		return
	}

	// Call service layer
	exists, err := h.service.CheckSubscriptionProductExists(packageName, productID)
	if err != nil {
		log.Printf("❌ Failed to check subscription product status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{"Exists": exists})
}

// ✅ Handler: Get Full Base Plan Details
func (h *SubscriptionCatalogHandler) GetBasePlanDetailsHandler(c *gin.Context) {

	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")

	// ✅ Check for missing parameters and return specific errors
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}

	basePlan, err := h.service.GetBasePlanDetails(packageName, productID, basePlanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"basePlan": basePlan})
}

// ✅ Handler: Check if Base Plan is Active
func (h *SubscriptionCatalogHandler) CheckBasePlanActiveHandler(c *gin.Context) {

	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")

	// ✅ Check for missing parameters and return specific errors
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}

	isActive, err := h.service.IsBasePlanActive(packageName, productID, basePlanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_active": isActive})
}

// ✅ API Handler to Check Base Plan Availability in a Region
func (h *SubscriptionCatalogHandler) CheckBasePlanAvailabilityInRegionHandler(c *gin.Context) {
	// 🔹 Extract Query Parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	regionCode := c.Query("region_code")

	// ✅ Validate required parameters
	var missingFields []string
	if packageName == "" {
		missingFields = append(missingFields, "package_name")
	}
	if productID == "" {
		missingFields = append(missingFields, "product_id")
	}
	if basePlanID == "" {
		missingFields = append(missingFields, "base_plan_id")
	}
	if regionCode == "" {
		missingFields = append(missingFields, "region_code")
	}

	if len(missingFields) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Missing required parameters",
			"missingFields": missingFields,
		})
		return
	}

	// 🔹 Check If Base Plan is Available
	available, err := h.service.IsBasePlanAvailableInRegion(packageName, productID, basePlanID, regionCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Respond with Availability Status
	c.JSON(http.StatusOK, gin.H{"available": available})
}

// ✅ API Handler for Listing Base Plan Names
func (h *SubscriptionCatalogHandler) ListBasePlanNamesHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")

	if packageName == "" || productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	basePlans, err := h.service.ListBasePlanNames(packageName, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"base_plans": basePlans})
}

// ✅ HTTP Handler for Getting Regional Base Plan Price
func (h *SubscriptionCatalogHandler) GetRegionalBasePlanPriceHandler(c *gin.Context) {

	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	regionCode := c.Query("region_code")

	// ✅ Validate required parameters
	var missingFields []string
	if packageName == "" {
		missingFields = append(missingFields, "package_name")
	}
	if productID == "" {
		missingFields = append(missingFields, "product_id")
	}
	if basePlanID == "" {
		missingFields = append(missingFields, "base_plan_id")
	}
	if regionCode == "" {
		missingFields = append(missingFields, "region_code")
	}

	if len(missingFields) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Missing required parameters",
			"missingFields": missingFields,
		})
		return
	}

	// ✅ Fetch Base Plan Price (With Fallback)
	price, err := h.service.GetRegionalBasePlanPrice(packageName, productID, basePlanID, regionCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"base_plan_price": price})
}

// ✅ Handler for Fetching Other Regions Base Plan Price
func (h *SubscriptionCatalogHandler) GetOtherRegionsBasePlanPriceHandler(c *gin.Context) {
	// 🔹 Extract Query Parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	currency := c.Query("currency") // Expecting "USD" or "EUR"

	// 🔹 Validate Required Parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: currency"})
		return
	}

	// 🔹 Fetch Price from Service
	price, err := h.service.GetOtherRegionsBasePlanPrice(packageName, productID, basePlanID, currency)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Return the Price
	c.JSON(http.StatusOK, gin.H{"price": price})
}

// ✅ HTTP Handler for Fetching Subscription Offer
func (h *SubscriptionCatalogHandler) GetSubscriptionOfferDetailsHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")

	// ✅ Validate required query parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_name is required"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_plan_id is required"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offer_id is required"})
		return
	}

	// ✅ Fetch the offer
	offer, err := h.service.GetSubscriptionOffer(packageName, productID, basePlanID, offerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"offer": offer})
}

// ✅ API Handler for Listing Offer Names for a Base Plan
func (h *SubscriptionCatalogHandler) ListOfferNamesForBasePlanHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")

	// ✅ Validate required query parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_name is required"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_plan_id is required"})
		return
	}

	// Get the list of offer names for the base plan
	offerNames, err := h.service.ListOfferNamesForBasePlan(packageName, productID, basePlanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the offer names in response
	c.JSON(http.StatusOK, gin.H{"offer_names": offerNames})
}

// ✅ API Handler for Checking if an Offer is Active
func (h *SubscriptionCatalogHandler) CheckSubscriptionOfferActiveHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")

	// ✅ Validate required query parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_name is required"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base_plan_id is required"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offer_id is required"})
		return
	}

	isActive, err := h.service.IsSubscriptionOfferActive(packageName, productID, basePlanID, offerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_active": isActive})
}

// ✅ Check Offer Availability in a Region - Handler
func (h *SubscriptionCatalogHandler) CheckSubscriptionOfferAvailabilityInRegionHandler(c *gin.Context) {
	// 🔹 Extract Query Parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")
	regionCode := c.Query("region_code")

	// 🔹 Validate Required Parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: offer_id"})
		return
	}
	if regionCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: region_code"})
		return
	}

	// 🔹 Check Offer Availability in Region
	available, err := h.service.CheckSubscriptionOfferAvailabilityInRegion(packageName, productID, basePlanID, offerID, regionCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Return Response
	c.JSON(http.StatusOK, gin.H{"offer_available": available})
}

// ✅ API: Get Offer Phases
func (h *SubscriptionCatalogHandler) GetOfferPhasesHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")

	// ✅ Validate Required Parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: offer_id"})
		return
	}

	phases, err := h.service.GetOfferPhases(packageName, productID, basePlanID, offerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"phases": phases})
}

// ✅ HTTP Handler: Check if a Subscription Offer Phase Exists
func (h *SubscriptionCatalogHandler) CheckSubscriptionOfferPhaseExistsHandler(c *gin.Context) {
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")

	// ✅ Validate Required Query Parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: offer_id"})
		return
	}

	// ✅ Check if Offer Phase Exists
	exists, err := h.service.IsOfferPhaseExists(packageName, productID, basePlanID, offerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ✅ Return Response
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

// ✅ GetCurrentOfferPhaseHandler - Handles request to fetch the current offer phase.
func (h *SubscriptionCatalogHandler) GetCurrentOfferPhaseHandler(c *gin.Context) {
	// 🔹 Extract Query Parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")
	startTimeStr := c.Query("start_time") // Expecting ISO 8601 format: "2024-08-18T16:38:33.598Z"
	checkTimeStr := c.Query("check_time") // Expecting ISO 8601 format: "2024-08-18T16:38:33.598Z"

	// 🔹 Validate Required Parameters
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: offer_id"})
		return
	}
	if startTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: start_time"})
		return
	}

	if checkTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: check_time"})
		return
	}

	// ✅ Handle Start Time Parsing with Milliseconds
	var startTime time.Time
	var checkTime time.Time
	var err error

	// Try Parsing with Milliseconds
	startTime, err = time.Parse(time.RFC3339Nano, startTimeStr) // Supports milliseconds (RFC3339Nano)
	if err != nil {
		// Try Without Milliseconds
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format. Expected RFC3339 or RFC3339Nano"})
			return
		}
	}

	// Try Parsing with Milliseconds
	checkTime, err = time.Parse(time.RFC3339Nano, checkTimeStr) // Supports milliseconds (RFC3339Nano)
	if err != nil {
		// Try Without Milliseconds
		checkTime, err = time.Parse(time.RFC3339, checkTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check_time format. Expected RFC3339 or RFC3339Nano"})
			return
		}
	}

	// 🔹 Fetch Current Offer Phase from Service
	currentPhaseIndex, err := h.service.GetCurrentOfferPhaseIndex(packageName, productID, basePlanID, offerID, startTime, checkTime)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Return the Current Offer Phase
	c.JSON(http.StatusOK, gin.H{"current_phase_index": currentPhaseIndex})
}

// ✅ GetRegionalOfferPhasePriceHandler - Handles request to fetch offer phase price for a specific region
func (h *SubscriptionCatalogHandler) GetRegionalOfferPhasePriceHandler(c *gin.Context) {
	// 🔹 Extract Query Parameters
	packageName := c.Query("package_name")
	productID := c.Query("product_id")
	basePlanID := c.Query("base_plan_id")
	offerID := c.Query("offer_id")
	phaseIndexStr := c.Query("phase_index") // Expecting an integer
	regionCode := c.Query("region_code")

	// 🔹 Validate Required Parameters Individually with Detailed Errors
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: package_name"})
		return
	}
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: product_id"})
		return
	}
	if basePlanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: base_plan_id"})
		return
	}
	if offerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: offer_id"})
		return
	}
	if phaseIndexStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: phase_index"})
		return
	}
	if regionCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameter: region_code"})
		return
	}

	// 🔹 Parse Phase Index as Integer
	phaseIndex, err := strconv.Atoi(phaseIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter: phase_index must be an integer"})
		return
	}

	// 🔹 Fetch Regional Offer Phase Price from Service
	price, err := h.service.GetRegionalOfferPhasePrice(packageName, productID, basePlanID, offerID, phaseIndex, regionCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 🔹 Return the Offer Phase Price
	c.JSON(http.StatusOK, gin.H{"price": price})
}
