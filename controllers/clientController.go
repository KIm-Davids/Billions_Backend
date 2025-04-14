package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

//func CreateClient(c *gin.Context) {
//	var req struct {
//		Username string `json:"username"`
//		Email    string `json:"email"`
//		Password string `json:"password"`
//	}
//
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//
//	// Hash the password
//	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
//		return
//	}
//
//	//address, err := utils.GenerateAddress(10)
//	if err != nil {
//		log.Fatal("Error generating address:", err)
//	}
//	// Create the client
//	user := models.User{
//		Username: req.Username,
//		Email:    req.Email,
//		Password: string(hashedPassword),
//	}
//
//	if err := initializers.DB.Create(&user).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
//		return
//	}
//
//	ref, err := utils.GenerateAddress(5)
//	if err != nil {
//		log.Fatal("Error generating address:", err)
//	}
//	// Create the client profile
//	client := models.User{
//		UserID:     user.ID,
//		ReferrerID: ref,
//		Balance:    0.00,
//	}
//
//	if err := initializers.DB.Create(&client).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client profile"})
//		return
//	}
//
//	if err := initializers.DB.Preload("User").First(&client, client.ID).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load client with user info"})
//		return
//	}
//
//	c.JSON(http.StatusCreated, gin.H{
//		"message": "Client created successfully",
//		"Client":  client,
//	})
//}

func CreateClient(c *gin.Context) {
	var req struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		ReferredBy string `json:"referral"` // referral ID entered by user (can be null)
	}

	// Validate input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Generate unique referral ID
	var referID string
	for {
		temp, _ := utils.GenerateAddress(6)
		var count int64
		initializers.DB.Model(&models.User{}).Where("refer_id = ?", temp).Count(&count)
		if count == 0 {
			referID = temp
			break
		}
	}

	// Create user
	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   string(hashedPassword),
		ReferID:    referID,
		ReferredBy: req.ReferredBy,
		Balance:    0.0,
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"refer_id":   user.ReferID,
			"referredBy": user.ReferredBy,
			"balance":    user.Balance,
		},
	})
}

