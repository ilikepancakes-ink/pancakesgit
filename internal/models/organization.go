package models

import (
	"time"
)

// Organization represents an organization
type Organization struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	Location    string    `json:"location"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Teams        []Team               `json:"teams" gorm:"foreignKey:OrganizationID"`
	Members      []OrganizationMember `json:"members" gorm:"foreignKey:OrganizationID"`
	Repositories []Repository         `json:"repositories" gorm:"foreignKey:OwnerID"`

	// Computed fields
	MemberCount     int `json:"member_count" gorm:"-"`
	RepositoryCount int `json:"repository_count" gorm:"-"`
	TeamCount       int `json:"team_count" gorm:"-"`
}

// Team represents a team within an organization
type Team struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	OrganizationID uint      `json:"organization_id" gorm:"not null;index"`
	Name           string    `json:"name" gorm:"not null"`
	Description    string    `json:"description"`
	Permission     string    `json:"permission" gorm:"not null;default:'read'"` // read, write, admin
	IsVisible      bool      `json:"is_visible" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Organization Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
	Members      []TeamMember `json:"members" gorm:"foreignKey:TeamID"`

	// Computed fields
	MemberCount int `json:"member_count" gorm:"-"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_organization_team_name,unique"`
}

// OrganizationMember represents a member of an organization
type OrganizationMember struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	OrganizationID uint      `json:"organization_id" gorm:"not null;index"`
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	Role           string    `json:"role" gorm:"not null;default:'member'"` // member, admin, owner
	IsPublic       bool      `json:"is_public" gorm:"default:false"`
	JoinedAt       time.Time `json:"joined_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Organization Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
	User         User         `json:"user" gorm:"foreignKey:UserID"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_organization_user_member,unique"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TeamID    uint      `json:"team_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Role      string    `json:"role" gorm:"not null;default:'member'"` // member, maintainer
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Team Team `json:"team" gorm:"foreignKey:TeamID"`
	User User `json:"user" gorm:"foreignKey:UserID"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_team_user_member,unique"`
}
