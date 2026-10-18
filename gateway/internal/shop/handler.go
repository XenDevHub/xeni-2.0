package shop

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/xeni-ai/gateway/internal/models"
	"github.com/xeni-ai/gateway/pkg/response"
)

// Handler holds shop dependencies.
type Handler struct {
	DB *gorm.DB
}

// NewHandler creates a new shop handler.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// CreateShop handles POST /api/shops — create a shop for the current user.
func (h *Handler) CreateShop(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, _ := uuid.Parse(userID)

	// Check if user already has a shop
	var existing models.Shop
	if err := h.DB.Where("user_id = ?", uid).First(&existing).Error; err == nil {
		return response.BadRequest(c, "You already have a shop. Use PUT to update.")
	}

	var req struct {
		ShopName            string  `json:"shop_name"`
		ShopDescription     *string `json:"shop_description"`
		ShopLogoURL         *string `json:"shop_logo_url"`
		ShopBannerURL       *string `json:"shop_banner_url"`
		PreferredLanguage   string  `json:"preferred_language"`
		CourierPreference   string  `json:"courier_preference"`
		BkashMerchantNumber *string `json:"bkash_merchant_number"`
		NagadMerchantNumber *string `json:"nagad_merchant_number"`
		OwnerMobile         *string `json:"owner_mobile"`
		District            *string `json:"district"`
		DeliveryChargeInside  *float64 `json:"delivery_charge_inside"`
		DeliveryChargeOutside *float64 `json:"delivery_charge_outside"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.ShopName == "" {
		return response.BadRequest(c, "shop_name is required")
	}
	if req.PreferredLanguage == "" {
		req.PreferredLanguage = "bn"
	}
	if req.CourierPreference == "" {
		req.CourierPreference = "pathao"
	}

	shop := models.Shop{
		UserID:              uid,
		ShopName:            req.ShopName,
		ShopDescription:     req.ShopDescription,
		ShopLogoURL:         req.ShopLogoURL,
		ShopBannerURL:       req.ShopBannerURL,
		PreferredLanguage:   req.PreferredLanguage,
		CourierPreference:   req.CourierPreference,
		BkashMerchantNumber: req.BkashMerchantNumber,
		NagadMerchantNumber: req.NagadMerchantNumber,
		OwnerMobile:         req.OwnerMobile,
		District:            req.District,
		PaymentVerificationMode: "manual",
	}
	if req.DeliveryChargeInside != nil {
		shop.DeliveryChargeInside = *req.DeliveryChargeInside
	}
	if req.DeliveryChargeOutside != nil {
		shop.DeliveryChargeOutside = *req.DeliveryChargeOutside
	}

	// Auto-set WhatsApp number from owner mobile if not already set
	if req.OwnerMobile != nil && *req.OwnerMobile != "" {
		shop.WhatsAppNumber = req.OwnerMobile
	}

	if err := h.DB.Create(&shop).Error; err != nil {
		return response.InternalError(c)
	}

	return response.Created(c, shop)
}

// GetMyShop handles GET /api/shops/me — get current user's shop.
func (h *Handler) GetMyShop(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, _ := uuid.Parse(userID)

	var shop models.Shop
	if err := h.DB.Preload("ConnectedPages").Where("user_id = ?", uid).First(&shop).Error; err != nil {
		return response.NotFound(c, "No shop found. Create one first.")
	}

	return response.Success(c, shop)
}

// UpdateMyShop handles PUT /api/shops/me — update current user's shop.
func (h *Handler) UpdateMyShop(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, _ := uuid.Parse(userID)

	var shop models.Shop
	if err := h.DB.Where("user_id = ?", uid).First(&shop).Error; err != nil {
		return response.NotFound(c, "No shop found")
	}

	var req struct {
		ShopName            *string `json:"shop_name"`
		ShopDescription     *string `json:"shop_description"`
		ShopLogoURL         *string `json:"shop_logo_url"`
		PreferredLanguage   *string `json:"preferred_language"`
		CourierPreference   *string `json:"courier_preference"`
		BkashMerchantNumber *string                 `json:"bkash_merchant_number"`
		NagadMerchantNumber *string                 `json:"nagad_merchant_number"`
		OwnerMobile         *string                 `json:"owner_mobile"`
		District            *string                 `json:"district"`
		DeliveryChargeInside  *float64               `json:"delivery_charge_inside"`
		DeliveryChargeOutside *float64               `json:"delivery_charge_outside"`
		PaymentVerificationMode *string              `json:"payment_verification_mode"`
		BkashAppKey         *string                 `json:"bkash_app_key"`
		BkashAppSecret      *string                 `json:"bkash_app_secret"`
		NagadMerchantID     *string                 `json:"nagad_merchant_id"`
		NagadMerchantKey    *string                 `json:"nagad_merchant_key"`
		SSLCommerzStoreID   *string                 `json:"sslcommerz_store_id"`
		SSLCommerzStorePass *string                 `json:"sslcommerz_store_passwd"`
		SSLCommerzIsLive    *bool                   `json:"sslcommerz_is_live"`
		AutoReplyEnabled    *bool                   `json:"auto_reply_enabled"`
		AutoOrderEnabled    *bool                   `json:"auto_order_enabled"`
		Integrations        *map[string]interface{} `json:"integrations"`
		CustomAgentRules    *string                 `json:"custom_agent_rules"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	updates := make(map[string]interface{})
	if req.ShopName != nil {
		updates["shop_name"] = *req.ShopName
	}
	if req.ShopDescription != nil {
		updates["shop_description"] = *req.ShopDescription
	}
	if req.ShopLogoURL != nil {
		updates["shop_logo_url"] = *req.ShopLogoURL
	}
	if req.ShopBannerURL != nil {
		updates["shop_banner_url"] = *req.ShopBannerURL
	}
	if req.PreferredLanguage != nil {
		updates["preferred_language"] = *req.PreferredLanguage
	}
	if req.CourierPreference != nil {
		updates["courier_preference"] = *req.CourierPreference
	}
	if req.BkashMerchantNumber != nil {
		updates["bkash_merchant_number"] = *req.BkashMerchantNumber
	}
	if req.NagadMerchantNumber != nil {
		updates["nagad_merchant_number"] = *req.NagadMerchantNumber
	}
	if req.OwnerMobile != nil {
		updates["owner_mobile"] = *req.OwnerMobile
		// Sync WhatsApp number
		if *req.OwnerMobile != "" {
			updates["whatsapp_number"] = *req.OwnerMobile
		}
	}
	if req.District != nil {
		updates["district"] = *req.District
	}
	if req.SSLCommerzStoreID != nil {
		updates["sslcommerz_store_id"] = *req.SSLCommerzStoreID
	}
	if req.SSLCommerzStorePass != nil && *req.SSLCommerzStorePass != "" {
		updates["sslcommerz_store_passwd"] = *req.SSLCommerzStorePass
	}
	if req.SSLCommerzIsLive != nil {
		updates["sslcommerz_is_live"] = *req.SSLCommerzIsLive
	}
	if req.DeliveryChargeInside != nil {
		updates["delivery_charge_inside"] = *req.DeliveryChargeInside
	}
	if req.DeliveryChargeOutside != nil {
		updates["delivery_charge_outside"] = *req.DeliveryChargeOutside
	}
	if req.PaymentVerificationMode != nil {
		updates["payment_verification_mode"] = *req.PaymentVerificationMode
	}
	if req.BkashAppKey != nil {
		updates["bkash_app_key"] = *req.BkashAppKey
	}
	if req.BkashAppSecret != nil {
		updates["bkash_app_secret"] = *req.BkashAppSecret
	}
	if req.NagadMerchantID != nil {
		updates["nagad_merchant_id"] = *req.NagadMerchantID
	}
	if req.NagadMerchantKey != nil {
		updates["nagad_merchant_key"] = *req.NagadMerchantKey
	}
	if req.AutoReplyEnabled != nil {
		updates["auto_reply_enabled"] = *req.AutoReplyEnabled
	}
	if req.AutoOrderEnabled != nil {
		updates["auto_order_enabled"] = *req.AutoOrderEnabled
	}
	if req.Integrations != nil {
		updates["integrations"] = *req.Integrations
	}
	if req.CustomAgentRules != nil {
		updates["custom_agent_rules"] = *req.CustomAgentRules
	}

	if len(updates) > 0 {
		if err := h.DB.Model(&shop).Updates(updates).Error; err != nil {
			return response.InternalError(c)
		}
	}

	h.DB.Preload("ConnectedPages").First(&shop, shop.ID)
	return response.Success(c, shop)
}

// GetIntegrations handles GET /api/shops/integrations — get masked integration statuses.
func (h *Handler) GetIntegrations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, _ := uuid.Parse(userID)

	var shop models.Shop
	if err := h.DB.Select("bkash_app_key, bkash_app_secret, bkash_username, bkash_password, nagad_merchant_id, nagad_merchant_key, pathao_client_id, pathao_client_secret, pathao_username, pathao_password, steadfast_api_key, steadfast_secret_key").Where("user_id = ?", uid).First(&shop).Error; err != nil {
		return response.NotFound(c, "No shop found")
	}

	return response.Success(c, map[string]interface{}{
		"bkash": map[string]interface{}{
			"is_configured": shop.BkashAppKey != nil && *shop.BkashAppKey != "",
		},
		"nagad": map[string]interface{}{
			"is_configured": shop.NagadMerchantID != nil && *shop.NagadMerchantID != "",
		},
		"pathao": map[string]interface{}{
			"is_configured": shop.PathaoClientID != nil && *shop.PathaoClientID != "",
		},
		"steadfast": map[string]interface{}{
			"is_configured": shop.SteadfastAPIKey != nil && *shop.SteadfastAPIKey != "",
		},
	})
}

