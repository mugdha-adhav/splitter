package routes

import (
	"net/http"
	"splitter/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	expenses, err := repository.ListExpenses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expenses"})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

func createExpense(c *gin.Context) {
	var expense db.Expense
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate UUID for new expense
	expense.ID = uuid.New().String()

	if err := repository.CreateExpense(&expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func getExpense(c *gin.Context) {
	expenseId := c.Param("expenseId")
	expense, err := repository.GetExpense(expenseId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	c.JSON(http.StatusOK, expense)
}

func updateExpense(c *gin.Context) {
	expenseId := c.Param("expenseId")
	var expense db.Expense
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the ID from path parameter
	expense.ID = expenseId

	if err := repository.UpdateExpense(&expense); err != nil {
		if err.Error() == "expense not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func deleteExpense(c *gin.Context) {
	expenseId := c.Param("expenseId")
	if err := repository.DeleteExpense(expenseId); err != nil {
		if err.Error() == "expense not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense"})
		return
	}
	c.Status(http.StatusNoContent)
}

func getExpensesByUser(c *gin.Context) {
	userId := c.Param("userId")
	expenses, err := repository.GetUserExpenses(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user expenses"})
		return
	}
	c.JSON(http.StatusOK, expenses)
}
