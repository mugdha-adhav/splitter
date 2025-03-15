package handlers

import (
	"fmt"
	"net/http"

	"example.com/backend/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ExpenseCreate(c *gin.Context) {
	// Get user ID from JWT claims
	claims := jwt.ExtractClaims(c)
	currentUserID := uint(claims[identityKey].(float64))

	type CreateExpenseRequest struct {
		Amount  float64 `json:"amount" binding:"required" form:"amount"`
		GroupID uint    `json:"group_id" binding:"required" form:"group_id"`
		UserIDs []uint  `json:"user_ids" form:"user_ids"`
	}

	var req CreateExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request. Amount and group ID are required",
			"error":   err.Error(),
		})
		return
	}

	// Verify group exists
	var group models.Group
	if result := h.db.Preload("Members").First(&group, "id = ?", req.GroupID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Group not found",
		})
		return
	}

	// Verify creator is a member of the group
	creatorIsMember := false
	for _, member := range group.Members {
		if member.ID == currentUserID {
			creatorIsMember = true
			break
		}
	}

	if !creatorIsMember {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "You must be a member of the group to create an expense",
		})
		return
	}

	// Deduplicate UserIDs
	userIDMap := make(map[uint]bool)
	var uniqueUserIDs []uint
	for _, id := range req.UserIDs {
		if _, exists := userIDMap[id]; !exists {
			userIDMap[id] = true
			uniqueUserIDs = append(uniqueUserIDs, id)
		}
	}

	var users []models.User
	// Only validate users if UserIDs are provided
	if len(uniqueUserIDs) > 0 {
		if result := h.db.Find(&users, uniqueUserIDs); result.Error != nil || len(users) != len(uniqueUserIDs) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "One or more users not found",
			})
			return
		}

		// Verify all users are members of the group
		for _, user := range users {
			isMember := false
			for _, member := range group.Members {
				if member.ID == user.ID {
					isMember = true
					break
				}
			}
			if !isMember {
				c.JSON(http.StatusForbidden, gin.H{
					"message": fmt.Sprintf("User %d is not a member of the group", user.ID),
				})
				return
			}
		}
	}

	// Create expense
	expense := models.Expense{
		Amount:      req.Amount,
		GroupID:     req.GroupID,
		CreatedByID: currentUserID,
		Users:       users,
	}

	if err := h.db.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create expense",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Expense created successfully",
		"expense": gin.H{
			"id":         expense.ID,
			"amount":     expense.Amount,
			"group_id":   expense.GroupID,
			"created_by": expense.CreatedByID,
			"user_ids":   uniqueUserIDs,
		},
	})
}
