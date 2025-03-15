package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model

	Name     string `gorm:"type:varchar(40);unique;not null" json:"name,omitempty" form:"name,omitempty"`
	Password string `gorm:"size:255;not null" json:"password,omitempty" form:"password,omitempty"`
	Email    string `gorm:"type:varchar(40);unique;not null" json:"email" form:"email,omitempty"`
	// Add relationships
	OwnedGroups     []Group   `gorm:"foreignKey:OwnerRefer"`
	Groups          []Group   `gorm:"many2many:user_groups;"`
	CreatedExpenses []Expense `gorm:"foreignKey:CreatedByID"`
	SharedExpenses  []Expense `gorm:"many2many:user_expenses;"`
}

type Group struct {
	gorm.Model

	Name       string    `gorm:"type:varchar(40);not null" json:"name" form:"name"`
	OwnerRefer uint      `gorm:"not null" json:"owner_id" form:"owner_id"`
	Owner      User      `gorm:"foreignKey:OwnerRefer;constraint:OnDelete:CASCADE;"`
	Members    []User    `gorm:"many2many:user_groups;"`
	Expenses   []Expense `gorm:"foreignKey:GroupID"`
}

type Expense struct {
	gorm.Model

	Amount      float64 `gorm:"not null" json:"amount" form:"amount"`
	GroupID     uint    `gorm:"not null" json:"group_id" form:"group_id"`
	Group       Group   `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedByID uint    `gorm:"not null" json:"created_by_id" form:"created_by_id"`
	CreatedBy   User    `gorm:"constraint:OnDelete:CASCADE;"`
	Users       []User  `gorm:"many2many:user_expenses;"`
}

// makePasswordHash generates a password hash
func makePasswordHash(password string) (string, error) {
	data, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return string(data), nil
}

func dbInit() (*gorm.DB, error) {
	// Connect to DB
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Group{}, &Expense{})

	// Seed User
	{
		hashedPassword, err := makePasswordHash("password")
		if err != nil {
			return nil, fmt.Errorf("failed to seed user: %w", err)
		}

		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&User{
			Name:     "defaultUser",
			Email:    "default@example.com",
			Password: hashedPassword,
		}).Error; err != nil {
			return nil, fmt.Errorf("failed to seed user: %w", err)
		}
	}

	return db, nil
}

func main() {
	db, err := dbInit()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/login", func(c *gin.Context) {
		type LoginRequest struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password" binding:"required"`
		}

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request. Password and either email or username are required",
			})
			return
		}

		// Validate that either email or username is provided
		if req.Email == "" && req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Either email or username must be provided",
			})
			return
		}

		if req.Email != "" && req.Name != "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Please provide either email or username, not both",
			})
			return
		}

		var dbUser User

		// Find user by name or email
		if req.Name != "" {
			if result := db.Where("name = ?", req.Name).First(&dbUser); result.Error != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "User not found",
				})
				return
			}
		} else {
			if result := db.Where("email = ?", req.Email).First(&dbUser); result.Error != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "User not found",
				})
				return
			}
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid password",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user": gin.H{
				"id":    dbUser.ID,
				"name":  dbUser.Name,
				"email": dbUser.Email,
			},
		})
	})

	r.POST("/register", func(c *gin.Context) {
		type RegisterRequest struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request. Name, email, and password are required",
			})
			return
		}

		var dbUser User

		// Check if email already exists
		result := db.Where("email = ?", req.Email).First(&dbUser)
		if result.Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Email already exists",
			})
			return
		}

		// Check if name already exists
		result = db.Where("name = ?", req.Name).First(&dbUser)
		if result.Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Username already exists",
			})
			return
		}

		// Hash password
		pass, err := makePasswordHash(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error while processing registration",
			})
			return
		}

		// Create new user
		user := User{
			Name:     req.Name,
			Email:    req.Email,
			Password: pass,
		}

		if err := db.Create(&user).Error; err != nil {
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
	})

	r.POST("/group", func(c *gin.Context) {
		type CreateGroupRequest struct {
			Name      string `json:"name" binding:"required"`
			OwnerID   uint   `json:"owner_id" binding:"required"`
			MemberIDs []uint `json:"member_ids"` // Optional member IDs
		}

		var req CreateGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request. Group name and owner ID are required",
			})
			return
		}

		// Verify owner exists
		var owner User
		if result := db.First(&owner, "id = ?", req.OwnerID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Owner not found",
			})
			return
		}

		// Initialize members with owner
		members := []User{owner}

		// Add additional members if provided
		if len(req.MemberIDs) > 0 {
			var additionalMembers []User
			if result := db.Find(&additionalMembers, req.MemberIDs); result.Error != nil || len(additionalMembers) != len(req.MemberIDs) {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "One or more member IDs are invalid",
				})
				return
			}

			// Add members, ensuring we don't add the owner twice
			for _, member := range additionalMembers {
				if member.ID != owner.ID {
					members = append(members, member)
				}
			}
		}

		// Create group
		group := Group{
			Name:       req.Name,
			OwnerRefer: req.OwnerID,
			Members:    members,
		}

		if err := db.Create(&group).Error; err != nil {
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
	})

	r.POST("/expense", func(c *gin.Context) {
		type CreateExpenseRequest struct {
			Amount      float64 `json:"amount" binding:"required"`
			GroupID     uint    `json:"group_id" binding:"required"`
			CreatedByID uint    `json:"created_by_id" binding:"required"`
			UserIDs     []uint  `json:"user_ids"`
		}

		var req CreateExpenseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request. Amount, group ID, and creator ID are required",
				"error":   err.Error(),
			})
			return
		}

		// Verify group exists
		var group Group
		if result := db.Preload("Members").First(&group, "id = ?", req.GroupID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Group not found",
			})
			return
		}

		// Verify creator exists and is a member of the group
		var creator User
		if result := db.First(&creator, "id = ?", req.CreatedByID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Expense creator not found",
			})
			return
		}

		creatorIsMember := false
		for _, member := range group.Members {
			if member.ID == creator.ID {
				creatorIsMember = true
				break
			}
		}

		if !creatorIsMember {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Expense creator must be a member of the group",
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

		var users []User
		// Only validate users if UserIDs are provided
		if len(uniqueUserIDs) > 0 {
			if result := db.Find(&users, uniqueUserIDs); result.Error != nil || len(users) != len(uniqueUserIDs) {
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
		expense := Expense{
			Amount:      req.Amount,
			GroupID:     req.GroupID,
			CreatedByID: req.CreatedByID,
			Users:       users,
		}

		if err := db.Create(&expense).Error; err != nil {
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
				"user_ids":   req.UserIDs,
			},
		})
	})

	r.Run()
}
