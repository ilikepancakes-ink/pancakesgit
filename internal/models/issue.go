package models

import (
	"time"
)

// Issue represents a repository issue
type Issue struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Number       int       `json:"number" gorm:"not null;index"`
	Title        string    `json:"title" gorm:"not null"`
	Body         string    `json:"body" gorm:"type:text"`
	AuthorID     uint      `json:"author_id" gorm:"not null;index"`
	AssigneeID   *uint     `json:"assignee_id" gorm:"index"`
	MilestoneID  *uint     `json:"milestone_id" gorm:"index"`
	State        string    `json:"state" gorm:"not null;default:'open'"` // open, closed
	IsPullRequest bool     `json:"is_pull_request" gorm:"default:false"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`
	Author     User       `json:"author" gorm:"foreignKey:AuthorID"`
	Assignee   *User      `json:"assignee" gorm:"foreignKey:AssigneeID"`
	Milestone  *Milestone `json:"milestone" gorm:"foreignKey:MilestoneID"`
	Comments   []Comment  `json:"comments" gorm:"foreignKey:IssueID"`
	Labels     []Label    `json:"labels" gorm:"many2many:issue_labels;"`

	// Computed fields
	CommentCount int `json:"comment_count" gorm:"-"`
}

// PullRequest represents a pull request
type PullRequest struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Number       int       `json:"number" gorm:"not null;index"`
	Title        string    `json:"title" gorm:"not null"`
	Body         string    `json:"body" gorm:"type:text"`
	AuthorID     uint      `json:"author_id" gorm:"not null;index"`
	AssigneeID   *uint     `json:"assignee_id" gorm:"index"`
	MilestoneID  *uint     `json:"milestone_id" gorm:"index"`
	State        string    `json:"state" gorm:"not null;default:'open'"` // open, closed, merged
	HeadBranch   string    `json:"head_branch" gorm:"not null"`
	BaseBranch   string    `json:"base_branch" gorm:"not null"`
	HeadSHA      string    `json:"head_sha"`
	BaseSHA      string    `json:"base_sha"`
	MergedAt     *time.Time `json:"merged_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`
	Author     User       `json:"author" gorm:"foreignKey:AuthorID"`
	Assignee   *User      `json:"assignee" gorm:"foreignKey:AssigneeID"`
	Milestone  *Milestone `json:"milestone" gorm:"foreignKey:MilestoneID"`
	Comments   []Comment  `json:"comments" gorm:"foreignKey:PullRequestID"`
	Labels     []Label    `json:"labels" gorm:"many2many:pull_request_labels;"`

	// Computed fields
	CommentCount int `json:"comment_count" gorm:"-"`
}

// Comment represents a comment on an issue or pull request
type Comment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	IssueID       *uint     `json:"issue_id" gorm:"index"`
	PullRequestID *uint     `json:"pull_request_id" gorm:"index"`
	AuthorID      uint      `json:"author_id" gorm:"not null;index"`
	Body          string    `json:"body" gorm:"not null;type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Issue       *Issue       `json:"issue" gorm:"foreignKey:IssueID"`
	PullRequest *PullRequest `json:"pull_request" gorm:"foreignKey:PullRequestID"`
	Author      User         `json:"author" gorm:"foreignKey:AuthorID"`
}

// Label represents a label for issues and pull requests
type Label struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Name         string    `json:"name" gorm:"not null"`
	Color        string    `json:"color" gorm:"not null"` // Hex color code
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository   Repository    `json:"repository" gorm:"foreignKey:RepositoryID"`
	Issues       []Issue       `json:"issues" gorm:"many2many:issue_labels;"`
	PullRequests []PullRequest `json:"pull_requests" gorm:"many2many:pull_request_labels;"`

	// Unique constraint
	_ struct{} `gorm:"uniqueIndex:idx_repository_label_name,unique"`
}

// Milestone represents a milestone for issues and pull requests
type Milestone struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Title        string    `json:"title" gorm:"not null"`
	Description  string    `json:"description"`
	State        string    `json:"state" gorm:"not null;default:'open'"` // open, closed
	DueDate      *time.Time `json:"due_date"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository   Repository    `json:"repository" gorm:"foreignKey:RepositoryID"`
	Issues       []Issue       `json:"issues" gorm:"foreignKey:MilestoneID"`
	PullRequests []PullRequest `json:"pull_requests" gorm:"foreignKey:MilestoneID"`

	// Computed fields
	OpenIssues   int `json:"open_issues" gorm:"-"`
	ClosedIssues int `json:"closed_issues" gorm:"-"`
	Progress     int `json:"progress" gorm:"-"` // Percentage
}

// Release represents a repository release
type Release struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	TagName      string    `json:"tag_name" gorm:"not null"`
	Name         string    `json:"name"`
	Body         string    `json:"body" gorm:"type:text"`
	AuthorID     uint      `json:"author_id" gorm:"not null;index"`
	IsDraft      bool      `json:"is_draft" gorm:"default:false"`
	IsPrerelease bool      `json:"is_prerelease" gorm:"default:false"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository Repository `json:"repository" gorm:"foreignKey:RepositoryID"`
	Author     User       `json:"author" gorm:"foreignKey:AuthorID"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"user_id" gorm:"index"`
	Action    string    `json:"action" gorm:"not null;index"`
	Resource  string    `json:"resource" gorm:"not null;index"`
	ResourceID *uint    `json:"resource_id" gorm:"index"`
	Details   string    `json:"details" gorm:"type:text"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User *User `json:"user" gorm:"foreignKey:UserID"`
}
