-- ============================================================
-- XENI AI — Autonomous E-Commerce Command Center
-- PostgreSQL Schema — Full Database Initialization
-- ============================================================
-- Run automatically via docker-entrypoint-initdb.d
-- Creates all 14 tables, indexes, triggers, and seed data
-- ============================================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ──────────────────────────────────────────────────────────
-- ENUM TYPES
-- ──────────────────────────────────────────────────────────

CREATE TYPE user_role AS ENUM ('user', 'admin', 'super_admin');
CREATE TYPE user_status AS ENUM ('pending', 'active', 'suspended');
CREATE TYPE auth_provider AS ENUM ('email', 'google', 'facebook');
CREATE TYPE otp_purpose AS ENUM ('email_verify', 'password_reset', 'two_fa');
CREATE TYPE plan_tier AS ENUM ('starter', 'professional', 'premium', 'enterprise');
CREATE TYPE subscription_status AS ENUM ('active', 'past_due', 'cancelled', 'trialing');
CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed', 'refunded');
CREATE TYPE agent_type AS ENUM ('conversation', 'order', 'inventory', 'creative', 'intelligence');
CREATE TYPE task_status AS ENUM ('queued', 'processing', 'completed', 'failed');
CREATE TYPE order_payment_method AS ENUM ('bkash', 'nagad', 'cod', 'sslcommerz');
CREATE TYPE order_payment_status AS ENUM ('pending', 'verified', 'failed', 'manual_required');
CREATE TYPE order_delivery_status AS ENUM ('pending', 'booked', 'in_transit', 'delivered', 'returned');
CREATE TYPE message_direction AS ENUM ('inbound', 'outbound');
CREATE TYPE message_sender_type AS ENUM ('customer', 'ai', 'human');
CREATE TYPE message_content_type AS ENUM ('text', 'image', 'audio');
CREATE TYPE conversation_handling_mode AS ENUM ('ai', 'human');
CREATE TYPE conversation_status AS ENUM ('open', 'resolved');
CREATE TYPE order_placed_by AS ENUM ('ai', 'human');
CREATE TYPE stock_movement_type AS ENUM ('sale', 'restock', 'adjustment', 'return');
CREATE TYPE review_status AS ENUM ('pending', 'approved', 'rejected');

-- ──────────────────────────────────────────────────────────
-- 1. USERS
-- ──────────────────────────────────────────────────────────

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email               VARCHAR(255) NOT NULL UNIQUE,
    password_hash       VARCHAR(255),                              -- NULL for OAuth users
    full_name           VARCHAR(255) NOT NULL,
    avatar_url          TEXT,
    role                user_role NOT NULL DEFAULT 'user',
    status              user_status NOT NULL DEFAULT 'pending',
    auth_provider       auth_provider NOT NULL DEFAULT 'email',
    google_id           VARCHAR(255) UNIQUE,
    facebook_id         VARCHAR(255) UNIQUE,                       -- Facebook Login user ID
    is_email_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    two_fa_enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    two_fa_secret       VARCHAR(255),                              -- encrypted TOTP secret
    preferred_language  VARCHAR(5) NOT NULL DEFAULT 'en',          -- "en" | "bn"
    last_login_at       TIMESTAMPTZ,
    suspended_reason    TEXT,
    whatsapp_number     VARCHAR(20),
    suspended_at        TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    admin_note          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX idx_users_facebook_id ON users(facebook_id) WHERE facebook_id IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- ──────────────────────────────────────────────────────────
-- 2. REFRESH TOKENS
-- ──────────────────────────────────────────────────────────

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,                         -- SHA-256 hash
    device_info     TEXT,
    ip_address      VARCHAR(45),                                   -- supports IPv6
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ──────────────────────────────────────────────────────────
-- 3. OTP CODES
-- ──────────────────────────────────────────────────────────

CREATE TABLE otp_codes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash       VARCHAR(255) NOT NULL,                         -- bcrypt hash of OTP
    purpose         otp_purpose NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    used            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_codes_user_id ON otp_codes(user_id);
CREATE INDEX idx_otp_codes_purpose ON otp_codes(user_id, purpose);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);

-- ──────────────────────────────────────────────────────────
-- 4. SHOPS
-- ──────────────────────────────────────────────────────────

