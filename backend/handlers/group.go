package handlers

import (
	"net/http"

	"example.com/backend/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GroupCreate(c *gin.Context) {
	// Get user ID from JWT claims
	claims := jwt.ExtractClaims(c)
	currentUserID := uint(claims[identityKey].(float64))

	type CreateGroupRequest struct {
		Name      string `json:"name" binding:"required" form:"name"`
		MemberIDs []uint `json:"member_ids" form:"member_ids"` // Optional member IDs
	}

	var req CreateGroupRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request. Group name is required",
		})
		return
	}

	// Initialize members with owner
	members := []models.User{{Model: gorm.Model{ID: currentUserID}}}

	// Add additional members if provided
	if len(req.MemberIDs) > 0 {
		var additionalMembers []models.User
		if result := h.db.Find(&additionalMembers, req.MemberIDs); result.Error != nil || len(additionalMembers) != len(req.MemberIDs) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "One or more member IDs are invalid",
			})
			return
		}

		// Add members, ensuring we don't add the owner twice
		for _, member := range additionalMembers {
			if member.ID != currentUserID {
				members = append(members, member)
			}
		}
	}

	// Create group
	group := models.Group{
		Name:       req.Name,
		OwnerRefer: currentUserID,
		Members:    members,
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create group",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Group created successfully",
		"group": gin.H{
			"id":           group.ID,
			"name":         group.Name,
			"owner_id":     group.OwnerRefer,
			"member_count": len(members),
		},
	})
}
