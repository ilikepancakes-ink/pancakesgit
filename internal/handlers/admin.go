package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"pancakesgit/internal/models"
	"pancakesgit/internal/services"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	userService  *services.UserService
	repoService  *services.RepositoryService
	badgeService *models.BadgeService
	config       interface{} // Replace with actual config type
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userService *services.UserService, repoService *services.RepositoryService, badgeService *models.BadgeService, config interface{}) *AdminHandler {
	return &AdminHandler{
		userService:  userService,
		repoService:  repoService,
		badgeService: badgeService,
		config:       config,
	}
}

// Dashboard renders the admin dashboard
func (h *AdminHandler) Dashboard(c *gin.Context) {
	user := GetCurrentUser(c)
	
	// Get dashboard statistics
	stats := h.getDashboardStats()
	recentActivity := h.getRecentActivity()
	recentUsers := h.getRecentUsers()
	adminUsers := h.getAdminUsers()
	largestRepos := h.getLargestRepos()
	mostActiveRepos := h.getMostActiveRepos()
	badgeStats := h.getBadgeStats()
	recentBadges := h.getRecentBadges()
	systemAlerts := h.getSystemAlerts()
	
	data := gin.H{
		"Title":           "Admin Dashboard",
		"User":            user,
		"Stats":           stats,
		"RecentActivity":  recentActivity,
		"RecentUsers":     recentUsers,
		"AdminUsers":      adminUsers,
		"LargestRepos":    largestRepos,
		"MostActiveRepos": mostActiveRepos,
		"BadgeStats":      badgeStats,
		"RecentBadges":    recentBadges,
		"SystemAlerts":    systemAlerts,
	}
	
	c.HTML(http.StatusOK, "admin/dashboard.html", data)
}

// UserManagement renders the user management page
func (h *AdminHandler) UserManagement(c *gin.Context) {
	user := GetCurrentUser(c)
	
	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort", "created_desc")
	
	// Calculate offset
	offset := (page - 1) * limit
	
	// Get users with filtering
	users, total := h.getUsersWithFilters(search, role, status, sortBy, limit, offset)
	
	// Get available badges for awarding
	availableBadges := h.getAvailableBadges()
	
	// Calculate pagination
	totalPages := (int(total) + limit - 1) / limit
	hasNextPage := page < totalPages
	hasPrevPage := page > 1
	
	data := gin.H{
		"Title":           "User Management",
		"User":            user,
		"Users":           users,
		"TotalUsers":      total,
		"CurrentPage":     page,
		"TotalPages":      totalPages,
		"HasNextPage":     hasNextPage,
		"HasPrevPage":     hasPrevPage,
		"NextPage":        page + 1,
		"PrevPage":        page - 1,
		"PageStart":       offset + 1,
		"PageEnd":         min(offset+limit, int(total)),
		"AvailableBadges": availableBadges,
	}
	
	c.HTML(http.StatusOK, "admin/users.html", data)
}

// BadgeManagement renders the badge management page
func (h *AdminHandler) BadgeManagement(c *gin.Context) {
	user := GetCurrentUser(c)
	
	// Get badges
	badges, _, _ := h.badgeService.GetBadges("", nil, 100, 0)
	
	// Get badge statistics
	badgeStats, _ := h.badgeService.GetBadgeStats()
	
	// Get recent badges
	recentBadges, _, _ := h.badgeService.GetBadges("", nil, 5, 0)
	
	data := gin.H{
		"Title":        "Badge Management",
		"User":         user,
		"Badges":       badges,
		"Stats":        badgeStats,
		"RecentBadges": recentBadges,
	}
	
	c.HTML(http.StatusOK, "admin/badges.html", data)
}

