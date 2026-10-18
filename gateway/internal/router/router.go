package router

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/xeni-ai/gateway/internal/admin"
	"github.com/xeni-ai/gateway/internal/agents"
	"github.com/xeni-ai/gateway/internal/auth"
	"github.com/xeni-ai/gateway/internal/billing"
	"github.com/xeni-ai/gateway/internal/cache"
	"github.com/xeni-ai/gateway/internal/config"
	"github.com/xeni-ai/gateway/internal/content"
	"github.com/xeni-ai/gateway/internal/conversations"
	"github.com/xeni-ai/gateway/internal/messenger"
	"github.com/xeni-ai/gateway/internal/middleware"
	"github.com/xeni-ai/gateway/internal/models"
	"github.com/xeni-ai/gateway/internal/notifications"
	"github.com/xeni-ai/gateway/internal/orders"
	"github.com/xeni-ai/gateway/internal/pages"
	"github.com/xeni-ai/gateway/internal/products"
	"github.com/xeni-ai/gateway/internal/public"
	"github.com/xeni-ai/gateway/internal/rabbitmq"
	"github.com/xeni-ai/gateway/internal/rules"
	"github.com/xeni-ai/gateway/internal/shop"
	"github.com/xeni-ai/gateway/internal/storage"
	"github.com/xeni-ai/gateway/internal/user"
	ws "github.com/xeni-ai/gateway/internal/websocket"
	"github.com/xeni-ai/gateway/pkg/email"
	jwtPkg "github.com/xeni-ai/gateway/pkg/jwt"

	"gorm.io/gorm"
)

