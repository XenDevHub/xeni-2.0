package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Shop ──

// Shop represents the shops table.
type Shop struct {
	ID                  uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID              uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	ShopName            string    `gorm:"size:255;not null" json:"shop_name"`
	ShopDescription     *string   `gorm:"type:text" json:"shop_description"`
	ShopLogoURL         *string   `gorm:"type:text" json:"shop_logo_url"`
	ShopBannerURL       *string   `gorm:"type:text" json:"shop_banner_url"`
	PreferredLanguage   string    `gorm:"size:5;default:'bn';not null" json:"preferred_language"`
	CourierPreference   string    `gorm:"size:20;default:'pathao';not null" json:"courier_preference"`
	BkashMerchantNumber *string   `gorm:"size:20" json:"bkash_merchant_number"`
	NagadMerchantNumber *string   `gorm:"size:20" json:"nagad_merchant_number"`
	WhatsAppNumber      *string   `gorm:"column:whatsapp_number;size:20" json:"whatsapp_number"`
	OwnerMobile         *string   `gorm:"size:20" json:"owner_mobile"`
	District            *string   `gorm:"size:100" json:"district"`
	DeliveryChargeInside  float64 `gorm:"type:decimal(8,2);default:60;not null" json:"delivery_charge_inside"`
	DeliveryChargeOutside float64 `gorm:"type:decimal(8,2);default:120;not null" json:"delivery_charge_outside"`
	PaymentVerificationMode string `gorm:"size:10;default:'manual';not null" json:"payment_verification_mode"`
	BkashAppKey         *string   `gorm:"size:255" json:"bkash_app_key"`
	BkashAppSecret      *string   `gorm:"size:255" json:"-"`
	BkashUsername       *string   `gorm:"size:100" json:"-"`
	BkashPassword       *string   `gorm:"size:255" json:"-"`
	NagadMerchantID     *string   `gorm:"size:255" json:"nagad_merchant_id"`
	NagadMerchantKey    *string   `gorm:"size:255" json:"-"`
	SSLCommerzStoreID   *string   `gorm:"column:sslcommerz_store_id;size:255" json:"sslcommerz_store_id"`
	SSLCommerzStorePass *string   `gorm:"column:sslcommerz_store_passwd;size:255" json:"-"`
	SSLCommerzIsLive    bool      `gorm:"column:sslcommerz_is_live;default:false;not null" json:"sslcommerz_is_live"`
	PathaoClientID      *string   `gorm:"size:255" json:"-"`
	PathaoClientSecret  *string   `gorm:"size:255" json:"-"`
	PathaoUsername      *string   `gorm:"size:100" json:"-"`
	PathaoPassword      *string   `gorm:"size:255" json:"-"`
	SteadfastAPIKey     *string   `gorm:"size:255" json:"-"`
	SteadfastSecretKey  *string   `gorm:"size:255" json:"-"`
	AutoReplyEnabled    bool      `gorm:"default:true;not null" json:"auto_reply_enabled"`
	AutoOrderEnabled    bool      `gorm:"default:true;not null" json:"auto_order_enabled"`
	Integrations        JSON      `gorm:"type:jsonb;default:'{}'" json:"integrations"`
	CustomAgentRules    *string   `gorm:"type:text" json:"custom_agent_rules"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User           User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ConnectedPages []ConnectedPage `gorm:"foreignKey:ShopID" json:"connected_pages,omitempty"`
	Products       []Product       `gorm:"foreignKey:ShopID" json:"products,omitempty"`
}

func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ── ConnectedPage ──

// ConnectedPage represents the connected_pages table.
type ConnectedPage struct {
	ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ShopID            uuid.UUID `gorm:"type:uuid;not null;index" json:"shop_id"`
	PageID            string    `gorm:"size:255;not null;uniqueIndex" json:"page_id"`
	PageName          string    `gorm:"size:255;not null" json:"page_name"`
	PageAccessToken   string    `gorm:"type:text;not null" json:"-"`
	PagePictureURL    *string   `gorm:"type:text" json:"page_picture_url"`
	WebhookSubscribed bool      `gorm:"default:false;not null" json:"webhook_subscribed"`
	IsActive          bool      `gorm:"default:true;not null" json:"is_active"`
	ConnectedAt       time.Time `gorm:"autoCreateTime" json:"connected_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Shop Shop `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"-"`
}

