package public

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/xeni-ai/gateway/internal/models"
	"github.com/xeni-ai/gateway/pkg/response"
)

// Handler serves public marketplace endpoints — no auth required.
type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// ── Public Product DTOs ──

type PublicProductResponse struct {
	ID                string          `json:"id"`
	ShopID            string          `json:"shop_id"`
	ShopName          string          `json:"shop_name"`
	ShopLogoURL       *string         `json:"shop_logo_url"`
	Name              string          `json:"name"`
	NameBN            *string         `json:"name_bn"`
	Description       *string         `json:"description"`
	DescriptionBN     *string         `json:"description_bn"`
	Price             float64         `json:"price"`
	SKU               *string         `json:"sku"`
	InitialStock      int             `json:"initial_stock"`
	CurrentStock      int             `json:"current_stock"`
	LowStockThreshold int             `json:"low_stock_threshold"`
	IsActive          bool            `json:"is_active"`
	IsOutOfStock      bool            `json:"is_out_of_stock"`
	HasVariants       bool            `json:"has_variants"`
	TotalSold         int             `json:"total_sold"`
	Images            json.RawMessage `json:"images"`
	Variants          []PublicVariant `json:"variants,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type PublicVariant struct {
	ID            string   `json:"id"`
	SKU           string   `json:"sku"`
	Color         *string  `json:"color"`
	Size          *string  `json:"size"`
	Stock         int      `json:"stock"`
	PriceModifier float64  `json:"price_modifier"`
}

type PublicShopResponse struct {
	ID                    string                  `json:"id"`
	ShopName              string                  `json:"shop_name"`
	ShopDescription       *string                 `json:"shop_description"`
	ShopLogoURL           *string                 `json:"shop_logo_url"`
	District              *string                 `json:"district"`
	BkashMerchantNumber   *string                 `json:"bkash_merchant_number"`
	NagadMerchantNumber   *string                 `json:"nagad_merchant_number"`
	DeliveryChargeInside  float64                 `json:"delivery_charge_inside"`
	DeliveryChargeOutside float64                 `json:"delivery_charge_outside"`
	CourierPreference     string                  `json:"courier_preference"`
	TotalProducts         int64                   `json:"total_products"`
	Products              []PublicProductResponse `json:"products,omitempty"`
	CreatedAt             time.Time               `json:"created_at"`
}

// buildProductResponse converts a models.Product + shop info to a public DTO.
func buildProductResponse(p models.Product, shop *models.Shop) PublicProductResponse {
	var shopName string
	var shopLogoURL *string
	if shop != nil {
		shopName = shop.ShopName
		shopLogoURL = shop.ShopLogoURL
	}

	variants := make([]PublicVariant, 0, len(p.Variants))
	for _, v := range p.Variants {
		variants = append(variants, PublicVariant{
			ID:            v.ID.String(),
			SKU:           v.SKU,
			Color:         v.Color,
			Size:          v.Size,
			Stock:         v.Stock,
			PriceModifier: v.PriceModifier,
		})
	}

	images := json.RawMessage(p.Images)
	if len(images) == 0 {
		images = json.RawMessage("[]")
	}

	return PublicProductResponse{
		ID:                p.ID.String(),
		ShopID:            p.ShopID.String(),
		ShopName:          shopName,
		ShopLogoURL:       shopLogoURL,
		Name:              p.Name,
		NameBN:            p.NameBN,
		Description:       p.Description,
		DescriptionBN:     p.DescriptionBN,
		Price:             p.Price,
		SKU:               p.SKU,
		InitialStock:      p.InitialStock,
		CurrentStock:      p.CurrentStock,
		LowStockThreshold: p.LowStockThreshold,
		IsActive:          p.IsActive,
		IsOutOfStock:      p.IsOutOfStock,
		HasVariants:       p.HasVariants,
		TotalSold:         p.TotalSold,
		Images:            images,
		Variants:          variants,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// ── GET /api/public/products ──
// Lists all active products across all shops with pagination and search.
func (h *Handler) ListProducts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	search := c.Query("q")
	shopID := c.Query("shop_id")
	sortBy := c.Query("sort", "created_at")

	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	query := h.DB.Model(&models.Product{}).Where("is_active = true")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR name_bn ILIKE ? OR sku ILIKE ?", like, like, like)
	}
	if shopID != "" {
		sid, err := uuid.Parse(shopID)
		if err == nil {
			query = query.Where("shop_id = ?", sid)
		}
	}

	// Validate sort column to prevent SQL injection
	allowedSort := map[string]string{
		"created_at": "products.created_at DESC",
		"price_asc":  "products.price ASC",
		"price_desc": "products.price DESC",
		"popular":    "products.total_sold DESC",
	}
	orderClause, ok := allowedSort[sortBy]
	if !ok {
		orderClause = "products.created_at DESC"
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	query.Preload("Variants").
		Order(orderClause).
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&products)

	// Collect shop IDs for batch fetch
	shopIDSet := make(map[uuid.UUID]struct{})
	for _, p := range products {
		shopIDSet[p.ShopID] = struct{}{}
	}
	shopIDs := make([]uuid.UUID, 0, len(shopIDSet))
	for sid := range shopIDSet {
		shopIDs = append(shopIDs, sid)
	}

	var shops []models.Shop
	if len(shopIDs) > 0 {
		h.DB.Where("id IN ?", shopIDs).Find(&shops)
	}
	shopMap := make(map[uuid.UUID]*models.Shop, len(shops))
	for i := range shops {
		shopMap[shops[i].ID] = &shops[i]
	}

	result := make([]PublicProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, buildProductResponse(p, shopMap[p.ShopID]))
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, result, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ── GET /api/public/products/:id ──
func (h *Handler) GetProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	pid, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid product ID")
	}

	var product models.Product
	if err := h.DB.Where("id = ? AND is_active = true", pid).
		Preload("Variants").
		First(&product).Error; err != nil {
		return response.NotFound(c, "Product not found")
	}

	var shop models.Shop
	h.DB.Where("id = ?", product.ShopID).First(&shop)

	return response.Success(c, buildProductResponse(product, &shop))
}

// ── GET /api/public/shops ──
func (h *Handler) ListShops(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	search := c.Query("q")

	if perPage > 50 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}

	query := h.DB.Model(&models.Shop{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("shop_name ILIKE ? OR shop_description ILIKE ?", like, like)
	}

	var total int64
	query.Count(&total)

	var shops []models.Shop
	query.Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&shops)

	// Get product counts per shop
	shopIDList := make([]uuid.UUID, 0, len(shops))
	for _, s := range shops {
		shopIDList = append(shopIDList, s.ID)
	}

	type countResult struct {
		ShopID uuid.UUID
		Count  int64
	}
	var counts []countResult
	if len(shopIDList) > 0 {
		h.DB.Model(&models.Product{}).
			Select("shop_id, COUNT(*) as count").
			Where("shop_id IN ? AND is_active = true", shopIDList).
			Group("shop_id").
			Scan(&counts)
	}
	countMap := make(map[uuid.UUID]int64)
	for _, cr := range counts {
		countMap[cr.ShopID] = cr.Count
	}

	result := make([]PublicShopResponse, 0, len(shops))
	for _, s := range shops {
		result = append(result, PublicShopResponse{
			ID:                    s.ID.String(),
			ShopName:              s.ShopName,
			ShopDescription:       s.ShopDescription,
			ShopLogoURL:           s.ShopLogoURL,
			District:              s.District,
			BkashMerchantNumber:   s.BkashMerchantNumber,
			NagadMerchantNumber:   s.NagadMerchantNumber,
			DeliveryChargeInside:  s.DeliveryChargeInside,
			DeliveryChargeOutside: s.DeliveryChargeOutside,
			CourierPreference:     s.CourierPreference,
			TotalProducts:         countMap[s.ID],
			CreatedAt:             s.CreatedAt,
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, result, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ── GET /api/public/shops/:id ──
func (h *Handler) GetShop(c *fiber.Ctx) error {
	idStr := c.Params("id")
	sid, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid shop ID")
	}

	var shop models.Shop
	if err := h.DB.Where("id = ?", sid).First(&shop).Error; err != nil {
		return response.NotFound(c, "Shop not found")
	}

	// Load the shop's active products
	var products []models.Product
	h.DB.Where("shop_id = ? AND is_active = true", sid).
		Preload("Variants").
		Order("created_at DESC").
		Limit(50).
		Find(&products)

	productDTOs := make([]PublicProductResponse, 0, len(products))
	for _, p := range products {
		productDTOs = append(productDTOs, buildProductResponse(p, &shop))
	}

	var totalProducts int64
	h.DB.Model(&models.Product{}).Where("shop_id = ? AND is_active = true", sid).Count(&totalProducts)

	return response.Success(c, PublicShopResponse{
		ID:                    shop.ID.String(),
		ShopName:              shop.ShopName,
		ShopDescription:       shop.ShopDescription,
		ShopLogoURL:           shop.ShopLogoURL,
		District:              shop.District,
		BkashMerchantNumber:   shop.BkashMerchantNumber,
		NagadMerchantNumber:   shop.NagadMerchantNumber,
		DeliveryChargeInside:  shop.DeliveryChargeInside,
		DeliveryChargeOutside: shop.DeliveryChargeOutside,
		CourierPreference:     shop.CourierPreference,
		TotalProducts:         totalProducts,
		Products:              productDTOs,
		CreatedAt:             shop.CreatedAt,
	})
}

// ── GET /api/public/search?q= ──
func (h *Handler) Search(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return response.BadRequest(c, "Search query is required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 50 {
		perPage = 50
	}

	like := "%" + q + "%"

	var total int64
	h.DB.Model(&models.Product{}).
		Where("is_active = true AND (name ILIKE ? OR name_bn ILIKE ? OR description ILIKE ?)", like, like, like).
		Count(&total)

	var products []models.Product
	h.DB.Where("is_active = true AND (name ILIKE ? OR name_bn ILIKE ? OR description ILIKE ?)", like, like, like).
		Preload("Variants").
		Order("total_sold DESC, created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&products)

	shopIDSet := make(map[uuid.UUID]struct{})
	for _, p := range products {
		shopIDSet[p.ShopID] = struct{}{}
	}
	shopIDs := make([]uuid.UUID, 0, len(shopIDSet))
	for sid := range shopIDSet {
		shopIDs = append(shopIDs, sid)
	}
	var shops []models.Shop
	if len(shopIDs) > 0 {
		h.DB.Where("id IN ?", shopIDs).Find(&shops)
	}
	shopMap := make(map[uuid.UUID]*models.Shop)
	for i := range shops {
		shopMap[shops[i].ID] = &shops[i]
	}

	result := make([]PublicProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, buildProductResponse(p, shopMap[p.ShopID]))
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, result, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ── POST /api/public/orders ── (Guest Checkout — no auth needed)
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	var req struct {
		ShopID          string `json:"shop_id"`
		CustomerName    string `json:"customer_name"`
		CustomerPhone   string `json:"customer_phone"`
		CustomerAddress string `json:"customer_address"`
		PaymentMethod   string `json:"payment_method"`
		PaymentTrxID    string `json:"payment_trx_id"`
		Notes           string `json:"notes"`
		OrderItems      []struct {
			ProductID string  `json:"product_id"`
			VariantID string  `json:"variant_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
			Name      string  `json:"name"`
		} `json:"order_items"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if req.ShopID == "" || req.CustomerName == "" || req.CustomerPhone == "" || req.CustomerAddress == "" {
		return response.BadRequest(c, "shop_id, customer_name, customer_phone, customer_address are required")
	}
	if len(req.OrderItems) == 0 {
		return response.BadRequest(c, "Order must have at least one item")
	}

	sid, err := uuid.Parse(req.ShopID)
	if err != nil {
		return response.BadRequest(c, "Invalid shop_id")
	}

	var shop models.Shop
	if err := h.DB.Where("id = ?", sid).First(&shop).Error; err != nil {
		return response.NotFound(c, "Shop not found")
	}

	// Calculate total from items
	var totalAmount float64
	type orderItemRaw struct {
		ProductID string  `json:"product_id"`
		VariantID *string `json:"variant_id,omitempty"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
		Name      string  `json:"name"`
	}
	rawItems := make([]orderItemRaw, 0, len(req.OrderItems))
	for _, item := range req.OrderItems {
		totalAmount += item.Price * float64(item.Quantity)
		var vid *string
		if item.VariantID != "" {
			vid = &item.VariantID
		}
		rawItems = append(rawItems, orderItemRaw{
			ProductID: item.ProductID,
			VariantID: vid,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Name:      item.Name,
		})
	}

	itemsJSON, _ := json.Marshal(rawItems)

	pm := models.OrderPaymentMethod(req.PaymentMethod)
	custName := req.CustomerName
	custPhone := req.CustomerPhone
	custAddr := req.CustomerAddress
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	order := models.Order{
		ShopID:          sid,
		CustomerName:    &custName,
		CustomerPhone:   &custPhone,
		CustomerAddress: &custAddr,
		OrderItems:      itemsJSON,
		TotalAmount:     totalAmount,
		PaymentMethod:   &pm,
		PaymentStatus:   models.OrderPayPending,
		DeliveryStatus:  models.DeliveryPending,
		PlacedBy:        models.PlacedByHuman,
		Notes:           notes,
	}

	// If payment TrxID provided, mark as manual_required for seller to verify
	if req.PaymentTrxID != "" {
		order.PaymentStatus = models.OrderPayManualReq
		order.PaymentTrxID = &req.PaymentTrxID
	}

	tx := h.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return response.InternalError(c)
	}

	// Decrement stock for each item
	for _, item := range req.OrderItems {
		pid, _ := uuid.Parse(item.ProductID)
		if item.VariantID != "" {
			vid, err := uuid.Parse(item.VariantID)
			if err == nil {
				tx.Model(&models.ProductVariant{}).
					Where("id = ? AND product_id = ?", vid, pid).
					UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			}
		} else {
			tx.Model(&models.Product{}).
				Where("id = ?", pid).
				Updates(map[string]interface{}{
					"current_stock":   gorm.Expr("current_stock - ?", item.Quantity),
					"total_sold":      gorm.Expr("total_sold + ?", item.Quantity),
					"is_out_of_stock": gorm.Expr("CASE WHEN current_stock - ? <= 0 THEN true ELSE false END", item.Quantity),
				})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return response.InternalError(c)
	}

	return response.Created(c, map[string]interface{}{
		"order_id":        order.ID.String(),
		"total_amount":    order.TotalAmount,
		"payment_status":  order.PaymentStatus,
		"delivery_status": order.DeliveryStatus,
		"created_at":      order.CreatedAt,
		"message":         "Order placed successfully! The seller will verify your payment.",
	})
}

// ── GET /api/public/orders/:id ── (Track order by ID)
func (h *Handler) TrackOrder(c *fiber.Ctx) error {
	idStr := c.Params("id")
	oid, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "Invalid order ID")
	}

	var order models.Order
	if err := h.DB.Where("id = ?", oid).First(&order).Error; err != nil {
		return response.NotFound(c, "Order not found")
	}

	var shop models.Shop
	h.DB.Where("id = ?", order.ShopID).First(&shop)

	return response.Success(c, map[string]interface{}{
		"order_id":        order.ID.String(),
		"shop_name":       shop.ShopName,
		"customer_name":   order.CustomerName,
		"customer_phone":  order.CustomerPhone,
		"total_amount":    order.TotalAmount,
		"payment_method":  order.PaymentMethod,
		"payment_status":  order.PaymentStatus,
		"delivery_status": order.DeliveryStatus,
		"courier_name":    order.CourierName,
		"tracking_number": order.TrackingNumber,
		"order_items":     order.OrderItems,
		"notes":           order.Notes,
		"created_at":      order.CreatedAt,
		"updated_at":      order.UpdatedAt,
	})
}
