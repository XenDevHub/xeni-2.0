package public

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/xeni-ai/gateway/internal/models"
	"github.com/xeni-ai/gateway/pkg/response"
)

// InitSSLCommerzRequest payload for initiating payment.
type InitSSLCommerzRequest struct {
	OrderID     string `json:"order_id"`
	FrontendURL string `json:"frontend_url"`
}

// SSLCommerzInitResponse parsed from SSLCommerz API response.
type SSLCommerzInitResponse struct {
	Status         string `json:"status"`
	FailedReason   string `json:"failedreason"`
	GatewayPageURL string `json:"GatewayPageURL"`
	Sessionkey     string `json:"sessionkey"`
}

// ── POST /api/public/payments/sslcommerz/init ──
// Initializes an SSLCommerz hosted payment session dynamically using either
// the specific shop's SSLCommerz credentials or the marketplace platform fallback.
func (h *Handler) InitSSLCommerzPayment(c *fiber.Ctx) error {
	var req InitSSLCommerzRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.OrderID == "" {
		return response.BadRequest(c, "order_id is required")
	}

	orderUUID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return response.BadRequest(c, "Invalid order_id UUID")
	}

	var order models.Order
	if err := h.DB.Where("id = ?", orderUUID).First(&order).Error; err != nil {
		return response.NotFound(c, "Order not found")
	}

	var shop models.Shop
	h.DB.Where("id = ?", order.ShopID).First(&shop)

	// Determine SSLCommerz credentials (Shop-specific or Platform Fallback)
	storeID := os.Getenv("SSLCZ_STORE_ID")
	storePass := os.Getenv("SSLCZ_STORE_PASS")
	isLive := os.Getenv("SSLCZ_IS_LIVE") == "true"

	// If shop has their own SSLCommerz credentials, prioritize them!
	if shop.SSLCommerzStoreID != nil && *shop.SSLCommerzStoreID != "" &&
		shop.SSLCommerzStorePass != nil && *shop.SSLCommerzStorePass != "" {
		storeID = *shop.SSLCommerzStoreID
		storePass = *shop.SSLCommerzStorePass
		isLive = shop.SSLCommerzIsLive
	}

	// Fallback to standard SSLCommerz test sandbox if unconfigured
	if storeID == "" {
		storeID = "testbox"
		storePass = "qwerty"
		isLive = false
	}

	apiURL := "https://sandbox.sslcommerz.com/gwprocess/v4/api.php"
	if isLive {
		apiURL = "https://securepay.sslcommerz.com/gwprocess/v4/api.php"
	}

	// Gateway Base for callbacks
	gatewayBase := os.Getenv("PUBLIC_GATEWAY_URL")
	if gatewayBase == "" {
		gatewayBase = "https://xeni.e-pic.co"
	}
	gatewayBase = strings.TrimRight(gatewayBase, "/")

	custName := "E-Pic Customer"
	if order.CustomerName != nil && *order.CustomerName != "" {
		custName = *order.CustomerName
	}
	custPhone := "01700000000"
	if order.CustomerPhone != nil && *order.CustomerPhone != "" {
		custPhone = *order.CustomerPhone
	}
	custAddress := "Dhaka, Bangladesh"
	if order.CustomerAddress != nil && *order.CustomerAddress != "" {
		custAddress = *order.CustomerAddress
	}

	formData := url.Values{}
	formData.Set("store_id", storeID)
	formData.Set("store_passwd", storePass)
	formData.Set("total_amount", fmt.Sprintf("%.2f", order.TotalAmount))
	formData.Set("currency", "BDT")
	formData.Set("tran_id", order.ID.String())
	formData.Set("success_url", gatewayBase+"/api/public/payments/sslcommerz/success")
	formData.Set("fail_url", gatewayBase+"/api/public/payments/sslcommerz/fail")
	formData.Set("cancel_url", gatewayBase+"/api/public/payments/sslcommerz/cancel")
	formData.Set("ipn_url", gatewayBase+"/api/public/payments/sslcommerz/ipn")
	formData.Set("cus_name", custName)
	formData.Set("cus_email", "customer@e-pic.co")
	formData.Set("cus_add1", custAddress)
	formData.Set("cus_city", "Dhaka")
	formData.Set("cus_country", "Bangladesh")
	formData.Set("cus_phone", custPhone)
	formData.Set("shipping_method", "NO")
	formData.Set("product_name", "E-Pic Marketplace Purchase")
	formData.Set("product_category", "Goods")
	formData.Set("product_profile", "general")

	client := &http.Client{Timeout: 15 * time.Second}
	httpResp, err := client.PostForm(apiURL, formData)
	if err != nil {
		return response.InternalError(c)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return response.InternalError(c)
	}

	var sslczResp SSLCommerzInitResponse
	if err := json.Unmarshal(bodyBytes, &sslczResp); err != nil {
		return response.BadRequest(c, "Failed to parse SSLCommerz response: "+string(bodyBytes))
	}

	if sslczResp.Status != "SUCCESS" {
		return response.BadRequest(c, "SSLCommerz session initialization failed: "+sslczResp.FailedReason)
	}

	return response.Success(c, fiber.Map{
		"gateway_url": sslczResp.GatewayPageURL,
		"sessionkey":  sslczResp.Sessionkey,
		"order_id":    order.ID.String(),
	})
}

