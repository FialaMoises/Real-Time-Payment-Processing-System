package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/api/handlers"
	"github.com/yourusername/real-time-payments/internal/account"
	"github.com/yourusername/real-time-payments/internal/auth"
	"github.com/yourusername/real-time-payments/internal/ledger"
	"github.com/yourusername/real-time-payments/internal/transaction"
	"github.com/yourusername/real-time-payments/internal/user"
	"github.com/yourusername/real-time-payments/pkg/middleware"
)

type Router struct {
	engine     *gin.Engine
	db         *sql.DB
	jwtService auth.JWTService
}

func NewRouter(db *sql.DB, jwtService auth.JWTService) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	return &Router{
		engine:     engine,
		db:         db,
		jwtService: jwtService,
	}
}

func (r *Router) Setup() *gin.Engine {
	// Global middlewares
	r.engine.Use(middleware.RecoveryMiddleware())
	r.engine.Use(middleware.LoggingMiddleware())
	r.engine.Use(middleware.CORSMiddleware())

	// Initialize repositories
	userRepo := user.NewRepository(r.db)
	accountRepo := account.NewRepository(r.db)
	txRepo := transaction.NewRepository(r.db)
	ledgerRepo := ledger.NewRepository(r.db)

	// Initialize services
	userService := user.NewService(userRepo, r.jwtService, 3600)
	accountService := account.NewService(accountRepo)
	txService := transaction.NewService(r.db, txRepo, accountRepo, ledgerRepo)
	ledgerService := ledger.NewService(ledgerRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, accountService)
	accountHandler := handlers.NewAccountHandler(accountService)
	txHandler := handlers.NewTransactionHandler(txService)
	ledgerHandler := handlers.NewLedgerHandler(ledgerService)
	healthHandler := handlers.NewHealthHandler(r.db)

	// Public routes
	public := r.engine.Group("/api/v1")
	{
		// Health
		public.GET("/health", healthHandler.Health)

		// Auth
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}

	// Protected routes (require authentication)
	protected := r.engine.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(r.jwtService))
	{
		// Accounts
		protected.GET("/accounts/me", accountHandler.GetMyAccount)
		protected.GET("/accounts/:id", accountHandler.GetAccountByID)
		protected.GET("/accounts/:id/balance", accountHandler.GetBalance)

		// Transactions
		protected.POST("/transactions/deposit", txHandler.Deposit)
		protected.POST("/transactions/withdrawal", txHandler.Withdrawal)
		protected.POST("/transactions/transfer", txHandler.Transfer)
		protected.GET("/transactions/:id", txHandler.GetTransactionByID)
		protected.GET("/transactions", txHandler.GetTransactions)

		// Ledger
		protected.GET("/ledger/:account_id", ledgerHandler.GetLedgerByAccountID)
	}

	return r.engine
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
