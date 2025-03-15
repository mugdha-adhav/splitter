package handlers

import (
	"log"
	"net/http"
	"time"

	"example.com/backend/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	goJWT "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Handler holds any dependencies
type Handler struct {
	db *gorm.DB
}

// New creates a new handler instance
func New(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var (
	identityKey = "id"
)

func Register(r *gin.Engine, handle *jwt.GinJWTMiddleware, h *Handler) {
	// Public APIs (no authentication required)
	api := r.Group("/api/v1")
	api.POST("/register", h.RegisterUserHandler)
	api.POST("/login", handle.LoginHandler)

	// Authenticated APIs
	auth := api.Group("/auth").Use(handle.MiddlewareFunc())

	auth.GET("/refresh_token", handle.RefreshHandler)
	auth.GET("/hello", helloHandler)
	auth.POST("/group", h.GroupCreate)
	auth.POST("/expense", h.ExpenseCreate)

	// Handle 404s
	r.NoRoute(handleNoRoute())
}

// InitParams initializes jwt middleware
func InitParams() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     payloadFunc(),
		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(),
		Authorizator:    authorizator(),
		Unauthorized:    unauthorized(),
		TokenLookup:     "header:Authorization",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":        http.StatusOK,
				"token":       token,
				"expire":      expire.Format(time.RFC3339),
				"message":     "Login successful",
				"redirect_to": "/auth/hello",
			})
		},
	}
}

func payloadFunc() func(data interface{}) goJWT.MapClaims {
	return func(data interface{}) goJWT.MapClaims {
		if v, ok := data.(*models.User); ok {
			return goJWT.MapClaims{
				identityKey: v.ID, // Store ID instead of Name
			}
		}
		return goJWT.MapClaims{}
	}
}

func identityHandler() func(c *gin.Context) interface{} {
	return func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		id := uint(claims[identityKey].(float64)) // Convert float64 to uint
		return &models.User{
			Model: gorm.Model{ID: id},
		}
	}
}

func authenticator() func(c *gin.Context) (interface{}, error) {
	return func(c *gin.Context) (interface{}, error) {
		var loginVals login
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}

		// Get database connection from context
		db := c.MustGet("db").(*gorm.DB)

		var dbUser models.User
		// Find user by username
		if result := db.Where("name = ?", loginVals.Username).First(&dbUser); result.Error != nil {
			return nil, jwt.ErrFailedAuthentication
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginVals.Password)); err != nil {
			return nil, jwt.ErrFailedAuthentication
		}

		return &dbUser, nil
	}
}

func authorizator() func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, c *gin.Context) bool {
		// Allow all authenticated users
		if _, ok := data.(*models.User); ok {
			return true
		}
		return false
	}
}

func unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func handleNoRoute() func(c *gin.Context) {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	}
}

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id := uint(claims[identityKey].(float64))

	// Fetch full user details from DB
	var dbUser models.User
	if result := c.MustGet("db").(*gorm.DB).First(&dbUser, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"userID":   id,
		"userName": dbUser.Name,
		"text":     "Hello World.",
	})
}