// MakeAdmin promotes a user to admin
func (h *AdminHandler) MakeAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}
	
	// Update user role
	err = h.userService.UpdateUserRole(uint(userID), "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	currentUser := GetCurrentUser(c)
	h.logAdminAction(currentUser, "make_admin", map[string]interface{}{
		"target_user_id": userID,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveAdmin removes admin privileges from a user
func (h *AdminHandler) RemoveAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}
	
	// Update user role
	err = h.userService.UpdateUserRole(uint(userID), "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	currentUser := GetCurrentUser(c)
	h.logAdminAction(currentUser, "remove_admin", map[string]interface{}{
		"target_user_id": userID,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteUser deletes a user account
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Reason is required"})
		return
	}
	
	// Delete user
	err = h.userService.DeleteUser(uint(userID), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	currentUser := GetCurrentUser(c)
	h.logAdminAction(currentUser, "delete_user", map[string]interface{}{
		"target_user_id": userID,
		"reason":         req.Reason,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SuspendUser suspends a user account
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Reason   string `json:"reason" binding:"required"`
		Duration int    `json:"duration"` // Duration in days, 0 for indefinite
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Reason is required"})
		return
	}
	
	// Suspend user
	err = h.userService.SuspendUser(uint(userID), req.Reason, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	currentUser := GetCurrentUser(c)
	h.logAdminAction(currentUser, "suspend_user", map[string]interface{}{
		"target_user_id": userID,
		"reason":         req.Reason,
		"duration":       req.Duration,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AwardBadgeToUser awards a badge to a user
func (h *AdminHandler) AwardBadgeToUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}
	
	var req struct {
		BadgeID uint   `json:"badge_id" binding:"required"`
		Reason  string `json:"reason"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Badge ID is required"})
		return
	}
	
	// Get current admin user
	currentUser := GetCurrentUser(c)
	adminID := h.getCurrentUserID(currentUser)
	
	// Award badge
	err = h.badgeService.AwardBadge(req.BadgeID, uint(userID), adminID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	h.logAdminAction(currentUser, "award_badge", map[string]interface{}{
		"target_user_id": userID,
		"badge_id":       req.BadgeID,
		"reason":         req.Reason,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CreateBadge creates a new badge
func (h *AdminHandler) CreateBadge(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"required"`
		Category    string `json:"category" binding:"required"`
		Rarity      string `json:"rarity" binding:"required"`
		Color       string `json:"color" binding:"required"`
		IsActive    bool   `json:"is_active"`
		IsPublic    bool   `json:"is_public"`
		Criteria    string `json:"criteria"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Get current admin user
	currentUser := GetCurrentUser(c)
	adminID := h.getCurrentUserID(currentUser)
	
	// Create badge
	badge := &models.Badge{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Rarity:      req.Rarity,
		Color:       req.Color,
		IsActive:    req.IsActive,
		IsPublic:    req.IsPublic,
		Criteria:    req.Criteria,
		CreatedBy:   adminID,
	}
	
	err := h.badgeService.CreateBadge(badge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	// Log the action
	h.logAdminAction(currentUser, "create_badge", map[string]interface{}{
		"badge_id":   badge.ID,
		"badge_name": badge.Name,
	})
	
	c.JSON(http.StatusOK, gin.H{"success": true, "badge": badge})
}

// Helper methods

func (h *AdminHandler) getDashboardStats() gin.H {
	// This would typically fetch real data from services
	return gin.H{
		"TotalUsers":       156,
		"NewUsersToday":    5,
		"TotalRepositories": 1247,
		"NewReposToday":    12,
		"StorageUsed":      "2.4 GB",
		"StorageGrowth":    "+120 MB",
		"ActiveUsers":      89,
	}
}

func (h *AdminHandler) getRecentActivity() []gin.H {
	return []gin.H{
		{
			"Type":        "user_created",
			"Icon":        "user-plus",
			"Description": "New user registered: alice",
			"User":        "alice",
			"TimeAgo":     "2 minutes ago",
			"Action":      "View",
			"ActionURL":   "/admin/users/123",
		},
		{
			"Type":        "repo_created",
			"Icon":        "git-branch",
			"Description": "Repository created: bob/secret-project",
			"User":        "bob",
			"TimeAgo":     "15 minutes ago",
			"Action":      "View",
			"ActionURL":   "/bob/secret-project",
		},
		{
			"Type":        "badge_awarded",
			"Icon":        "award",
			"Description": "Badge 'Early Adopter' awarded to charlie",
			"User":        "admin",
			"TimeAgo":     "1 hour ago",
		},
	}
}

func (h *AdminHandler) getRecentUsers() []gin.H {
	return []gin.H{
		{
			"ID":        "123",
			"Username":  "alice",
			"Email":     "alice@example.com",
			"AvatarURL": "/static/avatars/default.png",
			"CreatedAt": "2 hours ago",
		},
		{
			"ID":        "124",
			"Username":  "bob",
			"Email":     "bob@example.com",
			"AvatarURL": "/static/avatars/default.png",
			"CreatedAt": "5 hours ago",
		},
	}
}

func (h *AdminHandler) getAdminUsers() []gin.H {
	return []gin.H{
		{
			"ID":         "1",
			"Username":   "admin",
			"AvatarURL":  "/static/avatars/default.png",
			"LastActive": "5 minutes ago",
		},
	}
}

func (h *AdminHandler) getLargestRepos() []gin.H {
	return []gin.H{
		{
			"ID":    "1",
			"Owner": "alice",
			"Name":  "big-project",
			"Size":  "1.2 GB",
		},
		{
			"ID":    "2",
			"Owner": "bob",
			"Name":  "data-repo",
			"Size":  "890 MB",
		},
	}
}

func (h *AdminHandler) getMostActiveRepos() []gin.H {
	return []gin.H{
		{
			"Owner":           "alice",
			"Name":            "active-project",
			"CommitsThisWeek": 45,
		},
		{
			"Owner":           "bob",
			"Name":            "busy-repo",
			"CommitsThisWeek": 32,
		},
	}
}

func (h *AdminHandler) getBadgeStats() gin.H {
	stats, _ := h.badgeService.GetBadgeStats()
	return gin.H{
		"TotalBadges":   stats["total_badges"],
		"ActiveBadges":  stats["active_badges"],
		"BadgesAwarded": stats["badges_awarded"],
	}
}

func (h *AdminHandler) getRecentBadges() []gin.H {
	badges, _, _ := h.badgeService.GetBadges("", nil, 3, 0)
	result := make([]gin.H, len(badges))
	
	for i, badge := range badges {
		result[i] = gin.H{
			"ID":          badge.ID,
			"Name":        badge.Name,
			"Description": badge.Description,
			"IconURL":     badge.IconURL,
		}
	}
	
	return result
}

func (h *AdminHandler) getSystemAlerts() []gin.H {
	return []gin.H{
		{
			"Type":      "info",
			"Title":     "System Backup",
			"Message":   "Daily backup completed successfully",
			"TimeAgo":   "2 hours ago",
			"Action":    "View Logs",
			"ActionURL": "/admin/logs",
		},
		{
			"Type":      "warning",
			"Title":     "High Memory Usage",
			"Message":   "Memory usage is at 85%",
			"TimeAgo":   "4 hours ago",
			"Action":    "View System",
			"ActionURL": "/admin/system",
		},
	}
}

func (h *AdminHandler) getUsersWithFilters(search, role, status, sortBy string, limit, offset int) ([]gin.H, int64) {
	// This would typically use the user service to fetch filtered users
	// For now, returning mock data
	users := []gin.H{
		{
			"ID":              "123",
			"Username":        "alice",
			"Email":           "alice@example.com",
			"AvatarURL":       "/static/avatars/default.png",
			"Role":            "user",
			"Status":          "active",
			"RepositoryCount": 5,
			"TotalRepoSize":   "120 MB",
			"LastActive":      "2 hours ago",
			"LastActiveAt":    time.Now().Add(-2 * time.Hour),
			"CreatedAt":       "2024-01-10",
			"Badges": []gin.H{
				{"Name": "Early Adopter", "Color": "#238636"},
			},
		},
	}
	
	return users, 1
}

func (h *AdminHandler) getAvailableBadges() []gin.H {
	badges, _, _ := h.badgeService.GetBadges("", nil, 100, 0)
	result := make([]gin.H, len(badges))
	
	for i, badge := range badges {
		result[i] = gin.H{
			"ID":          badge.ID,
			"Name":        badge.Name,
			"Description": badge.Description,
			"Color":       badge.Color,
			"IconURL":     badge.IconURL,
		}
	}
	
	return result
}

func (h *AdminHandler) logAdminAction(user interface{}, action string, data map[string]interface{}) {
	// This would log admin actions to an audit log
	// Implementation depends on your logging system
}

func (h *AdminHandler) getCurrentUserID(user interface{}) uint {
	// Extract user ID from the user object
	// Implementation depends on your user model
	return 1 // Placeholder
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
