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
)

type User struct {
	gorm.Model

	ID       uuid.UUID `gorm:"primaryKey"`
	Name     string    `gorm:"type:varchar(40);unique" json:"name,omitempty" form:"name,omitempty"`
	Password string    `gorm:"size:255" json:"password,omitempty" form:"password,omitempty"`
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

		if err := db.FirstOrCreate(&User{
			ID:       uuid.New(),
			Name:     "defaultUser",
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
			result := db.Where("name = ?", user.Name).First(&dbUser)
			if result.Error != nil {
				c.JSON(401, gin.H{
					"message": "User not found",
				})
				return
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
	r.Run()
}