// Setup configures all routes for the Fiber application.
func Setup(
	app *fiber.App,
	cfg *config.Config,
	db *gorm.DB,
	redis *cache.Client,
	jwtManager *jwtPkg.Manager,
	wsHub *ws.Hub,
	agentHandler *agents.Handler,
	rmqClient *rabbitmq.Client,
	spacesClient *storage.SpacesClient,
	notifSvc *notifications.Service,
) {
	// ── Global Middleware ──
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} ${path}\n",
	}))
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.SecurityHeadersMiddleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.FrontendURL,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))

	// ── Health Check ──
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "xeni-gateway"})
	})

	// Prometheus metrics endpoint
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.SendString("# HELP xeni_gateway_up Gateway is running\n# TYPE xeni_gateway_up gauge\nxeni_gateway_up 1\n")
	})

	api := app.Group("/api")

	// ── Public Marketplace Routes (no auth — E-Pic Marketplace customer-facing) ──
	publicHandler := public.NewHandler(db)
	publicGroup := api.Group("/public")
	publicGroup.Get("/products", publicHandler.ListProducts)
	publicGroup.Get("/products/:id", publicHandler.GetProduct)
	publicGroup.Get("/shops", publicHandler.ListShops)
	publicGroup.Get("/shops/:id", publicHandler.GetShop)
	publicGroup.Get("/search", publicHandler.Search)
	publicGroup.Post("/orders", publicHandler.CreateOrder)
	publicGroup.Get("/orders/:id", publicHandler.TrackOrder)
	publicGroup.Post("/payments/sslcommerz/init", publicHandler.InitSSLCommerzPayment)
	publicGroup.Post("/payments/sslcommerz/success", publicHandler.SSLCommerzSuccess)
	publicGroup.Post("/payments/sslcommerz/fail", publicHandler.SSLCommerzFail)
	publicGroup.Post("/payments/sslcommerz/cancel", publicHandler.SSLCommerzCancel)
	publicGroup.Post("/payments/sslcommerz/ipn", publicHandler.SSLCommerzIPN)

	// ── Messenger & WhatsApp Webhooks (no auth — called by Meta) ──
	messengerHandler := messenger.NewHandler(db, cfg, rmqClient)
	
	// Messenger
	api.Get("/webhooks/messenger", messengerHandler.WebhookVerify)
	api.Post("/webhooks/messenger", messengerHandler.WebhookReceive)
	
	// WhatsApp (Uses same verification logic as Messenger)
	api.Get("/webhooks/whatsapp", messengerHandler.WebhookVerify)
	api.Post("/webhooks/whatsapp", messengerHandler.WebhookReceive)

	// ── Auth Routes ──
	emailSvc := email.NewResendService(cfg.Email.ResendAPIKey, cfg.Email.FromEmail)
	authHandler := auth.NewHandler(db, redis, jwtManager, cfg.App.FrontendURL, emailSvc)
	authGroup := api.Group("/auth")
	authRateLimit := middleware.RateLimitMiddleware(redis, 5, time.Minute, "auth")

	authGroup.Post("/register", authRateLimit, authHandler.Register)
	authGroup.Post("/login", authRateLimit, authHandler.Login)
	authGroup.Post("/refresh", authRateLimit, authHandler.RefreshToken)
	authGroup.Post("/verify-email", authRateLimit, authHandler.VerifyEmail)
	authGroup.Post("/resend-otp", authRateLimit, authHandler.ResendOTP)
	authGroup.Post("/forgot-password", authRateLimit, authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authRateLimit, authHandler.ResetPassword)
	authGroup.Post("/google/callback", authRateLimit, authHandler.GoogleCallback)
	authGroup.Post("/facebook/callback", authRateLimit, authHandler.FacebookCallback)

	// Auth routes that require authentication
	authProtected := authGroup.Group("", middleware.AuthMiddleware(jwtManager, redis))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Post("/2fa/enable", authHandler.Enable2FA)
	authProtected.Post("/2fa/verify", authHandler.Verify2FA)

	// ── User Routes ──
	userHandler := user.NewHandler(db)
	userGroup := api.Group("/user", middleware.AuthMiddleware(jwtManager, redis))
	apiRateLimit := middleware.RateLimitMiddleware(redis, cfg.RateLimit.Requests, cfg.RateLimit.Window, "api")
	userGroup.Use(apiRateLimit)

	userGroup.Get("/me", userHandler.GetMe)
	userGroup.Put("/me", userHandler.UpdateMe)
	userGroup.Put("/me/password", userHandler.ChangePassword)
	userGroup.Post("/me/avatar", userHandler.UploadAvatar)
	userGroup.Get("/global-rules", userHandler.GetGlobalAgentRules)

	// ── Shop Routes ──
	shopHandler := shop.NewHandler(db)
	rulesHandler := rules.NewHandler(db)

	shopGroup := api.Group("/shops", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	shopGroup.Post("", shopHandler.CreateShop)
	shopGroup.Get("/me", shopHandler.GetMyShop)
	shopGroup.Put("/me", shopHandler.UpdateMyShop)
	shopGroup.Get("/integrations", shopHandler.GetIntegrations)
	shopGroup.Put("/integrations", shopHandler.UpdateIntegrations)
	
	// Shop Custom Rules
	shopGroup.Get("/rules", rulesHandler.ListShopRules)
	shopGroup.Post("/rules", rulesHandler.CreateShopRule)
	shopGroup.Put("/rules/:id", rulesHandler.UpdateShopRule)
	shopGroup.Delete("/rules/:id", rulesHandler.DeleteShopRule)
	shopGroup.Patch("/rules/:id/toggle", rulesHandler.ToggleShopRule)

	// ── Facebook Pages Routes ──
	pagesHandler := pages.NewHandler(db, cfg, jwtManager)
	pagesGroup := api.Group("/pages", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	pagesGroup.Post("/connect", pagesHandler.ConnectPage)
	pagesGroup.Get("", pagesHandler.ListPages)
	pagesGroup.Delete("/:id", pagesHandler.DisconnectPage)
	pagesGroup.Post("/publish", pagesHandler.PublishPost)

	// Auth routes for Pages (don't use the standard AuthMiddleware as they might use query tokens or redirect)
	api.Get("/oauth/pages/facebook", pagesHandler.OAuthLogin)
	api.Get("/oauth/pages/facebook/callback", pagesHandler.OAuthCallback)

	// ── Product Routes ──
	productsHandler := products.NewHandler(db, spacesClient)
	productsGroup := api.Group("/products", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	productsGroup.Post("/upload", productsHandler.UploadImage)
	productsGroup.Post("", productsHandler.CreateProduct)
	productsGroup.Get("", productsHandler.ListProducts)
	productsGroup.Get("/:id", productsHandler.GetProduct)
	productsGroup.Put("/:id", productsHandler.UpdateProduct)
	productsGroup.Delete("/:id", productsHandler.DeleteProduct)

	// ── Order Routes ──
	ordersHandler := orders.NewHandler(db, rmqClient, wsHub, notifSvc)
	ordersGroup := api.Group("/orders", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	ordersGroup.Get("", ordersHandler.ListOrders)
	ordersGroup.Get("/stats", ordersHandler.GetOrderStats)
	ordersGroup.Get("/manual-review", ordersHandler.GetManualReviewOrders)
	ordersGroup.Get("/:id", ordersHandler.GetOrder)
	ordersGroup.Post("", ordersHandler.CreateOrder)
	ordersGroup.Put("/:id", ordersHandler.UpdateOrder)
	ordersGroup.Put("/:id/confirm-payment", ordersHandler.ConfirmPayment)
	ordersGroup.Put("/:id/reject-payment", ordersHandler.RejectPayment)

	// ── Conversation Routes ──
	convHandler := conversations.NewHandler(db, notifSvc)
	convGroup := api.Group("/conversations", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	convGroup.Get("", convHandler.ListConversations)
	convGroup.Get("/stats", convHandler.GetConversationStats)
	convGroup.Get("/:id/messages", convHandler.GetMessages)
	convGroup.Post("/:id/messages", convHandler.SendMessage)
	convGroup.Put("/:id/mode", convHandler.UpdateMode)

	// ── Billing Routes ──
	billingHandler := billing.NewHandler(db, redis, cfg, wsHub)
	billingGroup := api.Group("/billing")

	billingGroup.Get("/plans", billingHandler.GetPlans)
	// Webhook routes (no auth — called by payment gateways)
	billingGroup.Post("/webhook/sslcommerz/success", billingHandler.WebhookSSLCommerzSuccess)
	billingGroup.Post("/webhook/sslcommerz/fail", billingHandler.WebhookSSLCommerzFail)

	billingProtected := billingGroup.Group("", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	billingProtected.Get("/subscription", billingHandler.GetSubscription)
	billingProtected.Post("/subscribe/sslcommerz", billingHandler.SubscribeSSLCommerz)
	billingProtected.Get("/payments", billingHandler.GetPayments)
	billingProtected.Post("/cancel", billingHandler.CancelSubscription)

	// ── Agent Routes ──
	agentGroup := api.Group("/agents", middleware.AuthMiddleware(jwtManager, redis), apiRateLimit)
	agentGroup.Post("/:slug/run", agentHandler.RunAgent)
	agentGroup.Get("/:slug/tasks", agentHandler.GetTasks)
	agentGroup.Get("/:slug/tasks/:taskId", agentHandler.GetTask)
	agentGroup.Delete("/:slug/tasks/:taskId", agentHandler.DeleteTask)

	// ── Content Routes (no auth) ──
	contentRepo := content.NewRepository(db)
	contentSvc := content.NewService(contentRepo, redis, wsHub)
	contentHandler := content.NewContentHandler(db, contentSvc)

	publicContent := api.Group("/content")
	publicContent.Get("/hero", contentHandler.GetHero)
	publicContent.Get("/banner", contentHandler.GetBanner)
	publicContent.Get("/faq", contentHandler.GetFAQ)
	publicContent.Get("/reviews", contentHandler.GetApprovedReviews)

	// User review submission (authenticated users)
	api.Post("/content/reviews", middleware.AuthMiddleware(jwtManager, redis), contentHandler.SubmitReview)

	// ── Admin Routes ──
	adminSvc := admin.NewService(db, redis, wsHub)
	adminHandler := admin.NewHandler(adminSvc)
	adminGroup := api.Group("/admin",
		middleware.AuthMiddleware(jwtManager, redis),
		middleware.RBACMiddleware(models.RoleSuperAdmin, models.RoleAdmin), // Now Super Admin AND Admin
	)

	adminGroup.Get("/overview", adminHandler.GetOverview)
	adminGroup.Get("/users", adminHandler.ListUsers)
	adminGroup.Get("/users/export", adminHandler.ExportUsers)
	adminGroup.Get("/users/:id", adminHandler.GetUser)
	// Some actions might be super_admin only, but the global router lets admin in.
	// RBAC for these specific routes should be enforced explicitly. We wrap them:
	adminGroup.Put("/users/:id/role", middleware.RBACMiddleware(models.RoleSuperAdmin), adminHandler.ChangeUserRole)
	adminGroup.Put("/users/:id/plan", adminHandler.OverrideUserPlan)
	adminGroup.Put("/users/:id/status", adminHandler.ChangeUserStatus)
	adminGroup.Delete("/users/:id", middleware.RBACMiddleware(models.RoleSuperAdmin), adminHandler.DeleteUser)
	
	adminGroup.Get("/users/:id/tasks", adminHandler.GetUserTasks)
	adminGroup.Get("/users/:id/conversations", adminHandler.GetUserConversations)
	adminGroup.Get("/tasks", adminHandler.ListAllTasks)
	adminGroup.Get("/tasks/stats", adminHandler.GetTaskStats)
	adminGroup.Post("/tasks/:id/retry", adminHandler.RetryTask)
	
	adminGroup.Get("/transactions", adminHandler.ListTransactions)
	adminGroup.Get("/transactions/export", adminHandler.ExportTransactions)
	adminGroup.Get("/transactions/:id", adminHandler.GetTransaction)
	
	adminGroup.Get("/plans/:id", adminHandler.GetPlan)
	adminGroup.Put("/plans/:id", adminHandler.UpdatePlan)

	adminGroup.Get("/settings/:key", adminHandler.GetSystemSetting)
	adminGroup.Put("/settings/:key", adminHandler.UpdateSystemSetting)
	adminGroup.Get("/activity", adminHandler.GetRecentActivity)

	// Admin Global Rules
	adminGroup.Get("/rules", rulesHandler.ListGlobalRules)
	adminGroup.Post("/rules", rulesHandler.CreateGlobalRule)
	adminGroup.Put("/rules/:id", rulesHandler.UpdateGlobalRule)
	adminGroup.Delete("/rules/:id", rulesHandler.DeleteGlobalRule)
	adminGroup.Patch("/rules/:id/toggle", rulesHandler.ToggleGlobalRule)

	// ── Admin Content Routes ──
	adminContent := adminGroup.Group("/content")
	adminContent.Get("/hero", contentHandler.AdminGetHero)
	adminContent.Put("/hero", contentHandler.UpdateHero)
	adminContent.Get("/banner", contentHandler.AdminGetBanner)
	adminContent.Put("/banner", contentHandler.UpdateBanner)
	adminContent.Get("/faq", contentHandler.AdminGetFAQ)
	adminContent.Put("/faq", contentHandler.UpdateFAQ)
	
	adminContent.Get("/reviews", contentHandler.AdminListReviews)
	adminContent.Put("/reviews/reorder", contentHandler.ReorderReviews)
	adminContent.Get("/reviews/settings", contentHandler.GetReviewSettings)
	adminContent.Put("/reviews/settings", contentHandler.UpdateReviewSettings)
	adminContent.Put("/reviews/:id", contentHandler.AdminEditReview)
	adminContent.Put("/reviews/:id/approve", contentHandler.ApproveReview)
	adminContent.Put("/reviews/:id/reject", contentHandler.RejectReview)
	adminContent.Delete("/reviews/:id", contentHandler.DeleteReview)


	// ── Bootstrap Admin (only works if no super_admin exists) ──
	api.Post("/admin/bootstrap", middleware.AuthMiddleware(jwtManager, redis), func(c *fiber.Ctx) error {
		var count int64
		db.Model(&models.User{}).Where("role = ?", "super_admin").Count(&count)
		if count > 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Super admin already exists"})
		}
		userID := c.Locals("user_id").(string)
		db.Model(&models.User{}).Where("id = ?", userID).Update("role", "super_admin")
		return c.JSON(fiber.Map{"success": true, "message": "You are now super_admin"})
	})

	// ── WebSocket ──
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		token := conn.Query("token")
		if token == "" {
			conn.Close()
			return
		}

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			conn.Close()
			return
		}

		wsHub.Register(claims.UserID, conn)
		
		// If admin or super_admin, also add to admin room
		if claims.Role == string(models.RoleAdmin) || claims.Role == string(models.RoleSuperAdmin) {
			wsHub.RegisterAdmin(claims.UserID, conn)
		}
		
		defer wsHub.Unregister(claims.UserID)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
}
