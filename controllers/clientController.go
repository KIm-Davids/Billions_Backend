package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func CreateClient(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	address, err := utils.GenerateAddress(10)
	if err != nil {
		log.Fatal("Error generating address:", err)
	}
	// Create the client
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Address:  address,
		Role:     "client", // 👈 Important
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ref, err := utils.GenerateAddress(5)
	if err != nil {
		log.Fatal("Error generating address:", err)
	}
	// Create the client profile
	client := models.Client{
		UserID:  user.ID,
		Referer: ref,
		Balance: 0.00,
	}

	if err := initializers.DB.Create(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client profile"})
		return
	}

	if err := initializers.DB.Preload("User").First(&client, client.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load client with user info"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Client created successfully",
		"Client":  client,
	})
}
func Deposit(c *gin.Context) {
	var input models.Deposit
	var existingTx models.Deposit

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	//userRaw, exists := c.Get("user")
	//if !exists {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	//	return
	//}
	//
	//user, ok := userRaw.(models.User) // or *models.User if you stored a pointer
	//if !ok || user.ID == 0 {
	//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user object"})
	//	return
	//}

	if err := initializers.DB.Where("hash = ?", input.Hash).First(&existingTx).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash already exists"})
		return
	}

	tx := models.Deposit{
		UserID:      input.UserID,
		SenderName:  input.SenderName,
		Hash:        input.Hash,
		Status:      input.Status,
		Amount:      input.Amount,
		CreatedAt:   input.CreatedAt,
		PackageType: input.PackageType,
	}

	//refererInterest(, input.Amount, c)
	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

func Withdraw(c *gin.Context) {
	var input models.Withdraw

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, _ := c.Get("user")
	userID := user.(models.User).ID
	if userID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	tx := models.Withdraw{
		UserID:        userID,
		SenderName:    input.SenderName,
		SenderAddress: input.SenderAddress,
		WalletType:    input.WalletType,
		Status:        input.Status,
		Amount:        input.Amount,
		Description:   input.Description,
		CreatedAt:     input.CreatedAt,
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

//func refererInterest(user models.User, amount float64, c *gin.Context) {
//	var client models.Client
//	if err := initializers.DB.Where("user_id = ?", user.ID).First(&client).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
//		return
//	}
//
//	//if client.Balance <= 0 && !client.RefererInterestApplied {
//	//	interest := amount * 0.05
//	//	client.Balance += interest
//	//	client.RefererInterestApplied = true
//
//	if err := initializers.DB.Save(&client).Error; err != nil {
//		c.JSON(http.StatusConflict, gin.H{"error": "Unable to update client balance"})
//		return
//	}
//}

func GetBalance(c *gin.Context) {

	user, exists := c.Get("user")

	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized) // User not found in context
		return
	}

	userID := user.(models.User).ID

	if userID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	var dbUser models.Client
	if err := initializers.DB.First(&dbUser, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Send back the balance and last updated time
	c.JSON(http.StatusOK, gin.H{
		"balance":   dbUser.Balance,
		"updatedAt": dbUser.UpdatedAt,
	})

}
