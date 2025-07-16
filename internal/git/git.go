package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"pancakesgit/internal/config"
)

// Service provides Git functionality
type Service struct {
	config config.GitConfig
	repoPath string
}

// New creates a new Git service
func New(cfg config.GitConfig) *Service {
	return &Service{
		config: cfg,
		repoPath: "/data/repositories", // Default path
	}
}

// Repository represents a Git repository
type Repository struct {
	Name        string
	Path        string
	Description string
	IsPrivate   bool
}

// InitRepository initializes a new Git repository
func (s *Service) InitRepository(name, description string, isPrivate bool) (*Repository, error) {
	repoPath := filepath.Join(s.repoPath, name+".git")
	
	// Create directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Initialize bare repository
	_, err := git.PlainInit(repoPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	repo := &Repository{
		Name:        name,
		Path:        repoPath,
		Description: description,
		IsPrivate:   isPrivate,
	}

	return repo, nil
}

// CloneRepository clones a repository
func (s *Service) CloneRepository(url, name string) (*Repository, error) {
	repoPath := filepath.Join(s.repoPath, name)
	
	_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	repo := &Repository{
		Name: name,
		Path: repoPath,
	}

	return repo, nil
}

// GetRepository opens an existing repository
func (s *Service) GetRepository(name string) (*Repository, error) {
	repoPath := filepath.Join(s.repoPath, name+".git")
	
	_, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	repo := &Repository{
		Name: name,
		Path: repoPath,
	}

	return repo, nil
}

// GetCommits returns the commit history
func (s *Service) GetCommits(repoPath string, limit int) ([]*object.Commit, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commits []*object.Commit
	count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached")
		}
		commits = append(commits, c)
		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// DeleteRepository removes a repository
func (s *Service) DeleteRepository(name string) error {
	repoPath := filepath.Join(s.repoPath, name+".git")
	return os.RemoveAll(repoPath)
}
