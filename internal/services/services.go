package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"pancakesgit/internal/models"
)

// UserService handles user-related operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(username, email, password string, isAdmin bool) error {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return fmt.Errorf("user with username or email already exists")
	}

	// Create new user
	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: password, // In real implementation, this should be hashed
		IsActive:     true,
		IsAdmin:      isAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.db.Create(&user).Error
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.db.Save(user).Error
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id uint) error {
	return s.db.Delete(&models.User{}, id).Error
}

// RepositoryService handles repository-related operations
type RepositoryService struct {
	db *gorm.DB
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(db *gorm.DB) *RepositoryService {
	return &RepositoryService{db: db}
}

// CreateRepository creates a new repository
func (s *RepositoryService) CreateRepository(name, description string, ownerID uint, isPrivate bool) (*models.Repository, error) {
	repo := models.Repository{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		IsPrivate:   isPrivate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&repo).Error; err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepository retrieves a repository by ID
func (s *RepositoryService) GetRepository(id uint) (*models.Repository, error) {
	var repo models.Repository
	if err := s.db.Preload("Owner").First(&repo, id).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetRepositoryByName retrieves a repository by owner and name
func (s *RepositoryService) GetRepositoryByName(ownerID uint, name string) (*models.Repository, error) {
	var repo models.Repository
	if err := s.db.Where("owner_id = ? AND name = ?", ownerID, name).Preload("Owner").First(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

// UpdateRepository updates repository information
func (s *RepositoryService) UpdateRepository(repo *models.Repository) error {
	repo.UpdatedAt = time.Now()
	return s.db.Save(repo).Error
}

// DeleteRepository deletes a repository
func (s *RepositoryService) DeleteRepository(id uint) error {
	return s.db.Delete(&models.Repository{}, id).Error
}

// ListUserRepositories lists repositories for a user
func (s *RepositoryService) ListUserRepositories(userID uint) ([]models.Repository, error) {
	var repos []models.Repository
	if err := s.db.Where("owner_id = ?", userID).Find(&repos).Error; err != nil {
		return nil, err
	}
	return repos, nil
}
