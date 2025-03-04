package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupExpenseRoutes configures all expense-related routes
func setupExpenseRoutes(rg *gin.RouterGroup) {
	expenses := rg.Group("/expenses")
	{
		expenses.GET("", listExpenses)
		expenses.POST("", createExpense)
		expenses.GET("/:expenseId", getExpense)
		expenses.PUT("/:expenseId", updateExpense)
		expenses.DELETE("/:expenseId", deleteExpense)
		expenses.GET("/user/:userId", getExpensesByUser)
	}
}

func listExpenses(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET /expenses endpoint hit"})
}

func createExpense(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "POST /expenses endpoint hit"})
}

func getExpense(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET /expenses/:expenseId endpoint hit", "expenseId": c.Param("expenseId")})
}

func updateExpense(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "PUT /expenses/:expenseId endpoint hit", "expenseId": c.Param("expenseId")})
}

func deleteExpense(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func getExpensesByUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GET /expenses/user/:userId endpoint hit", "userId": c.Param("userId")})
}