CREATE TABLE shops (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    shop_name               VARCHAR(255) NOT NULL,
    shop_description        TEXT,
    shop_logo_url           TEXT,
    shop_banner_url         TEXT,
    preferred_language      VARCHAR(5) NOT NULL DEFAULT 'bn',      -- "bn" | "en"
    courier_preference      VARCHAR(20) NOT NULL DEFAULT 'pathao', -- "pathao" | "steadfast"
    bkash_merchant_number   VARCHAR(20),
    nagad_merchant_number   VARCHAR(20),
    sslcommerz_store_id     VARCHAR(255),
    sslcommerz_store_passwd VARCHAR(255),
    sslcommerz_is_live      BOOLEAN NOT NULL DEFAULT FALSE,
    auto_reply_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    auto_order_enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shops_user_id ON shops(user_id);

-- ──────────────────────────────────────────────────────────
-- 5. CONNECTED PAGES (Facebook Page connections)
-- ──────────────────────────────────────────────────────────

CREATE TABLE connected_pages (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    page_id             VARCHAR(255) NOT NULL,                     -- Facebook Page ID
    page_name           VARCHAR(255) NOT NULL,
    page_access_token   TEXT NOT NULL,                             -- AES-256 encrypted
    page_picture_url    TEXT,
    webhook_subscribed  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    connected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_connected_pages_shop_id ON connected_pages(shop_id);
CREATE UNIQUE INDEX idx_connected_pages_page_id ON connected_pages(page_id);

-- ──────────────────────────────────────────────────────────
-- 6. PRODUCTS
-- ──────────────────────────────────────────────────────────

CREATE TABLE products (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    name_bn             VARCHAR(255),                              -- Bangla product name
    description         TEXT,
    description_bn      TEXT,                                      -- Bangla description
    price               DECIMAL(12, 2) NOT NULL DEFAULT 0,         -- BDT
    sku                 VARCHAR(100),
    initial_stock       INTEGER NOT NULL DEFAULT 0,
    current_stock       INTEGER NOT NULL DEFAULT 0,
    low_stock_threshold INTEGER NOT NULL DEFAULT 5,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    is_out_of_stock     BOOLEAN NOT NULL DEFAULT FALSE,
    has_variants        BOOLEAN NOT NULL DEFAULT FALSE,
    total_sold          INTEGER NOT NULL DEFAULT 0,
    images              JSONB DEFAULT '[]'::JSONB,                 -- array of S3 URLs
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_shop_id ON products(shop_id);
CREATE INDEX idx_products_sku ON products(shop_id, sku) WHERE sku IS NOT NULL;
CREATE INDEX idx_products_active ON products(shop_id, is_active);
CREATE INDEX idx_products_stock ON products(shop_id, is_out_of_stock);

-- ──────────────────────────────────────────────────────────
-- 6a. PRODUCT VARIANTS
-- ──────────────────────────────────────────────────────────

CREATE TABLE product_variants (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku           VARCHAR(100) NOT NULL,
    color         VARCHAR(50),
    size          VARCHAR(50),
    price_modifier DECIMAL(12, 2) DEFAULT 0,
    stock         INTEGER NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);

-- ──────────────────────────────────────────────────────────
-- 6b. INVENTORY LOGS
-- ──────────────────────────────────────────────────────────

CREATE TABLE inventory_logs (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id   UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id   UUID REFERENCES product_variants(id) ON DELETE SET NULL,
    type         stock_movement_type NOT NULL,
    quantity     INTEGER NOT NULL,
    old_stock    INTEGER NOT NULL,
    new_stock    INTEGER NOT NULL,
    reference_id VARCHAR(255),
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_logs_product_id ON inventory_logs(product_id);
CREATE INDEX idx_inventory_logs_variant_id ON inventory_logs(variant_id);
CREATE INDEX idx_inventory_logs_created_at ON inventory_logs(created_at DESC);

-- ──────────────────────────────────────────────────────────
-- 7. ORDERS
-- ──────────────────────────────────────────────────────────

CREATE TABLE orders (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id                 UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_psid           VARCHAR(255),                          -- Facebook Page-Scoped ID
    customer_name           VARCHAR(255),
    customer_phone          VARCHAR(20),
    customer_address        TEXT,
    order_items             JSONB NOT NULL DEFAULT '[]'::JSONB,    -- [{product_id, quantity, unit_price}]
    total_amount            DECIMAL(12, 2) NOT NULL DEFAULT 0,     -- BDT
    payment_method          order_payment_method,
    payment_status          order_payment_status NOT NULL DEFAULT 'pending',
    payment_trx_id          VARCHAR(255),                          -- bKash/Nagad TrxID
    payment_screenshot_url  TEXT,                                   -- S3 URL
    delivery_status         order_delivery_status NOT NULL DEFAULT 'pending',
    courier_name            VARCHAR(50),                           -- "pathao" | "steadfast"
    tracking_number         VARCHAR(255),
    courier_booking_response JSONB,                                -- raw courier API response
    messenger_thread_id     UUID,                                  -- FK to conversations (optional)
    placed_by               order_placed_by NOT NULL DEFAULT 'human',
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_shop_id ON orders(shop_id);
CREATE INDEX idx_orders_customer_psid ON orders(customer_psid);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_delivery_status ON orders(delivery_status);
CREATE INDEX idx_orders_payment_trx_id ON orders(payment_trx_id) WHERE payment_trx_id IS NOT NULL;
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_shop_created ON orders(shop_id, created_at DESC);

-- ──────────────────────────────────────────────────────────
-- 8. CONVERSATIONS (Messenger threads)
-- ──────────────────────────────────────────────────────────

CREATE TABLE conversations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    page_id             VARCHAR(255) NOT NULL,                     -- Facebook Page ID
    customer_psid       VARCHAR(255) NOT NULL,                     -- Page-Scoped ID
    customer_name       VARCHAR(255),
    last_message_preview TEXT,
    last_message_at     TIMESTAMPTZ,
    unread_count        INTEGER NOT NULL DEFAULT 0,
    handling_mode       conversation_handling_mode NOT NULL DEFAULT 'ai',
    status              conversation_status NOT NULL DEFAULT 'open',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_shop_id ON conversations(shop_id);
CREATE INDEX idx_conversations_page_id ON conversations(page_id);
CREATE UNIQUE INDEX idx_conversations_page_psid ON conversations(page_id, customer_psid);
CREATE INDEX idx_conversations_status ON conversations(shop_id, status);
CREATE INDEX idx_conversations_last_msg ON conversations(shop_id, last_message_at DESC);

-- ──────────────────────────────────────────────────────────
-- 9. MESSAGES
-- ──────────────────────────────────────────────────────────

CREATE TABLE messages (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id     UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    direction           message_direction NOT NULL,                -- inbound | outbound
    sender_type         message_sender_type NOT NULL,              -- customer | ai | human
    content_type        message_content_type NOT NULL DEFAULT 'text',
    content_text        TEXT,
    content_url         TEXT,                                       -- image/audio URL
    messenger_mid       VARCHAR(255) UNIQUE,                       -- Meta message ID for dedup
    sent_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sent_at ON messages(conversation_id, sent_at DESC);
CREATE INDEX idx_messages_messenger_mid ON messages(messenger_mid) WHERE messenger_mid IS NOT NULL;

-- ──────────────────────────────────────────────────────────
-- 10. PLANS
-- ──────────────────────────────────────────────────────────

CREATE TABLE plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL,
    tier            plan_tier NOT NULL UNIQUE,
    price_monthly_bdt DECIMAL(12, 2) NOT NULL DEFAULT 0,           -- BDT per month
    features        JSONB NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    tagline         VARCHAR(100),
    tagline_bn      VARCHAR(200),
    cta_text        VARCHAR(50) DEFAULT 'Get Started',
    cta_text_bn     VARCHAR(100) DEFAULT 'শুরু করুন',
    is_most_popular  BOOLEAN DEFAULT FALSE,
    display_order   INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ──────────────────────────────────────────────────────────
-- 11. SUBSCRIPTIONS
-- ──────────────────────────────────────────────────────────

CREATE TABLE subscriptions (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    plan_id                 UUID NOT NULL REFERENCES plans(id),
    status                  subscription_status NOT NULL DEFAULT 'active',
    billing_cycle           VARCHAR(20) NOT NULL DEFAULT 'monthly',
    current_period_start    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end      TIMESTAMPTZ NOT NULL,
    cancelled_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- ──────────────────────────────────────────────────────────
-- 12. PAYMENTS
-- ──────────────────────────────────────────────────────────

CREATE TABLE payments (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id         UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    plan_id                 UUID NOT NULL REFERENCES plans(id),
    amount                  DECIMAL(12, 2) NOT NULL,
    currency                VARCHAR(3) NOT NULL DEFAULT 'BDT',
    status                  payment_status NOT NULL DEFAULT 'pending',
    gateway                 VARCHAR(20) NOT NULL DEFAULT 'sslcommerz',
    gateway_transaction_id  VARCHAR(255),                           -- SSLCommerz tran_id
    gateway_response        JSONB,                                  -- raw SSLCommerz response
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_gateway_tx ON payments(gateway_transaction_id) WHERE gateway_transaction_id IS NOT NULL;

-- ──────────────────────────────────────────────────────────
-- 13. AGENT TASKS
-- ──────────────────────────────────────────────────────────

CREATE TABLE agent_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_id         UUID REFERENCES shops(id) ON DELETE SET NULL,
    agent_type      agent_type NOT NULL,
    task_id         UUID NOT NULL UNIQUE,                           -- correlates with RabbitMQ/MongoDB
    status          task_status NOT NULL DEFAULT 'queued',
    mongo_doc_id    VARCHAR(255),                                   -- MongoDB ObjectId string
    error_message   TEXT,
    duration_ms     INTEGER,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_agent_tasks_user_id ON agent_tasks(user_id);
CREATE INDEX idx_agent_tasks_shop_id ON agent_tasks(shop_id);
CREATE INDEX idx_agent_tasks_task_id ON agent_tasks(task_id);
CREATE INDEX idx_agent_tasks_agent_type ON agent_tasks(agent_type);
CREATE INDEX idx_agent_tasks_status ON agent_tasks(status);
CREATE INDEX idx_agent_tasks_user_agent ON agent_tasks(user_id, agent_type);
CREATE INDEX idx_agent_tasks_created_at ON agent_tasks(created_at DESC);

-- ──────────────────────────────────────────────────────────
-- 14. AUDIT LOGS
-- ──────────────────────────────────────────────────────────

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    action          VARCHAR(255) NOT NULL,                          -- e.g. "user.login", "order.created"
    resource        VARCHAR(255),                                   -- e.g. "users", "orders"
    metadata        JSONB DEFAULT '{}',
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ──────────────────────────────────────────────────────────
-- 15. CONTENT SECTIONS
-- ──────────────────────────────────────────────────────────

CREATE TABLE content_sections (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    section_key VARCHAR(50) NOT NULL UNIQUE,
    content_en JSONB NOT NULL DEFAULT '{}',
    content_bn JSONB NOT NULL DEFAULT '{}',
    is_active  BOOLEAN DEFAULT TRUE,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_content_sections_key ON content_sections(section_key);

-- ──────────────────────────────────────────────────────────
-- 16. REVIEWS
-- ──────────────────────────────────────────────────────────

CREATE TABLE reviews (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewer_name     VARCHAR(100) NOT NULL,
    reviewer_avatar_url TEXT,
    plan_at_review     VARCHAR(50),
    star_rating       SMALLINT NOT NULL,
    review_text       TEXT NOT NULL,
    status           review_status NOT NULL DEFAULT 'pending',
    display_order     INTEGER DEFAULT 0,
    moderated_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    moderated_at      TIMESTAMPTZ,
    admin_note        TEXT,
    show_on_landing    BOOLEAN DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ──────────────────────────────────────────────────────────
-- 17. REVIEW SETTINGS
-- ──────────────────────────────────────────────────────────

CREATE TABLE review_settings (
    id                  INTEGER PRIMARY KEY DEFAULT 1,
    auto_approve_premium  BOOLEAN DEFAULT FALSE,
    show_star_rating      BOOLEAN DEFAULT TRUE,
    min_star_to_show       SMALLINT DEFAULT 4,
    max_reviews_on_landing INTEGER DEFAULT 6,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ──────────────────────────────────────────────────────────
-- 18. PLATFORM METRICS CACHE
-- ──────────────────────────────────────────────────────────

CREATE TABLE platform_metrics_cache (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_date           DATE NOT NULL UNIQUE,
    total_users           INTEGER DEFAULT 0,
    new_users_today        INTEGER DEFAULT 0,
    active_subscriptions  INTEGER DEFAULT 0,
    revenue_today         DECIMAL(12, 2) DEFAULT 0,
    revenue_month         DECIMAL(12, 2) DEFAULT 0,
    ai_tasks_today         INTEGER DEFAULT 0,
    messages_replied_today INTEGER DEFAULT 0,
    orders_processed_today INTEGER DEFAULT 0,
    task_success_rate      DECIMAL(5, 2) DEFAULT 0,
    computed_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ──────────────────────────────────────────────────────────
-- 19. SYSTEM SETTINGS
-- ──────────────────────────────────────────────────────────

CREATE TABLE system_settings (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_key   VARCHAR(255) NOT NULL UNIQUE,
    setting_value TEXT,
    description   TEXT,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by    UUID REFERENCES users(id) ON DELETE SET NULL
);

-- ──────────────────────────────────────────────────────────
-- 20. AGENT RULES
-- ──────────────────────────────────────────────────────────

CREATE TABLE agent_rules (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scope     VARCHAR(20) NOT NULL,
    shop_id   UUID REFERENCES shops(id) ON DELETE CASCADE,
    category  VARCHAR(50) NOT NULL,
    title     VARCHAR(255) NOT NULL,
    rule      TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    priority  INTEGER DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_agent_rules_scope ON agent_rules(scope);
CREATE INDEX idx_agent_rules_shop_id ON agent_rules(shop_id);

-- ──────────────────────────────────────────────────────────
-- UPDATED_AT TRIGGERS
-- ──────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_shops_updated_at
    BEFORE UPDATE ON shops
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_connected_pages_updated_at
    BEFORE UPDATE ON connected_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_plans_updated_at
    BEFORE UPDATE ON plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_product_variants_updated_at
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_content_sections_updated_at
    BEFORE UPDATE ON content_sections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_agent_rules_updated_at
    BEFORE UPDATE ON agent_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ──────────────────────────────────────────────────────────
-- SEED DATA: SUBSCRIPTION PLANS (BDT only)
-- ──────────────────────────────────────────────────────────

INSERT INTO plans (name, tier, price_monthly_bdt, features) VALUES
(
    'Starter',
    'starter',
    2500,
    '{
        "agents": ["conversation"],
        "max_orders_per_month": 200,
        "max_pages": 1,
        "storage_gb": 2,
        "description": "Conversation Agent only — auto-reply to Messenger DMs 24/7",
        "description_bn": "শুধুমাত্র কথোপকথন এজেন্ট — ২৪/৭ Messenger DM-এ স্বয়ংক্রিয় উত্তর"
    }'::JSONB
),
(
    'Professional',
    'professional',
    7500,
    '{
        "agents": ["conversation", "order", "inventory"],
        "max_orders_per_month": 1000,
        "max_pages": 3,
        "storage_gb": 10,
        "description": "Conversation + Order Processing + Inventory management",
        "description_bn": "কথোপকথন + অর্ডার প্রসেসিং + ইনভেন্টরি ম্যানেজমেন্ট"
    }'::JSONB
),
(
    'Premium',
    'premium',
    25000,
    '{
        "agents": ["conversation", "order", "inventory", "creative", "intelligence"],
        "max_orders_per_month": -1,
        "max_pages": 10,
        "storage_gb": 50,
        "description": "All 5 agents + AI image gen + auto-posting + sales intelligence",
        "description_bn": "সব ৫টি এজেন্ট + AI ছবি তৈরি + অটো-পোস্টিং + সেলস ইন্টেলিজেন্স"
    }'::JSONB
),
(
    'Enterprise',
    'enterprise',
    0,
    '{
        "agents": ["conversation", "order", "inventory", "creative", "intelligence"],
        "max_orders_per_month": -1,
        "max_pages": -1,
        "storage_gb": -1,
        "custom_pricing": true,
        "white_label": true,
        "api_access": true,
        "custom_erp": true,
        "description": "All agents + white-label + API access + custom ERP integration",
        "description_bn": "সব এজেন্ট + হোয়াইট-লেবেল + API অ্যাক্সেস + কাস্টম ERP ইন্টিগ্রেশন"
    }'::JSONB
);

-- ──────────────────────────────────────────────────────────
-- DONE
-- ──────────────────────────────────────────────────────────
-- Schema: 14 tables, 4 plans seeded
-- Total indexes: 40+
-- Triggers: 9 (auto update_at)
