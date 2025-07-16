package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pancakesgit/internal/config"
	"pancakesgit/internal/database"
	"pancakesgit/internal/encryption"
	"pancakesgit/internal/git"
	"pancakesgit/internal/handlers"
	"pancakesgit/internal/middleware"
	"pancakesgit/internal/models"
	"pancakesgit/internal/services"
	"pancakesgit/internal/storage"

	"github.com/gin-gonic/gin"
)

// App represents the main application
type App struct {
	config     *config.Config
	db         *database.Service
	encryption *encryption.Service
	storage    *storage.Service
	git        *git.Service
	userSvc    *services.UserService
	repoSvc    *services.RepositoryService
	server     *http.Server
	router     *gin.Engine
}

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
	}

	// Initialize components
	if err := app.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Setup routes
	app.setupRoutes()

	return app, nil
}

func (a *App) initializeComponents() error {
	// Initialize encryption service
	encSvc, err := encryption.New(a.config.Encryption)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}
	a.encryption = encSvc

	// Initialize database
	db, err := database.New(a.config.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = db

	// Initialize storage
	a.storage = storage.New(a.config.Storage)

	// Initialize Git service
	a.git = git.New(a.config.Git)

	// Initialize business services
	a.userSvc = services.NewUserService(a.db.DB)
	a.repoSvc = services.NewRepositoryService(a.db.DB)

	return nil
}

func (a *App) setupRoutes() {
	// Set Gin mode
	if !a.config.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	a.router = gin.New()

	// Global middleware
	a.router.Use(gin.Logger())
	a.router.Use(gin.Recovery())
	a.router.Use(middleware.CORS())

	// Static files
	a.router.Static("/static", "./static")
	a.router.LoadHTMLGlob("templates/*")

	// Public routes
	public := a.router.Group("/")
	{
		public.GET("/", handlers.HomeHandler)
		public.GET("/health", handlers.HealthHandler)
		public.GET("/healthz", handlers.HealthzHandler)
	}

	// Create HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", a.config.Server.Host, a.config.Server.Port),
		Handler:      a.router,
		ReadTimeout:  time.Duration(a.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.config.Server.WriteTimeout) * time.Second,
	}
}

// Start starts the application server
func (a *App) Start() error {

	// Start HTTP server
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Close database connection
	if a.db != nil {
		a.db.Close()
	}

	return nil
}

// Migrate runs database migrations
func (a *App) Migrate() error {
	return a.db.DB.AutoMigrate(
		&models.User{},
		&models.Repository{},
		&models.SSHKey{},
		&models.AccessToken{},
		&models.Webhook{},
		&models.Issue{},
		&models.PullRequest{},
		&models.Comment{},
		&models.Label{},
		&models.Milestone{},
		&models.Release{},
		&models.Star{},
		&models.Watch{},
		&models.Fork{},
		&models.Collaboration{},
		&models.Organization{},
		&models.Team{},
		&models.AuditLog{},
	)
}

// CreateUser creates a new user
func (a *App) CreateUser(username, email, password string, isAdmin bool) error {
	return a.userSvc.CreateUser(username, email, password, isAdmin)
}
