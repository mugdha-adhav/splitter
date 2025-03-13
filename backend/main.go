package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"primaryKey"`
	Name     string    `gorm:"type:varchar(40);unique" json:"name,omitempty"`
	Password string    `gorm:"size:255" json:"password,omitempty"`
	Email    string    `gorm:"type:varchar(40);unique" json:"email"`
	// Add relationships
	OwnedGroups []Group `gorm:"foreignKey:OwnerRefer"`
	Groups      []Group `gorm:"many2many:user_groups;"`
}

type Group struct {
	gorm.Model

	ID         uuid.UUID `gorm:"primaryKey"`
	Name       string    `gorm:"type:varchar(40)" json:"name"`
	OwnerRefer uuid.UUID
	Owner      User   `gorm:"foreignKey:OwnerRefer;constraint:OnDelete:CASCADE;"`
	Members    []User `gorm:"many2many:user_groups;"`
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
	db.AutoMigrate(&User{}, &Group{})

	// Seed User
	{
		hashedPassword, err := makePasswordHash("password")
		if err != nil {
			return nil, fmt.Errorf("failed to seed user: %w", err)
		}

		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&User{
			ID:       uuid.New(),
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
		c.JSON(200, gin.H{
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
			c.JSON(400, gin.H{
				"message": "Invalid request. Password and either email or username are required",
			})
			return
		}

		// Validate that either email or username is provided
		if req.Email == "" && req.Name == "" {
			c.JSON(400, gin.H{
				"message": "Either email or username must be provided",
			})
			return
		}

		if req.Email != "" && req.Name != "" {
			c.JSON(400, gin.H{
				"message": "Please provide either email or username, not both",
			})
			return
		}

		var dbUser User

		// Find user by name or email
		if req.Name != "" {
			if result := db.Where("name = ?", req.Name).First(&dbUser); result.Error != nil {
				c.JSON(404, gin.H{
					"message": "User not found",
				})
				return
			}
		} else {
			if result := db.Where("email = ?", req.Email).First(&dbUser); result.Error != nil {
				c.JSON(404, gin.H{
					"message": "User not found",
				})
				return
			}
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
			c.JSON(401, gin.H{
				"message": "Invalid password",
			})
			return
		}

		c.JSON(200, gin.H{
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
			c.JSON(400, gin.H{
				"message": "Invalid request. Name, email, and password are required",
			})
			return
		}

		var dbUser User

		// Check if email already exists
		result := db.Where("email = ?", req.Email).First(&dbUser)
		if result.Error == nil {
			c.JSON(400, gin.H{
				"message": "Email already exists",
			})
			return
		}

		// Check if name already exists
		result = db.Where("name = ?", req.Name).First(&dbUser)
		if result.Error == nil {
			c.JSON(400, gin.H{
				"message": "Username already exists",
			})
			return
		}

		// Hash password
		pass, err := makePasswordHash(req.Password)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error while processing registration",
			})
			return
		}

		// Create new user
		user := User{
			ID:       uuid.New(),
			Name:     req.Name,
			Email:    req.Email,
			Password: pass,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{
				"message": "Failed to create user",
			})
			return
		}

		c.JSON(201, gin.H{
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
			Name    string    `json:"name" binding:"required"`
			OwnerID uuid.UUID `json:"owner_id" binding:"required"`
		}

		var req CreateGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"message": "Invalid request. Group name and owner ID are required",
			})
			return
		}

		// Verify owner exists
		var owner User
		if result := db.First(&owner, "id = ?", req.OwnerID); result.Error != nil {
			c.JSON(404, gin.H{
				"message": "Owner not found",
			})
			return
		}

		// Create group
		group := Group{
			ID:         uuid.New(),
			Name:       req.Name,
			OwnerRefer: req.OwnerID,
			Members:    []User{owner}, // Add owner as a member too
		}

		if err := db.Create(&group).Error; err != nil {
			c.JSON(500, gin.H{
				"message": "Failed to create group",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Group created successfully",
			"group": gin.H{
				"id":       group.ID,
				"name":     group.Name,
				"owner_id": group.OwnerRefer,
			},
		})
	})

	r.Run()
}
