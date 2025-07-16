package models

import (
	"time"

	"gorm.io/gorm"
)

// Badge represents a user badge/achievement
type Badge struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex"`
	Description string    `json:"description" gorm:"not null"`
	Category    string    `json:"category" gorm:"not null;index"` // achievement, role, special, community
	Rarity      string    `json:"rarity" gorm:"not null"`         // common, rare, epic, legendary
	Color       string    `json:"color" gorm:"not null"`          // Hex color code
	IconURL     string    `json:"icon_url"`                       // URL to badge icon
	IconData    []byte    `json:"-" gorm:"type:bytea"`           // Encrypted icon data
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsPublic    bool      `json:"is_public" gorm:"default:true"` // Visible on profiles
	Criteria    string    `json:"criteria"`                      // Auto-award criteria
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Creator    User         `json:"creator" gorm:"foreignKey:CreatedBy"`
	UserBadges []UserBadge  `json:"user_badges" gorm:"foreignKey:BadgeID"`
	Awards     []BadgeAward `json:"awards" gorm:"foreignKey:BadgeID"`

	// Computed fields
	AwardedCount int `json:"awarded_count" gorm:"-"`
}

// UserBadge represents a badge awarded to a user
type UserBadge struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	BadgeID   uint      `json:"badge_id" gorm:"not null;index"`
	AwardedBy uint      `json:"awarded_by"` // Admin who awarded it (0 for auto-awarded)
	Reason    string    `json:"reason"`     // Reason for awarding
	IsVisible bool      `json:"is_visible" gorm:"default:true"`
	AwardedAt time.Time `json:"awarded_at"`

	// Relationships
	User     User  `json:"user" gorm:"foreignKey:UserID"`
	Badge    Badge `json:"badge" gorm:"foreignKey:BadgeID"`
	Awarder  User  `json:"awarder" gorm:"foreignKey:AwardedBy"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_user_badge,unique"`
}

