package routes

import (
	"net/http"
	"splitter/db"
	"strconv"

	"github.com/gin-gonic/gin"
)

func setupExpenseRoutes(rg *gin.RouterGroup) {
	expenses := rg.Group("/expenses")
	{
		expenses.GET("", listExpenses)
		expenses.POST("", createExpense)
		expenses.GET("/:expenseId", getExpense)
		expenses.PUT("/:expenseId", updateExpense)
		expenses.DELETE("/:expenseId", deleteExpense)
		expenses.GET("/user/:userId", getUserExpenses)
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

	if err := repository.CreateExpense(&expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create expense"})
		return
	}
	c.JSON(http.StatusCreated, expense)
}

func getExpense(c *gin.Context) {
	expenseId, err := strconv.ParseUint(c.Param("expenseId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	expense, err := repository.GetExpense(uint(expenseId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	c.JSON(http.StatusOK, expense)
}

func updateExpense(c *gin.Context) {
	expenseId, err := strconv.ParseUint(c.Param("expenseId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	var expense db.Expense
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense.ID = uint(expenseId)
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
	expenseId, err := strconv.ParseUint(c.Param("expenseId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expense ID"})
		return
	}

	if err := repository.DeleteExpense(uint(expenseId)); err != nil {
		if err.Error() == "expense not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense"})
		return
	}
	c.Status(http.StatusNoContent)
}

func getUserExpenses(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	expenses, err := repository.GetUserExpenses(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user expenses"})
		return
	}
	c.JSON(http.StatusOK, expenses)
}
