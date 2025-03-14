package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"primaryKey"`
	Name     string    `gorm:"type:varchar(40);unique" json:"name,omitempty" form:"name,omitempty"`
	Password string    `gorm:"size:255" json:"password" form:"password,omitempty"`
	Email    string    `gorm:"type:varchar(40);unique" json:"email" form:"email,omitempty"`
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
	db.AutoMigrate(&User{})

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
		var user User
		if c.ShouldBind(&user) == nil {
			var dbUser User

			if user.Name != "" && user.Email == "" {
				result := db.Where("name = ?", user.Name).First(&dbUser)
				if result.Error != nil {
					c.JSON(401, gin.H{
						"message": "User not found",
					})
					return
				}
			}

			if user.Email != "" && user.Name == "" {
				result := db.Where("email = ?", user.Email).First(&dbUser)
				if result.Error != nil {
					c.JSON(401, gin.H{
						"message": "Email not found",
					})
					return
				}
			}

			err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
			if err != nil {
				c.JSON(401, gin.H{
					"message": "Invalid password",
				})
				return
			}

			c.JSON(200, gin.H{
				"message": "Login successful",
			})
		}
	})

	r.POST("/register", func(c *gin.Context) {
		var user User
		if c.ShouldBind(&user) == nil {
			var dbUser User

			if user.Password == "" || user.Email == "" || user.Name == "" {
				c.JSON(401, gin.H{
					"message": "Please enter all details",
				})
				return
			}

			result := db.Where("email = ?", user.Email).First(&dbUser)
			if result.Error == nil {
				c.JSON(401, gin.H{
					"message": "Email already exists",
				})
				return
			}

			result = db.Where("name = ?", user.Name).First(&dbUser)
			if result.Error == nil {
				c.JSON(401, gin.H{
					"message": "Name already exists",
				})
				return
			}

			pass, err := makePasswordHash(user.Password)
			if err != nil {
				c.JSON(401, gin.H{
					"message": "Failed to register user",
				})
				return
			}

			if err := db.Create(&User{
				ID:       uuid.New(),
				Name:     user.Name,
				Email:    user.Email,
				Password: pass,
			}).Error; err != nil {
				c.JSON(401, gin.H{
					"message": "Failed to register user",
				})
				return
			}

			c.JSON(200, gin.H{
				"message": "Registration successful",
			})
		}
	})
	r.Run()
}
