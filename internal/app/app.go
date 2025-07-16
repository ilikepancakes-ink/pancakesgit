package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pancakesgit/pancakesgit/internal/config"
	"github.com/pancakesgit/pancakesgit/internal/database"
	"github.com/pancakesgit/pancakesgit/internal/encryption"
	"github.com/pancakesgit/pancakesgit/internal/git"
	"github.com/pancakesgit/pancakesgit/internal/handlers"
	"github.com/pancakesgit/pancakesgit/internal/middleware"
	"github.com/pancakesgit/pancakesgit/internal/models"
	"github.com/pancakesgit/pancakesgit/internal/services"
	"github.com/pancakesgit/pancakesgit/internal/storage"
)

// App represents the main application
type App struct {
	config     *config.Config
	db         *database.DB
	encryption *encryption.Service
	storage    *storage.Service
	git        *git.Service
	userSvc    *services.UserService
	repoSvc    *services.RepositoryService
	authSvc    *services.AuthService
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
	db, err := database.New(a.config.Database, a.encryption)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = db

	// Initialize storage
	storageSvc, err := storage.New(a.config.Storage, a.encryption)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	a.storage = storageSvc

	// Initialize Git service
	gitSvc, err := git.New(a.config.Git, a.storage, a.encryption)
	if err != nil {
		return fmt.Errorf("failed to initialize git service: %w", err)
	}
	a.git = gitSvc

	// Initialize business services
	a.userSvc = services.NewUserService(a.db, a.encryption)
	a.repoSvc = services.NewRepositoryService(a.db, a.storage, a.git, a.encryption)
	a.authSvc = services.NewAuthService(a.db, a.config.Security, a.encryption)

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
	a.router.Use(middleware.Security(a.config.Security))
	a.router.Use(middleware.Privacy(a.config.Privacy))

	// Static files
	a.router.Static("/static", "./static")
	a.router.LoadHTMLGlob("templates/*")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(a.authSvc, a.userSvc, a.config)
	userHandler := handlers.NewUserHandler(a.userSvc, a.config)
	repoHandler := handlers.NewRepositoryHandler(a.repoSvc, a.userSvc, a.config)
	gitHandler := handlers.NewGitHandler(a.git, a.repoSvc, a.authSvc, a.config)

	// Public routes
	public := a.router.Group("/")
	{
		public.GET("/", handlers.HomeHandler)
		public.GET("/health", handlers.HealthHandler)
		public.GET("/api/healthz", handlers.HealthzHandler)
		
		// Auth routes
		public.GET("/login", authHandler.LoginPage)
		public.POST("/login", authHandler.Login)
		public.GET("/register", authHandler.RegisterPage)
		public.POST("/register", authHandler.Register)
		public.POST("/logout", authHandler.Logout)
		public.GET("/forgot-password", authHandler.ForgotPasswordPage)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.GET("/reset-password", authHandler.ResetPasswordPage)
		public.POST("/reset-password", authHandler.ResetPassword)
	}

	// Git HTTP routes (with basic auth)
	gitHTTP := a.router.Group("/")
	gitHTTP.Use(middleware.GitAuth(a.authSvc))
	{
		gitHTTP.GET("/:username/:repo/info/refs", gitHandler.InfoRefs)
		gitHTTP.POST("/:username/:repo/git-upload-pack", gitHandler.UploadPack)
		gitHTTP.POST("/:username/:repo/git-receive-pack", gitHandler.ReceivePack)
	}

	// Protected routes (require session auth)
	protected := a.router.Group("/")
	protected.Use(middleware.Auth(a.authSvc))
	{
		// User routes
		protected.GET("/settings", userHandler.Settings)
		protected.POST("/settings", userHandler.UpdateSettings)
		protected.GET("/settings/security", userHandler.SecuritySettings)
		protected.POST("/settings/security", userHandler.UpdateSecuritySettings)
		protected.GET("/settings/privacy", userHandler.PrivacySettings)
		protected.POST("/settings/privacy", userHandler.UpdatePrivacySettings)

		// Repository routes
		protected.GET("/new", repoHandler.NewRepoPage)
		protected.POST("/new", repoHandler.CreateRepo)
		protected.GET("/:username/:repo", repoHandler.ViewRepo)
		protected.GET("/:username/:repo/settings", repoHandler.RepoSettings)
		protected.POST("/:username/:repo/settings", repoHandler.UpdateRepoSettings)
		protected.DELETE("/:username/:repo", repoHandler.DeleteRepo)
		
		// Repository content routes
		protected.GET("/:username/:repo/tree/*path", repoHandler.ViewTree)
		protected.GET("/:username/:repo/blob/*path", repoHandler.ViewBlob)
		protected.GET("/:username/:repo/commits", repoHandler.ViewCommits)
		protected.GET("/:username/:repo/commit/:sha", repoHandler.ViewCommit)
		protected.GET("/:username/:repo/branches", repoHandler.ViewBranches)
		protected.GET("/:username/:repo/tags", repoHandler.ViewTags)
		
		// Repository management
		protected.GET("/:username/:repo/clone", repoHandler.CloneInfo)
		protected.POST("/:username/:repo/archive", repoHandler.CreateArchive)
	}

	// Admin routes
	admin := a.router.Group("/admin")
	admin.Use(middleware.Auth(a.authSvc))
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/", handlers.AdminDashboard)
		admin.GET("/users", userHandler.AdminUserList)
		admin.GET("/users/:id", userHandler.AdminUserDetail)
		admin.POST("/users/:id/disable", userHandler.AdminDisableUser)
		admin.POST("/users/:id/enable", userHandler.AdminEnableUser)
		admin.GET("/repositories", repoHandler.AdminRepoList)
		admin.GET("/system", handlers.AdminSystemInfo)
		admin.GET("/logs", handlers.AdminLogs)
	}

	// API routes
	api := a.router.Group("/api/v1")
	api.Use(middleware.APIAuth(a.authSvc))
	{
		api.GET("/user", userHandler.GetCurrentUser)
		api.GET("/user/repos", repoHandler.GetUserRepos)
		api.POST("/user/repos", repoHandler.CreateRepoAPI)
		api.GET("/repos/:username/:repo", repoHandler.GetRepoAPI)
		api.PATCH("/repos/:username/:repo", repoHandler.UpdateRepoAPI)
		api.DELETE("/repos/:username/:repo", repoHandler.DeleteRepoAPI)
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
	// Start Git SSH server if enabled
	if a.config.Git.EnableSSH {
		go func() {
			if err := a.git.StartSSHServer(); err != nil {
				fmt.Printf("SSH server error: %v\n", err)
			}
		}()
	}

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
	return a.db.AutoMigrate(
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
