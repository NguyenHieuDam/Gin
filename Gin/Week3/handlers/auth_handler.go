package handlers

import (
	"net/http"

	"WEEK3/models"
	"WEEK3/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu không hợp lệ",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Đăng ký thành công",
		"user":    user,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dữ liệu không hợp lệ",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Trong thực tế, bạn nên tạo JWT token ở đây
	// For now, we'll just return the user info
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"user":    user,
		"token":   "fake-jwt-token", // Thay bằng JWT token thực
	})
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Trong thực tế, lấy user ID từ JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Không có quyền truy cập",
		})
		return
	}

	userID := userIDStr.(uint)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile updates user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Implementation for updating user profile
	c.JSON(http.StatusOK, gin.H{
		"message": "Chức năng cập nhật profile chưa được implement",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Trong thực tế, invalidate JWT token
	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng xuất thành công",
	})
}