// BadgeAward represents the history of badge awards
type BadgeAward struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BadgeID   uint      `json:"badge_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	AwardedBy uint      `json:"awarded_by"`
	Reason    string    `json:"reason"`
	Action    string    `json:"action"` // awarded, revoked
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Badge   Badge `json:"badge" gorm:"foreignKey:BadgeID"`
	User    User  `json:"user" gorm:"foreignKey:UserID"`
	Awarder User  `json:"awarder" gorm:"foreignKey:AwardedBy"`
}

// BadgeCategory represents badge categories
type BadgeCategory struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null;uniqueIndex"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
}

// Badge methods

// BeforeCreate sets default values
func (b *Badge) BeforeCreate(tx *gorm.DB) error {
	if b.Color == "" {
		b.Color = "#238636" // Default green
	}
	return nil
}

// GetAwardedCount returns the number of times this badge has been awarded
func (b *Badge) GetAwardedCount(db *gorm.DB) int {
	var count int64
	db.Model(&UserBadge{}).Where("badge_id = ?", b.ID).Count(&count)
	return int(count)
}

// IsAwardedToUser checks if this badge is awarded to a specific user
func (b *Badge) IsAwardedToUser(db *gorm.DB, userID uint) bool {
	var count int64
	db.Model(&UserBadge{}).Where("badge_id = ? AND user_id = ?", b.ID, userID).Count(&count)
	return count > 0
}

// CanBeAwarded checks if this badge can be awarded (is active)
func (b *Badge) CanBeAwarded() bool {
	return b.IsActive
}

// GetRarityWeight returns a numeric weight for rarity sorting
func (b *Badge) GetRarityWeight() int {
	switch b.Rarity {
	case "legendary":
		return 4
	case "epic":
		return 3
	case "rare":
		return 2
	case "common":
		return 1
	default:
		return 0
	}
}

// UserBadge methods

// BeforeCreate sets the awarded timestamp
func (ub *UserBadge) BeforeCreate(tx *gorm.DB) error {
	if ub.AwardedAt.IsZero() {
		ub.AwardedAt = time.Now()
	}
	return nil
}

// Badge service methods (would be in a service layer)

// BadgeService provides badge-related operations
type BadgeService struct {
	db *gorm.DB
}

// NewBadgeService creates a new badge service
func NewBadgeService(db *gorm.DB) *BadgeService {
	return &BadgeService{db: db}
}

// CreateBadge creates a new badge
func (s *BadgeService) CreateBadge(badge *Badge) error {
	return s.db.Create(badge).Error
}

// GetBadge retrieves a badge by ID
func (s *BadgeService) GetBadge(id uint) (*Badge, error) {
	var badge Badge
	err := s.db.Preload("Creator").First(&badge, id).Error
	if err != nil {
		return nil, err
	}
	badge.AwardedCount = badge.GetAwardedCount(s.db)
	return &badge, nil
}

// GetBadges retrieves badges with filtering and pagination
func (s *BadgeService) GetBadges(category string, isActive *bool, limit, offset int) ([]Badge, int64, error) {
	query := s.db.Model(&Badge{}).Preload("Creator")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	var total int64
	query.Count(&total)

	var badges []Badge
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&badges).Error
	if err != nil {
		return nil, 0, err
	}

	// Load awarded counts
	for i := range badges {
		badges[i].AwardedCount = badges[i].GetAwardedCount(s.db)
	}

	return badges, total, nil
}

// AwardBadge awards a badge to a user
func (s *BadgeService) AwardBadge(badgeID, userID, awardedBy uint, reason string) error {
	// Check if badge exists and is active
	var badge Badge
	if err := s.db.First(&badge, badgeID).Error; err != nil {
		return err
	}
	if !badge.CanBeAwarded() {
		return gorm.ErrRecordNotFound // Badge is not active
	}

	// Check if user already has this badge
	if badge.IsAwardedToUser(s.db, userID) {
		return gorm.ErrDuplicatedKey // User already has this badge
	}

	// Create user badge
	userBadge := &UserBadge{
		UserID:    userID,
		BadgeID:   badgeID,
		AwardedBy: awardedBy,
		Reason:    reason,
	}

	// Create award history
	award := &BadgeAward{
		BadgeID:   badgeID,
		UserID:    userID,
		AwardedBy: awardedBy,
		Reason:    reason,
		Action:    "awarded",
	}

	// Use transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(userBadge).Error; err != nil {
			return err
		}
		return tx.Create(award).Error
	})
}

// RevokeBadge removes a badge from a user
func (s *BadgeService) RevokeBadge(badgeID, userID, revokedBy uint, reason string) error {
	// Create revocation history
	award := &BadgeAward{
		BadgeID:   badgeID,
		UserID:    userID,
		AwardedBy: revokedBy,
		Reason:    reason,
		Action:    "revoked",
	}

	// Use transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete user badge
		if err := tx.Where("badge_id = ? AND user_id = ?", badgeID, userID).Delete(&UserBadge{}).Error; err != nil {
			return err
		}
		// Create history record
		return tx.Create(award).Error
	})
}

// GetUserBadges retrieves all badges for a user
func (s *BadgeService) GetUserBadges(userID uint, visibleOnly bool) ([]UserBadge, error) {
	query := s.db.Preload("Badge").Where("user_id = ?", userID)
	if visibleOnly {
		query = query.Where("is_visible = ?", true).Joins("JOIN badges ON badges.id = user_badges.badge_id").Where("badges.is_public = ?", true)
	}

	var userBadges []UserBadge
	err := query.Order("awarded_at DESC").Find(&userBadges).Error
	return userBadges, err
}

// GetBadgeStats returns badge statistics
func (s *BadgeService) GetBadgeStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total badges
	var totalBadges int64
	s.db.Model(&Badge{}).Count(&totalBadges)
	stats["total_badges"] = totalBadges

	// Active badges
	var activeBadges int64
	s.db.Model(&Badge{}).Where("is_active = ?", true).Count(&activeBadges)
	stats["active_badges"] = activeBadges

	// Total awards
	var totalAwards int64
	s.db.Model(&UserBadge{}).Count(&totalAwards)
	stats["badges_awarded"] = totalAwards

	// Most popular badge
	var popularBadge struct {
		BadgeID uint
		Count   int64
		Name    string
	}
	s.db.Model(&UserBadge{}).
		Select("badge_id, COUNT(*) as count, badges.name").
		Joins("JOIN badges ON badges.id = user_badges.badge_id").
		Group("badge_id, badges.name").
		Order("count DESC").
		First(&popularBadge)
	stats["popular_badge"] = popularBadge.Name

	return stats, nil
}

// UpdateBadge updates an existing badge
func (s *BadgeService) UpdateBadge(badge *Badge) error {
	return s.db.Save(badge).Error
}

// DeleteBadge deletes a badge (soft delete)
func (s *BadgeService) DeleteBadge(id uint) error {
	return s.db.Delete(&Badge{}, id).Error
}

// GetBadgeCategories returns all badge categories
func (s *BadgeService) GetBadgeCategories() ([]BadgeCategory, error) {
	var categories []BadgeCategory
	err := s.db.Order("sort_order ASC, name ASC").Find(&categories).Error
	return categories, err
}
