package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

func GetAllUsers(c *gin.Context) {
	var users []models.User
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func ConfirmDeposit(c *gin.Context) {
	type ConfirmRequest struct {
		Email string `json:"email"`
		Hash  string `json:"hash"`
	}

	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the user
	var user models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Find the deposit record that belongs to this user and is still pending
	var deposit models.Deposit

	// Fetch the deposit with a specific DepositID, email, and status as "pending"
	if err := initializers.DB.Where("email = ? AND status = ? AND hash = ?", req.Email, "pending", req.Hash).
		First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found or already confirmed"})
		return
	}

	// Start a new transaction
	tx := initializers.DB.Begin()

	// Update the deposit status
	deposit.Status = "confirmed"
	if err := tx.Save(&deposit).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deposit status"})
		return
	}

	// Update user balance
	user.Balance += deposit.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
		return
	}

	// Commit the transaction
	tx.Commit()

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Deposit confirmed and balance updated"})
}

func RejectDeposit(c *gin.Context) {
	type RejectRequest struct {
		Email string `json:"email"`
		Hash  string `json:"hash"` // Or use another identifier
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the user by email
	var user models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Find the deposit record that belongs to this user and is still pending
	var deposit models.Deposit
	if err := initializers.DB.
		Where("email = ? AND status = ? AND hash = ?", req.Email, "pending", req.Hash).
		First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdraw not found or already processed"})
		return
	}

	// Update deposit status to 'rejected'
	deposit.Status = "rejected"

	// Save the updated deposit record
	if err := initializers.DB.Save(&deposit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdraw status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit rejected"})
}

func GetAllDeposits(c *gin.Context) {
	// Get admin email from query or token (here we're using query for simplicity)
	adminEmail := c.Query("email")

	if adminEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Check if email is admin
	if adminEmail != "admin10k4u1234@gmail.com" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: admin access only"})
		return
	}

	// Get all deposits
	var deposits []models.Deposit
	if err := initializers.DB.Order("created_at desc").Find(&deposits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deposits": deposits})
}

func GetAllWithdrawals(c *gin.Context) {
	// Get admin email from query
	adminEmail := c.Query("email")

	if adminEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Check if email is admin
	if adminEmail != "admin10k4u1234@gmail.com" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: admin access only"})
		return
	}

	// Get all withdrawals
	var withdrawals []models.Withdraw
	if err := initializers.DB.Order("created_at desc").Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

func ConfirmWithdrawProfit(c *gin.Context) {
	type ConfirmRequest struct {
		Email      string `json:"email"`
		WithdrawId uint   `json:"withdrawId"`
	}

	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find the user
	var user models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Find the withdrawal record using DepositID, email, and status = "pending"
	var withdrawal models.Withdraw
	if err := initializers.DB.Where("email = ? AND status = ? AND withdraw_id = ?", req.Email, "pending", req.WithdrawId).
		First(&withdrawal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found or already confirmed"})
		return
	}

	// Start a transaction
	tx := initializers.DB.Begin()

	// Update withdrawal status to completed
	withdrawal.Status = "completed"
	if err := tx.Save(&withdrawal).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdrawal status"})
		return
	}

	// Deduct from user's profit (or balance if that's what you're using)
	if user.Balance < withdrawal.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient profit balance"})
		return
	}
	user.Profit -= withdrawal.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
		return
	}

	// Commit the transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal confirmed and balance updated"})
}

func RejectWithdraw(c *gin.Context) {
	type RejectRequest struct {
		Email      string `json:"email"`
		WithdrawID uint   `json:"withdrawId"`
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var withdrawal models.Withdraw
	if err := initializers.DB.Where("email = ? AND withdraw_id = ? AND status = ?", req.Email, req.WithdrawID, "pending").
		First(&withdrawal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found or already processed"})
		return
	}

	// Update status to rejected
	withdrawal.Status = "rejected"
	if err := initializers.DB.Save(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal rejected successfully"})
}
