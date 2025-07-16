package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pancakesgit/pancakesgit/internal/app"
	"github.com/pancakesgit/pancakesgit/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "pancakesgit",
		Short: "PancakesGit - Privacy-focused Git service",
		Long: `PancakesGit is a privacy-focused, encrypted Git service that prioritizes
user privacy and data security while providing a familiar Git hosting experience.`,
		Run: runServer,
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("PancakesGit %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", date)
		},
	}

	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: runMigrations,
	}

	var createUserCmd = &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user",
		Run:   createUser,
	}

	rootCmd.AddCommand(versionCmd, migrateCmd, createUserCmd)

	// Add flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringP("port", "p", "3000", "server port")
	rootCmd.PersistentFlags().StringP("host", "h", "0.0.0.0", "server host")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(cmd)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize and start the application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Printf("Starting PancakesGit server on %s:%s", cfg.Server.Host, cfg.Server.Port)
	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runMigrations(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(cmd)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	log.Println("Running database migrations...")
	if err := application.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations completed successfully")
}

func createUser(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(cmd)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Check if this is automatic admin creation from environment
	if os.Getenv("ADMIN_CREATE") == "true" {
		username := os.Getenv("ADMIN_USERNAME")
		email := os.Getenv("ADMIN_EMAIL")
		password := os.Getenv("ADMIN_PASSWORD")

		if username == "" || email == "" || password == "" {
			log.Fatalf("Admin credentials not provided in environment variables")
		}

		if err := application.CreateUser(username, email, password, true); err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		log.Printf("Admin user '%s' created successfully", username)
		return
	}

	// Interactive user creation
	fmt.Print("Username: ")
	var username string
	fmt.Scanln(&username)

	fmt.Print("Email: ")
	var email string
	fmt.Scanln(&email)

	fmt.Print("Password: ")
	var password string
	fmt.Scanln(&password)

	fmt.Print("Is admin? (y/N): ")
	var isAdminInput string
	fmt.Scanln(&isAdminInput)
	isAdmin := isAdminInput == "y" || isAdminInput == "Y"

	if err := application.CreateUser(username, email, password, isAdmin); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	log.Printf("User '%s' created successfully", username)
}
