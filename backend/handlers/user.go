package handlers

import (
	"net/http"

	"example.com/backend/database"
	"example.com/backend/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterUserHandler(c *gin.Context) {
	type RegisterRequest struct {
		Name     string `json:"name" binding:"required" form:"name"`
		Email    string `json:"email" binding:"required" form:"email"`
		Password string `json:"password" binding:"required" form:"password"`
	}

	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request. Name, email, and password are required",
		})
		return
	}

	var dbUser models.User

	// Check if email already exists
	result := h.db.Where("email = ?", req.Email).First(&dbUser)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Email already exists",
		})
		return
	}

	// Check if name already exists
	result = h.db.Where("name = ?", req.Name).First(&dbUser)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username already exists",
		})
		return
	}

	// Hash password
	pass, err := database.MakePasswordHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error while processing registration",
		})
		return
	}

	// Create new user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: pass,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}
