package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
)

func RegisterAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if c.Bind(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	//	return
	//}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	//address, err := utils.GenerateAddress(10)
	if err != nil {
		log.Fatal("Error generating address:", err)
	}

	// Create the User
	user := models.Admin{
		//Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		//Address:  address,
		//Role:     "admin", // 👈 Important
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	admin := models.Admin{
		AdminID: user.ID,
	}

	if err := initializers.DB.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin profile"})
		return
	}

	if err := initializers.DB.Preload("User").First(&admin, admin.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load admin with user info"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin created successfully",
		"admin":   admin,
	})
}

//
//func GetTransactions(c *gin.Context) {
//
//	var transactions []models.Transaction
//	//if err := initializers.DB.Find(&transactions).Error; err != nil {
//
//	if err := initializers.DB.Order("created_at desc").Find(&transactions).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve transactions"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
//	log.Printf("Fetched transactions: %+v\n", transactions)
//
//}

//func GetUsers(c *gin.Context) {
//	var users []models.User
//
//	user, exists := c.Get("user")
//
//	if !exists {
//		c.AbortWithStatus(http.StatusUnauthorized) // User not found in context
//		return
//	}
//
//	userID := user.(models.User).ID
//
//	if userID == 0 {
//		c.AbortWithStatus(http.StatusUnauthorized)
//	}
//
//	if err := initializers.DB.Find(&users).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
//		return
//	}
//	c.JSON(http.StatusOK, users)
//}

type UpdateBalanceInput struct {
	UserID        uint    `json:"user_id" binding:"required"`
	Balance       float64 `json:"balance" binding:"required"`
	PackageName   string  `json:"package_name"`                      // Optional
	TransactionID uint    `json:"transaction_id" binding:"required"` // Optional
	NewStatus     string  `json:"new_status" binding:"required"`
}

//func UpdateUserBalance(c *gin.Context) {
//	var input UpdateBalanceInput
//
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//
//	var user models.User
//	if err := initializers.DB.First(&user, input.UserID).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	// Update balance and optionally package name
//	user.Balance = input.Balance
//	if input.PackageName != "" {
//		user.Package = input.PackageName
//	}
//
//	if err := initializers.DB.Save(&user).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message": "User updated successfully",
//		"user":    user,
//	})
//}

func UpdateUserBalance(c *gin.Context) {
	var input UpdateBalanceInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	if err := initializers.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update balance
	user.Balance = input.Balance

	// Optionally update package name
	if input.PackageName != "" {
		user.Package = input.PackageName
	}

	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Optionally update transaction status
	if input.TransactionID != 0 && input.NewStatus != "" {
		allowedStatuses := map[string]bool{
			"pending":   true,
			"active":    true,
			"failed":    true,
			"cancelled": true,
		}

		if !allowedStatuses[strings.ToLower(input.NewStatus)] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
			return
		}

		if err := initializers.DB.Model(&models.Deposit{}).
			Where("id = ?", input.TransactionID).
			Update("status", input.NewStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User and transaction (if any) updated successfully",
		"user":    user,
	})
}
