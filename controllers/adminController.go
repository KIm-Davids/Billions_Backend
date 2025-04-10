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

	address, err := utils.GenerateAddress(10)
	if err != nil {
		log.Fatal("Error generating address:", err)
	}

	// Create the User
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Address:  address,
		Role:     "admin", // 👈 Important
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	admin := models.Admin{
		UserID: user.ID,
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

func GetTransactions(c *gin.Context) {

	var transactions []models.Transaction
	//if err := initializers.DB.Find(&transactions).Error; err != nil {

	if err := initializers.DB.Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})

}
