package controllers

import (
	"chat-app/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	UserStore map[string]*models.User
}

func NewAuthController() *AuthController {
	return &AuthController{
		UserStore: make(map[string]*models.User),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user *models.User
	for _, u := range ac.UserStore {
		if u.Username == req.Username || u.Email == req.Email {
			user = u
			break
		}
	}

	if user == nil {
		user = models.NewUser(req.Username, req.Email)
		ac.UserStore[user.ID] = user
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "token": user.ID})
}
