package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Username         string         `json:"username" gorm:"uniqueIndex;not null"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash     string         `json:"-" gorm:"not null"`
	DisplayName      string         `json:"display_name"`
	Bio              string         `json:"bio"`
	Location         string         `json:"location"`
	Website          string         `json:"website"`
	Company          string         `json:"company"`
	AvatarURL        string         `json:"avatar_url"`
	IsActive         bool           `json:"is_active" gorm:"default:true"`
	IsAdmin          bool           `json:"is_admin" gorm:"default:false"`
	Role             string         `json:"role" gorm:"default:user"`     // user, admin, moderator
	Status           string         `json:"status" gorm:"default:active"` // active, suspended, banned
	EmailVerified    bool           `json:"email_verified" gorm:"default:false"`
	TwoFactorEnabled bool           `json:"two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret  string         `json:"-"`
	LastLoginAt      *time.Time     `json:"last_login_at"`
	SuspendedUntil   *time.Time     `json:"suspended_until"`
	SuspensionReason string         `json:"suspension_reason"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Repositories   []Repository    `json:"repositories" gorm:"foreignKey:OwnerID"`
	SSHKeys        []SSHKey        `json:"ssh_keys" gorm:"foreignKey:UserID"`
	AccessTokens   []AccessToken   `json:"access_tokens" gorm:"foreignKey:UserID"`
	Stars          []Star          `json:"stars" gorm:"foreignKey:UserID"`
	Watches        []Watch         `json:"watches" gorm:"foreignKey:UserID"`
	Collaborations []Collaboration `json:"collaborations" gorm:"foreignKey:UserID"`
	Issues         []Issue         `json:"issues" gorm:"foreignKey:AuthorID"`
	AssignedIssues []Issue         `json:"assigned_issues" gorm:"foreignKey:AssigneeID"`
	PullRequests   []PullRequest   `json:"pull_requests" gorm:"foreignKey:AuthorID"`
	Comments       []Comment       `json:"comments" gorm:"foreignKey:AuthorID"`
	AuditLogs      []AuditLog      `json:"audit_logs" gorm:"foreignKey:UserID"`

	// Badge relationships
	UserBadges    []UserBadge  `json:"user_badges" gorm:"foreignKey:UserID"`
	CreatedBadges []Badge      `json:"created_badges" gorm:"foreignKey:CreatedBy"`
	BadgeAwards   []BadgeAward `json:"badge_awards" gorm:"foreignKey:UserID"`

	// Computed fields
	Badges          []Badge `json:"badges" gorm:"-"`
	RepositoryCount int     `json:"repository_count" gorm:"-"`
	FollowerCount   int     `json:"follower_count" gorm:"-"`
	FollowingCount  int     `json:"following_count" gorm:"-"`
	StarCount       int     `json:"star_count" gorm:"-"`
	StarredCount    int     `json:"starred_count" gorm:"-"`
	IsOnline        bool    `json:"is_online" gorm:"-"`
	JoinedDate      string  `json:"joined_date" gorm:"-"`
}

// User methods

// BeforeCreate sets default values
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	if u.Status == "" {
		u.Status = "active"
	}
	if u.AvatarURL == "" {
		u.AvatarURL = "/static/avatars/default.png"
	}
	return nil
}

// IsActiveUser checks if the user is active and not suspended
func (u *User) IsActiveUser() bool {
	if !u.IsActive || u.Status == "banned" {
		return false
	}

	if u.Status == "suspended" && u.SuspendedUntil != nil {
		return time.Now().After(*u.SuspendedUntil)
	}

	return u.Status == "active"
}

// CanAccessRepository checks if user can access a repository
func (u *User) CanAccessRepository(repo *Repository) bool {
	if !u.IsActiveUser() {
		return false
	}

	// Owner can always access
	if repo.OwnerID == u.ID {
		return true
	}

	// Public repositories are accessible to all active users
	if !repo.IsPrivate {
		return true
	}

	// Admins can access all repositories
	if u.IsAdmin {
		return true
	}

	// Check collaborations (would need to query database)
	return false
}

// GetDisplayName returns the display name or username
func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Username
}

// LoadBadges loads user badges from database
func (u *User) LoadBadges(db *gorm.DB) error {
	var userBadges []UserBadge
	err := db.Preload("Badge").Where("user_id = ? AND is_visible = ?", u.ID, true).
		Joins("JOIN badges ON badges.id = user_badges.badge_id").
		Where("badges.is_public = ?", true).
		Order("user_badges.awarded_at DESC").
		Find(&userBadges).Error

	if err != nil {
		return err
	}

	u.Badges = make([]Badge, len(userBadges))
	for i, ub := range userBadges {
		u.Badges[i] = ub.Badge
	}

	return nil
}

// LoadCounts loads computed counts for the user
func (u *User) LoadCounts(db *gorm.DB) error {
	var count int64

	// Repository count
	db.Model(&Repository{}).Where("owner_id = ?", u.ID).Count(&count)
	u.RepositoryCount = int(count)

	// Star count (stars received)
	db.Table("stars").
		Joins("JOIN repositories ON repositories.id = stars.repository_id").
		Where("repositories.owner_id = ?", u.ID).
		Count(&count)
	u.StarCount = int(count)

	// Starred count (stars given)
	db.Model(&Star{}).Where("user_id = ?", u.ID).Count(&count)
	u.StarredCount = int(count)

	return nil
}

// SetJoinedDate sets the formatted joined date
func (u *User) SetJoinedDate() {
	u.JoinedDate = u.CreatedAt.Format("January 2006")
}