// ── POST /api/public/payments/sslcommerz/success ──
func (h *Handler) SSLCommerzSuccess(c *fiber.Ctx) error {
	tranID := c.FormValue("tran_id")
	valID := c.FormValue("val_id")

	if tranID != "" {
		orderUUID, err := uuid.Parse(tranID)
		if err == nil {
			var order models.Order
			if err := h.DB.Where("id = ?", orderUUID).First(&order).Error; err == nil {
				order.PaymentStatus = models.OrderPayVerified
				pm := models.PaymentSSLCommerz
				order.PaymentMethod = &pm
				order.PaymentTrxID = &valID
				now := time.Now()
				order.VerifiedAt = &now
				vBy := "sslcommerz"
				order.VerifiedBy = &vBy
				h.DB.Save(&order)
			}
		}
	}

	frontendURL := os.Getenv("FRONTEND_STORE_URL")
	if frontendURL == "" {
		frontendURL = "https://epic.e-pic.co"
	}
	frontendURL = strings.TrimRight(frontendURL, "/")

	return c.Redirect(fmt.Sprintf("%s/order/confirmation?id=%s&status=success&val_id=%s", frontendURL, tranID, valID), fiber.StatusSeeOther)
}

// ── POST /api/public/payments/sslcommerz/fail ──
func (h *Handler) SSLCommerzFail(c *fiber.Ctx) error {
	tranID := c.FormValue("tran_id")

	if tranID != "" {
		orderUUID, err := uuid.Parse(tranID)
		if err == nil {
			var order models.Order
			if err := h.DB.Where("id = ?", orderUUID).First(&order).Error; err == nil {
				order.PaymentStatus = models.OrderPayFailed
				h.DB.Save(&order)
			}
		}
	}

	frontendURL := os.Getenv("FRONTEND_STORE_URL")
	if frontendURL == "" {
		frontendURL = "https://epic.e-pic.co"
	}
	frontendURL = strings.TrimRight(frontendURL, "/")

	return c.Redirect(fmt.Sprintf("%s/order/confirmation?id=%s&status=failed", frontendURL, tranID), fiber.StatusSeeOther)
}

// ── POST /api/public/payments/sslcommerz/cancel ──
func (h *Handler) SSLCommerzCancel(c *fiber.Ctx) error {
	frontendURL := os.Getenv("FRONTEND_STORE_URL")
	if frontendURL == "" {
		frontendURL = "https://epic.e-pic.co"
	}
	frontendURL = strings.TrimRight(frontendURL, "/")

	return c.Redirect(fmt.Sprintf("%s/checkout?status=cancelled", frontendURL), fiber.StatusSeeOther)
}

// ── POST /api/public/payments/sslcommerz/ipn ──
func (h *Handler) SSLCommerzIPN(c *fiber.Ctx) error {
	tranID := c.FormValue("tran_id")
	valID := c.FormValue("val_id")
	status := c.FormValue("status")

	if tranID != "" && status == "VALID" {
		orderUUID, err := uuid.Parse(tranID)
		if err == nil {
			var order models.Order
			if err := h.DB.Where("id = ?", orderUUID).First(&order).Error; err == nil {
				order.PaymentStatus = models.OrderPayVerified
				order.PaymentTrxID = &valID
				now := time.Now()
				order.VerifiedAt = &now
				vBy := "sslcommerz_ipn"
				order.VerifiedBy = &vBy
				h.DB.Save(&order)
			}
		}
	}

	return c.SendStatus(fiber.StatusOK)
}