// UpdateIntegrations handles PUT /api/shops/integrations — update merchant integrations.
func (h *Handler) UpdateIntegrations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	uid, _ := uuid.Parse(userID)

	var shop models.Shop
	if err := h.DB.Where("user_id = ?", uid).First(&shop).Error; err != nil {
		return response.NotFound(c, "No shop found")
	}

	var req struct {
		BkashAppKey        *string `json:"bkash_app_key"`
		BkashAppSecret     *string `json:"bkash_app_secret"`
		BkashUsername      *string `json:"bkash_username"`
		BkashPassword      *string `json:"bkash_password"`
		NagadMerchantID    *string `json:"nagad_merchant_id"`
		NagadMerchantKey   *string `json:"nagad_merchant_key"`
		PathaoClientID     *string `json:"pathao_client_id"`
		PathaoClientSecret *string `json:"pathao_client_secret"`
		PathaoUsername     *string `json:"pathao_username"`
		PathaoPassword     *string `json:"pathao_password"`
		SteadfastAPIKey    *string `json:"steadfast_api_key"`
		SteadfastSecretKey *string `json:"steadfast_secret_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	updates := make(map[string]interface{})

	// Only update fields that are explicitly provided (not nil) and not empty masks like "********"
	if req.BkashAppKey != nil && *req.BkashAppKey != "********" { updates["bkash_app_key"] = *req.BkashAppKey }
	if req.BkashAppSecret != nil && *req.BkashAppSecret != "********" { updates["bkash_app_secret"] = *req.BkashAppSecret }
	if req.BkashUsername != nil && *req.BkashUsername != "********" { updates["bkash_username"] = *req.BkashUsername }
	if req.BkashPassword != nil && *req.BkashPassword != "********" { updates["bkash_password"] = *req.BkashPassword }

	if req.NagadMerchantID != nil && *req.NagadMerchantID != "********" { updates["nagad_merchant_id"] = *req.NagadMerchantID }
	if req.NagadMerchantKey != nil && *req.NagadMerchantKey != "********" { updates["nagad_merchant_key"] = *req.NagadMerchantKey }

	if req.PathaoClientID != nil && *req.PathaoClientID != "********" { updates["pathao_client_id"] = *req.PathaoClientID }
	if req.PathaoClientSecret != nil && *req.PathaoClientSecret != "********" { updates["pathao_client_secret"] = *req.PathaoClientSecret }
	if req.PathaoUsername != nil && *req.PathaoUsername != "********" { updates["pathao_username"] = *req.PathaoUsername }
	if req.PathaoPassword != nil && *req.PathaoPassword != "********" { updates["pathao_password"] = *req.PathaoPassword }

	if req.SteadfastAPIKey != nil && *req.SteadfastAPIKey != "********" { updates["steadfast_api_key"] = *req.SteadfastAPIKey }
	if req.SteadfastSecretKey != nil && *req.SteadfastSecretKey != "********" { updates["steadfast_secret_key"] = *req.SteadfastSecretKey }

	if len(updates) > 0 {
		h.DB.Model(&shop).Updates(updates)
	}

	return response.Success(c, map[string]string{"message": "Integrations updated successfully"})
}