func (cp *ConnectedPage) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}

// ── Product ──

// Product represents the products table.
type Product struct {
	ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ShopID            uuid.UUID `gorm:"type:uuid;not null;index" json:"shop_id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	NameBN            *string   `gorm:"column:name_bn;size:255" json:"name_bn"`
	Description       *string   `gorm:"type:text" json:"description"`
	DescriptionBN     *string   `gorm:"column:description_bn;type:text" json:"description_bn"`
	Price             float64   `gorm:"type:decimal(12,2);default:0;not null" json:"price"`
	SKU               *string   `gorm:"size:100" json:"sku"`
	InitialStock      int       `gorm:"default:0;not null" json:"initial_stock"`
	CurrentStock      int       `gorm:"default:0;not null" json:"current_stock"`
	LowStockThreshold int       `gorm:"default:5;not null" json:"low_stock_threshold"`
	IsActive          bool      `gorm:"default:true;not null" json:"is_active"`
	IsOutOfStock      bool      `gorm:"default:false;not null" json:"is_out_of_stock"`
	HasVariants       bool      `gorm:"default:false;not null" json:"has_variants"`
	TotalSold         int       `gorm:"default:0;not null" json:"total_sold"`
	Images            JSON      `gorm:"type:jsonb;default:'[]'" json:"images"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Shop          Shop             `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"-"`
	Variants      []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	InventoryLogs []InventoryLog   `gorm:"foreignKey:ProductID" json:"inventory_logs,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ── Order ──

type OrderPaymentMethod string

const (
	PaymentBkash OrderPaymentMethod = "bkash"
	PaymentNagad OrderPaymentMethod = "nagad"
	PaymentCOD   OrderPaymentMethod = "cod"
	PaymentSSLCommerz OrderPaymentMethod = "sslcommerz"
)

type OrderPaymentStatus string

const (
	OrderPayPending   OrderPaymentStatus = "pending"
	OrderPayVerified  OrderPaymentStatus = "verified"
	OrderPayFailed    OrderPaymentStatus = "failed"
	OrderPayManualReq OrderPaymentStatus = "manual_required"
)

type OrderDeliveryStatus string

const (
	DeliveryPending   OrderDeliveryStatus = "pending"
	DeliveryBooked    OrderDeliveryStatus = "booked"
	DeliveryInTransit OrderDeliveryStatus = "in_transit"
	DeliveryDelivered OrderDeliveryStatus = "delivered"
	DeliveryReturned  OrderDeliveryStatus = "returned"
)

type OrderPlacedBy string

const (
	PlacedByAI    OrderPlacedBy = "ai"
	PlacedByHuman OrderPlacedBy = "human"
)

// Order represents the orders table.
type Order struct {
	ID                     uuid.UUID           `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ShopID                 uuid.UUID           `gorm:"type:uuid;not null;index" json:"shop_id"`
	CustomerPSID           *string             `gorm:"column:customer_psid;size:255" json:"customer_psid"`
	CustomerName           *string             `gorm:"size:255" json:"customer_name"`
	CustomerPhone          *string             `gorm:"size:20" json:"customer_phone"`
	CustomerAddress        *string             `gorm:"type:text" json:"customer_address"`
	OrderItems             JSON                `gorm:"type:jsonb;default:'[]';not null" json:"order_items"`
	TotalAmount            float64             `gorm:"type:decimal(12,2);default:0;not null" json:"total_amount"`
	PaymentMethod          *OrderPaymentMethod `gorm:"type:order_payment_method" json:"payment_method"`
	PaymentStatus          OrderPaymentStatus  `gorm:"type:order_payment_status;default:'pending';not null" json:"payment_status"`
	PaymentTrxID           *string             `gorm:"size:255" json:"payment_trx_id"`
	PaymentScreenshotURL   *string             `gorm:"type:text" json:"payment_screenshot_url"`
	DeliveryStatus         OrderDeliveryStatus `gorm:"type:order_delivery_status;default:'pending';not null" json:"delivery_status"`
	CourierName            *string             `gorm:"size:50" json:"courier_name"`
	TrackingNumber         *string             `gorm:"size:255" json:"tracking_number"`
	CourierBookingResponse JSON                `gorm:"type:jsonb" json:"courier_booking_response"`
	MessengerThreadID      *uuid.UUID          `gorm:"type:uuid" json:"messenger_thread_id"`
	PlacedBy               OrderPlacedBy       `gorm:"type:order_placed_by;default:'human';not null" json:"placed_by"`
	Notes                  *string             `gorm:"type:text" json:"notes"`
	VerifiedBy             *string             `gorm:"size:20" json:"verified_by"`
	VerifiedAt             *time.Time          `json:"verified_at"`
	AdminNote              *string             `gorm:"type:text" json:"admin_note"`
	CreatedAt              time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time           `gorm:"autoUpdateTime" json:"updated_at"`

	Shop Shop `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"-"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// ── Conversation ──

type ConversationHandlingMode string

const (
	HandlingModeAI    ConversationHandlingMode = "ai"
	HandlingModeHuman ConversationHandlingMode = "human"
)

type ConversationStatus string

const (
	ConversationOpen     ConversationStatus = "open"
	ConversationResolved ConversationStatus = "resolved"
)

// Conversation represents the conversations table.
type Conversation struct {
	ID                 uuid.UUID                `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ShopID             uuid.UUID                `gorm:"type:uuid;not null;index" json:"shop_id"`
	PageID             string                   `gorm:"size:255;not null;index" json:"page_id"`
	CustomerPSID       string                   `gorm:"column:customer_psid;size:255;not null" json:"customer_psid"`
	CustomerName       *string                  `gorm:"size:255" json:"customer_name"`
	LastMessagePreview *string                  `gorm:"type:text" json:"last_message_preview"`
	LastMessageAt      *time.Time               `json:"last_message_at"`
	UnreadCount        int                      `gorm:"default:0;not null" json:"unread_count"`
	HandlingMode       ConversationHandlingMode `gorm:"type:conversation_handling_mode;default:'ai';not null" json:"handling_mode"`
	Status             ConversationStatus       `gorm:"type:conversation_status;default:'open';not null" json:"status"`
	CreatedAt          time.Time                `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time                `gorm:"autoUpdateTime" json:"updated_at"`

	Shop     Shop      `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"-"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (cv *Conversation) BeforeCreate(tx *gorm.DB) error {
	if cv.ID == uuid.Nil {
		cv.ID = uuid.New()
	}
	return nil
}

// ── Message ──

type MessageDirection string

const (
	DirectionInbound  MessageDirection = "inbound"
	DirectionOutbound MessageDirection = "outbound"
)

type MessageSenderType string

const (
	SenderCustomer MessageSenderType = "customer"
	SenderAI       MessageSenderType = "ai"
	SenderHuman    MessageSenderType = "human"
)

type MessageContentType string

const (
	ContentText  MessageContentType = "text"
	ContentImage MessageContentType = "image"
	ContentAudio MessageContentType = "audio"
)

// Message represents the messages table.
type Message struct {
	ID             uuid.UUID          `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ConversationID uuid.UUID          `gorm:"type:uuid;not null;index" json:"conversation_id"`
	Direction      MessageDirection   `gorm:"type:message_direction;not null" json:"direction"`
	SenderType     MessageSenderType  `gorm:"type:message_sender_type;not null" json:"sender_type"`
	ContentType    MessageContentType `gorm:"type:message_content_type;default:'text';not null" json:"content_type"`
	ContentText    *string            `gorm:"type:text" json:"content_text"`
	ContentURL     *string            `gorm:"type:text" json:"content_url"`
	MessengerMID   *string            `gorm:"column:messenger_mid;size:255;uniqueIndex" json:"messenger_mid"`
	SentAt         time.Time          `gorm:"autoCreateTime" json:"sent_at"`
	CreatedAt      time.Time          `gorm:"autoCreateTime" json:"created_at"`

	Conversation Conversation `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
