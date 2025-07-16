package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HomeHandler handles the home page
func HomeHandler(c *gin.Context) {
	user := GetCurrentUser(c)
	
	data := gin.H{
		"Title": "Home",
		"User":  user,
		"Year":  time.Now().Year(),
	}
	
	if user != nil {
		// User is logged in, show dashboard
		dashboardData := getDashboardData(c, user)
		for k, v := range dashboardData {
			data[k] = v
		}
	}
	
	c.HTML(http.StatusOK, "home.html", data)
}

// getDashboardData retrieves data for the dashboard
func getDashboardData(c *gin.Context, user interface{}) gin.H {
	// This would typically fetch real data from services
	// For now, we'll return mock data
	
	return gin.H{
		"Stats": gin.H{
			"RepositoryCount":   12,
			"StarCount":        45,
			"CommitCount":      128,
			"CollaboratorCount": 8,
		},
		"RecentActivity": []gin.H{
			{
				"Icon":          "git-commit",
				"Description":   "Pushed 3 commits to main branch",
				"Repository":    "my-awesome-project",
				"RepositoryURL": "/user/my-awesome-project",
				"TimeAgo":       "2 hours ago",
			},
			{
				"Icon":          "star",
				"Description":   "Starred privacy-tools/encryption-lib",
				"Repository":    "privacy-tools/encryption-lib",
				"RepositoryURL": "/privacy-tools/encryption-lib",
				"TimeAgo":       "5 hours ago",
			},
			{
				"Icon":          "git-branch",
				"Description":   "Created new repository",
				"Repository":    "secret-project",
				"RepositoryURL": "/user/secret-project",
				"TimeAgo":       "1 day ago",
			},
		},
		"RecentRepositories": []gin.H{
			{
				"Name":          "my-awesome-project",
				"Description":   "A privacy-focused web application with end-to-end encryption",
				"IsPrivate":     true,
				"IsArchived":    false,
				"Language":      "Go",
				"LanguageColor": "#00ADD8",
				"StarCount":     15,
				"UpdatedAt":     "2 hours ago",
			},
			{
				"Name":          "crypto-utils",
				"Description":   "Cryptographic utilities for secure applications",
				"IsPrivate":     false,
				"IsArchived":    false,
				"Language":      "Rust",
				"LanguageColor": "#dea584",
				"StarCount":     8,
				"UpdatedAt":     "1 day ago",
			},
			{
				"Name":          "privacy-docs",
				"Description":   "Documentation for privacy-preserving software development",
				"IsPrivate":     false,
				"IsArchived":    false,
				"Language":      "Markdown",
				"LanguageColor": "#083fa1",
				"StarCount":     22,
				"UpdatedAt":     "3 days ago",
			},
		},
		"StarredRepositories": []gin.H{
			{
				"Owner":         "privacy-tools",
				"Name":          "encryption-lib",
				"Description":   "High-performance encryption library with zero-knowledge architecture",
				"IsPrivate":     false,
				"Language":      "C++",
				"LanguageColor": "#f34b7d",
				"StarCount":     1205,
				"StarredAt":     "5 hours ago",
			},
			{
				"Owner":         "secure-dev",
				"Name":          "audit-toolkit",
				"Description":   "Security audit tools for modern applications",
				"IsPrivate":     false,
				"Language":      "Python",
				"LanguageColor": "#3572A5",
				"StarCount":     892,
				"StarredAt":     "2 days ago",
			},
		},
		"PullRequests": []gin.H{
			{
				"Number":     42,
				"Title":      "Add end-to-end encryption for user data",
				"Repository": "team/secure-app",
				"Author":     "alice",
				"State":      "open",
				"TimeAgo":    "3 hours ago",
			},
			{
				"Number":     38,
				"Title":      "Fix memory leak in crypto module",
				"Repository": "crypto-utils",
				"Author":     "bob",
				"State":      "review",
				"TimeAgo":    "1 day ago",
			},
		},
		"Issues": []gin.H{
			{
				"Number":     156,
				"Title":      "Implement zero-knowledge proof verification",
				"Repository": "privacy-tools/zkp-lib",
				"Author":     "charlie",
				"State":      "open",
				"TimeAgo":    "4 hours ago",
				"Labels": []gin.H{
					{"Name": "enhancement", "Color": "#a2eeef"},
					{"Name": "crypto", "Color": "#d73a4a"},
				},
			},
			{
				"Number":     89,
				"Title":      "Security vulnerability in authentication flow",
				"Repository": "my-awesome-project",
				"Author":     "security-bot",
				"State":      "open",
				"TimeAgo":    "6 hours ago",
				"Labels": []gin.H{
					{"Name": "security", "Color": "#d73a4a"},
					{"Name": "high-priority", "Color": "#b60205"},
				},
			},
		},
	}
}