func Deposit(c *gin.Context) {
	var input models.Deposit
	var existingTx models.Deposit
	var depositCount int64

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	// Check if it's the user's first deposit
	initializers.DB.Model(&models.Deposit{}).Where("user_id = ?", input.UserID).Count(&depositCount)

	// If it's the user's first deposit
	if depositCount == 0 {
		var user models.User
		// Get the user by input.UserID (this is the referred user)
		if err := initializers.DB.First(&user, input.UserID).Error; err == nil && user.ReferID != "" {
			// Reward the referrer (e.g., credit a bonus)
			rewardReferrer(user.ReferID, input.UserID, input.Amount)
		}
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

	// Check for duplicate transaction hash
	if err := initializers.DB.Where("hash = ?", input.Hash).First(&existingTx).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash already exists"})
		return
	}

	tx := models.Deposit{
		UserID:      input.UserID,
		Email:       input.Email,
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

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Fetch the user's deposit info to get the deposit date using input.UserID
	var deposit models.Deposit
	if err := initializers.DB.Where("user_id = ?", input.UserID).Order("created_at desc").First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
		return
	}

	// Define withdrawal waiting period per package
	var waitingPeriodDays int
	var user models.User
	if err := initializers.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	switch strings.ToLower(user.Package) {
	case "test":
		waitingPeriodDays = 15
	case "pro":
		waitingPeriodDays = 30
	case "premium":
		waitingPeriodDays = 40
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown package type"})
		return
	}

	// Calculate the earliest withdrawal date (deposit date + waiting period)
	earliestWithdrawDate := deposit.CreatedAt.Add(time.Duration(waitingPeriodDays) * 24 * time.Hour)

	// Check if the current date is before the earliest withdrawal date
	if time.Now().Before(earliestWithdrawDate) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":       "Withdrawals are not allowed yet",
			"unlock_date": earliestWithdrawDate.Format("2006-01-02"),
		})
		return
	}

	// Proceed to log withdrawal
	tx := models.Withdraw{
		UserID:        input.UserID,
		SenderName:    input.SenderName,
		SenderAddress: input.SenderAddress,
		WalletType:    input.WalletType,
		Status:        input.Status,
		Amount:        input.Amount,
		Description:   input.Description,
		CreatedAt:     time.Now(),
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

func rewardReferrer(referrerID string, referredID string, depositAmount float64) {
	bonusAmount := depositAmount * 0.05 // For example, 10% of the deposit amount as a bonus

	// Update referrer's balance
	err := initializers.DB.Model(&models.User{}).
		Where("id = ?", referrerID).
		UpdateColumn("balance", gorm.Expr("balance + ?", bonusAmount)).Error
	if err != nil {
		log.Println("Failed to reward referrer:", err)
		return
	}

	// Log the referral bonus
	newBonus := models.ReferralBonus{
		ReferrerID: referrerID,
		ReferredID: referredID, // The user who made the deposit
		Amount:     bonusAmount,
		RewardedAt: time.Now(),
	}

	// Save the bonus in the ReferralBonus table
	if err := initializers.DB.Create(&newBonus).Error; err != nil {
		log.Println("Failed to log referral bonus:", err)
	}
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

var profitRates = map[string]float64{
	"test":    0.008,
	"pro":     0.01,
	"premium": 0.02,
}

func GenerateDailyProfits() {
	var deposits []models.Deposit

	// You could filter by active users or deposits
	initializers.DB.Find(&deposits)

	for _, d := range deposits {
		rate, ok := profitRates[strings.ToLower(d.PackageType)]
		if !ok {
			continue
		}

		profit := d.Amount * rate

		// Credit the user
		initializers.DB.Model(&models.User{}).
			Where("id = ?", d.UserID).
			UpdateColumn("balance", gorm.Expr("balance + ?", profit))

		// (Optional) Log profit entry
		initializers.DB.Create(&models.Profit{
			UserID:    d.UserID,
			Amount:    profit,
			Source:    "daily profit",
			CreatedAt: time.Now(),
		})
	}
}

//func GetBalance(c *gin.Context) {
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
//	var dbUser models.Client
//	if err := initializers.DB.First(&dbUser, userID).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
//		return
//	}
//
//	// Send back the balance and last updated time
//	c.JSON(http.StatusOK, gin.H{
//		"balance":   dbUser.Balance,
//		"updatedAt": dbUser.UpdatedAt,
//	})
//
//}

func GetWithdrawDate(c *gin.Context) {
	// You can receive email or user ID from query params
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Fetch the user by email
	var user models.User
	if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Fetch the user's latest deposit
	var deposit models.Deposit
	if err := initializers.DB.
		Where("user_id = ?", user.ID).
		Order("created_at desc").
		First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
		return
	}

	// Determine the number of days based on the package
	var waitingDays int
	switch strings.ToLower(user.Package) {
	case "test":
		waitingDays = 15
	case "pro":
		waitingDays = 30
	case "premium":
		waitingDays = 40
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
		return
	}

	// Calculate withdraw date
	withdrawDate := deposit.CreatedAt.Add(time.Duration(waitingDays) * 24 * time.Hour)

	// Return the withdraw date
	c.JSON(http.StatusOK, gin.H{
		"withdraw_date": withdrawDate.Format("2006-01-02"),
		"package":       user.Package,
		"days_waiting":  waitingDays,
	})
}

func GetUserInfo(c *gin.Context) {

	// Define a struct to hold the incoming data (email)
	var requestData struct {
		Email string `json:"email" binding:"required"`
	}

	// Parse the request body into the struct
	if err := c.ShouldBindJSON(&requestData); err != nil {
		// If the email is missing or invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	// Now use the email from the requestData struct
	email := requestData.Email

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var user models.User
	if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server side error User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":      user.Balance,
		"packageType":  user.Package,
		"referralCode": user.ReferID,
	})
}
