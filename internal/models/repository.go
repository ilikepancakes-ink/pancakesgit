package models

import (
	"time"
)

// Repository represents a Git repository
type Repository struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;index"`
	Description string    `json:"description"`
	OwnerID     uint      `json:"owner_id" gorm:"not null;index"`
	IsPrivate   bool      `json:"is_private" gorm:"default:false"`
	IsFork      bool      `json:"is_fork" gorm:"default:false"`
	ParentID    *uint     `json:"parent_id" gorm:"index"`
	DefaultBranch string  `json:"default_branch" gorm:"default:'main'"`
	Size        int64     `json:"size" gorm:"default:0"`
	Language    string    `json:"language"`
	Topics      string    `json:"topics"` // JSON array as string
	Website     string    `json:"website"`
	IsArchived  bool      `json:"is_archived" gorm:"default:false"`
	IsTemplate  bool      `json:"is_template" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PushedAt    *time.Time `json:"pushed_at"`

	// Relationships
	Owner        User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Parent       *Repository    `json:"parent" gorm:"foreignKey:ParentID"`
	Forks        []Repository   `json:"forks" gorm:"foreignKey:ParentID"`
	Stars        []Star         `json:"stars" gorm:"foreignKey:RepositoryID"`
	Watches      []Watch        `json:"watches" gorm:"foreignKey:RepositoryID"`
	Issues       []Issue        `json:"issues" gorm:"foreignKey:RepositoryID"`
	PullRequests []PullRequest  `json:"pull_requests" gorm:"foreignKey:RepositoryID"`
	Releases     []Release      `json:"releases" gorm:"foreignKey:RepositoryID"`
	Webhooks     []Webhook      `json:"webhooks" gorm:"foreignKey:RepositoryID"`
	Collaborations []Collaboration `json:"collaborations" gorm:"foreignKey:RepositoryID"`

	// Computed fields
	StarCount    int `json:"star_count" gorm:"-"`
	WatchCount   int `json:"watch_count" gorm:"-"`
	ForkCount    int `json:"fork_count" gorm:"-"`
	IssueCount   int `json:"issue_count" gorm:"-"`
	PullRequestCount int `json:"pull_request_count" gorm:"-"`
}

// Star represents a repository star
type Star struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_user_repository_star,unique"`
}

// Watch represents a repository watch
type Watch struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_user_repository_watch,unique"`
}

// Fork represents a repository fork relationship
type Fork struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	ParentID     uint      `json:"parent_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`
	Parent     Repository `json:"parent" gorm:"foreignKey:ParentID"`
}

// Collaboration represents repository collaboration
type Collaboration struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Permission   string    `json:"permission" gorm:"not null"` // read, write, admin
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User       User       `json:"user" gorm:"foreignKey:UserID"`
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_user_repository_collab,unique"`
}

// SSHKey represents a user's SSH key
type SSHKey struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Title       string    `json:"title" gorm:"not null"`
	Key         string    `json:"key" gorm:"not null;type:text"`
	Fingerprint string    `json:"fingerprint" gorm:"not null;uniqueIndex"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	LastUsed    *time.Time `json:"last_used"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// AccessToken represents an API access token
type AccessToken struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Token       string    `json:"token" gorm:"not null;uniqueIndex"`
	Scopes      string    `json:"scopes"` // JSON array as string
	LastUsed    *time.Time `json:"last_used"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Webhook represents a repository webhook
type Webhook struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	URL          string    `json:"url" gorm:"not null"`
	Secret       string    `json:"secret"`
	Events       string    `json:"events"` // JSON array as string
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`
}