// HealthHandler handles health checks
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "pancakesgit",
	})
}

// HealthzHandler handles Kubernetes-style health checks
func HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// AdminDashboard handles the admin dashboard
func AdminDashboard(c *gin.Context) {
	user := GetCurrentUser(c)
	
	data := gin.H{
		"Title": "Admin Dashboard",
		"User":  user,
		"Stats": gin.H{
			"TotalUsers":        156,
			"TotalRepositories": 1247,
			"ActiveUsers":       89,
			"StorageUsed":       "2.4 GB",
			"SystemLoad":        "12%",
			"Uptime":           "15 days",
		},
		"RecentUsers": []gin.H{
			{
				"Username":  "alice",
				"Email":     "alice@example.com",
				"CreatedAt": "2 hours ago",
				"IsActive":  true,
			},
			{
				"Username":  "bob",
				"Email":     "bob@example.com",
				"CreatedAt": "5 hours ago",
				"IsActive":  true,
			},
		},
		"SystemAlerts": []gin.H{
			{
				"Type":    "info",
				"Message": "System backup completed successfully",
				"Time":    "1 hour ago",
			},
			{
				"Type":    "warning",
				"Message": "High memory usage detected",
				"Time":    "3 hours ago",
			},
		},
	}
	
	c.HTML(http.StatusOK, "admin/dashboard.html", data)
}

// AdminSystemInfo handles system information page
func AdminSystemInfo(c *gin.Context) {
	user := GetCurrentUser(c)
	
	data := gin.H{
		"Title": "System Information",
		"User":  user,
		"System": gin.H{
			"Version":     "1.0.0",
			"GoVersion":   "1.21.0",
			"Platform":    "linux/amd64",
			"StartTime":   "2024-01-15 10:30:00",
			"Uptime":      "15 days, 4 hours",
			"CPUUsage":    "12%",
			"MemoryUsage": "45%",
			"DiskUsage":   "23%",
		},
		"Database": gin.H{
			"Type":        "PostgreSQL",
			"Version":     "15.4",
			"Size":        "1.2 GB",
			"Connections": 15,
		},
		"Storage": gin.H{
			"Type":           "Local",
			"TotalSpace":     "100 GB",
			"UsedSpace":      "23 GB",
			"AvailableSpace": "77 GB",
			"Repositories":   1247,
		},
	}
	
	c.HTML(http.StatusOK, "admin/system.html", data)
}

// AdminLogs handles the logs viewer
func AdminLogs(c *gin.Context) {
	user := GetCurrentUser(c)
	
	data := gin.H{
		"Title": "System Logs",
		"User":  user,
		"Logs": []gin.H{
			{
				"Timestamp": "2024-01-15 14:30:15",
				"Level":     "INFO",
				"Message":   "User alice logged in successfully",
				"Source":    "auth",
			},
			{
				"Timestamp": "2024-01-15 14:29:45",
				"Level":     "INFO",
				"Message":   "Repository created: alice/new-project",
				"Source":    "repository",
			},
			{
				"Timestamp": "2024-01-15 14:28:30",
				"Level":     "WARN",
				"Message":   "Failed login attempt for user bob",
				"Source":    "auth",
			},
		},
	}
	
	c.HTML(http.StatusOK, "admin/logs.html", data)
}

// GetCurrentUser is a helper function to get the current user from context
func GetCurrentUser(c *gin.Context) interface{} {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user
}
