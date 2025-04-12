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
	var input models.Transaction

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userID := user.(models.User).ID
	if userID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	tx := models.Transaction{
		UserID:        userID,
		SenderName:    input.SenderName,
		SenderAddress: input.SenderAddress,
		Type:          input.Type,
		Status:        input.Status,
		Amount:        input.Amount,
		Description:   input.Description,
		PackageType:   input.PackageType,
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

func Withdraw(c *gin.Context) {
	var input models.Transaction

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

	tx := models.Transaction{
		UserID:        userID,
		SenderName:    input.SenderName,
		SenderAddress: input.SenderAddress,
		Type:          input.Type,
		Status:        input.Status,
		Amount:        input.Amount,
		Description:   input.Description,
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

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
