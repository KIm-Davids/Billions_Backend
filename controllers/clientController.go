package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"fmt"
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

	// Step 1: Ensure user exists and get their ID
	var user models.User
	if err := initializers.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// If not found, create user
		//user = models.User{
		//	Email: input.Email,
		//	// add other fields as needed (e.g., name, referID, etc.)
		//}
		//if err := initializers.DB.Create(&user).Error; err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		//	return
		//}
	}

	// Check if it's the user's first deposit
	initializers.DB.Model(&models.Deposit{}).Where("email = ?", input.Email).Count(&depositCount)

	// If it's the user's first deposit
	if depositCount == 0 {
		var user models.User
		// Get the user by input.UserID (this is the referred user)
		if err := initializers.DB.First(&user, input.Email).Error; err == nil && user.ReferID != "" {
			// Reward the referrer (e.g., credit a bonus)
			rewardReferrer(user.ReferID, input.Email, input.Amount)
		}
	}

	// Check for duplicate transaction hash
	if err := initializers.DB.Where("hash = ?", input.Hash).First(&existingTx).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash already exists"})
		return
	}

	// Log the deposit transaction
	tx := models.Deposit{
		//UserID:      input.UserID,
		Email:       input.Email,
		Hash:        input.Hash,
		Status:      input.Status,
		Amount:      input.Amount,
		CreatedAt:   input.CreatedAt,
		PackageType: input.PackageType,
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	// Update the user's balance if the deposit status is confirmed
	if input.Status == "confirmed" {
		var user models.User
		if err := initializers.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.Balance += input.Amount

		// Update the user's package field based on the deposit package type
		user.Package = input.PackageType // Set the user's package to the package type of this deposit

		if err := initializers.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
			return
		}

		// Successfully logged the transaction and updated user balance
		c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
	}
	//else {
	//	// If status is not confirmed, return an appropriate response
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit status not confirmed"})
	//}
}

//
//func Withdraw(c *gin.Context) {
//	var input models.Withdraw
//
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//
//	// Fetch the user's deposit info to get the deposit date using input.UserID
//	var withdrawal models.Withdraw
//	if err := initializers.DB.Where("email = ?", input.Email).Order("created_at desc").First(&withdrawal).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
//		return
//	}
//
//	// Define withdrawal waiting period per package
//	var waitingPeriodDays int
//	var user models.User
//	if err := initializers.DB.First(&user, input.UserID).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	switch strings.ToLower(user.Package) {
//	case "test":
//		waitingPeriodDays = 15
//	case "pro":
//		waitingPeriodDays = 30
//	case "premium":
//		waitingPeriodDays = 40
//	default:
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown package type"})
//		return
//	}
//
//	// Calculate the earliest withdrawal date (deposit date + waiting period)
//	earliestWithdrawDate := deposit.CreatedAt.Add(time.Duration(waitingPeriodDays) * 24 * time.Hour)
//
//	// Check if the current date is before the earliest withdrawal date
//	if time.Now().Before(earliestWithdrawDate) {
//		c.JSON(http.StatusForbidden, gin.H{
//			"error":       "Withdrawals are not allowed yet",
//			"unlock_date": earliestWithdrawDate.Format("2006-01-02"),
//		})
//		return
//	}
//
//	// Proceed to log withdrawal
//	tx := models.Withdraw{
//		//UserID:      input.UserID,
//		WalletType:  input.WalletType,
//		Status:      input.Status,
//		Amount:      input.Amount,
//		Description: input.Description,
//		CreatedAt:   time.Now(),
//	}
//
//	if err := initializers.DB.Create(&tx).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
//}

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

func GenerateDailyProfits(c *gin.Context) {
	var deposits []models.Deposit
	var userProfits map[string]float64 = make(map[string]float64)

	// Get the current time in the Africa/Lagos timezone
	location, _ := time.LoadLocation("Africa/Lagos")
	currentTime := time.Now().In(location)

	// Check if the current time is before 7 AM
	if currentTime.Hour() < 7 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profits can only be generated after 7 AM"})
		return
	}

	// Check if profits have already been generated for today
	// We are using the date part of the current date to check
	var todayProfit models.Profit
	if err := initializers.DB.Where("DATE(created_at) = ?", currentTime.Format("2006-01-02")).First(&todayProfit).Error; err == nil {
		// If profits for today already exist, skip generation
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profits for today have already been generated"})
		return
	}

	// Fetch all deposits for users who are eligible for profit
	initializers.DB.Find(&deposits)

	// Loop through each deposit and calculate daily profit
	for _, d := range deposits {
		// Determine the profit rate based on the package type
		var profitRate float64
		switch strings.ToLower(d.PackageType) {
		case "test package":
			profitRate = 0.008 // 0.8%
		case "pro package":
			profitRate = 0.01 // 1%
		case "premium package":
			profitRate = 0.012 // 1.2%
		default:
			continue // Skip unknown packages
		}

		// Calculate the profit for this deposit
		profit := d.Amount * profitRate

		// Log the individual profit entry for detailed records in the database
		profitEntry := models.Profit{
			Email:     d.Email,
			Amount:    profit,
			Source:    "daily profit",
			CreatedAt: time.Now(),
			Date:      currentTime, // Save the date when the profit was generated
		}

		// Save the individual profit entry in the Profit table for detailed tracking
		if err := initializers.DB.Create(&profitEntry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log profit"})
			return
		}

		// Accumulate the profit for the user using email as the key
		userProfits[d.Email] += profit
	}

	// Prepare the profit data to send to the frontend
	var totalProfits []map[string]interface{}
	for email, totalProfit := range userProfits {
		totalProfits = append(totalProfits, map[string]interface{}{
			"email":  email,
			"profit": totalProfit,
		})
	}

	// Return the accumulated profits (only the total) to the frontend
	c.JSON(http.StatusOK, gin.H{"profits": totalProfits})
}

//
//func GetUserProfits(c *gin.Context) {
//	email := c.Query("email")
//
//	// Fetch profits for a particular user based on their email
//	var profits []models.Profit
//	if err := initializers.DB.Where("email = ?", email).Find(&profits).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profits"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"profits": profits})
//}

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
	// Define a struct for the expected JSON body
	var request struct {
		Email string `json:"email" binding:"required"`
	}

	// Parse JSON body
	if err := c.ShouldBindJSON(&request); err != nil || request.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	email := request.Email

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

	// Get the most recent confirmed deposit
	var deposit models.Deposit
	if err := initializers.DB.
		Where("email = ? AND status = ?", email, "confirmed").
		Order("created_at desc").
		First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
		return
	}

	// Determine the number of days based on the package
	var waitingDays int
	switch strings.ToLower(deposit.PackageType) {
	case "test package":
		waitingDays = 15
	case "pro package":
		waitingDays = 30
	case "premium package":
		waitingDays = 40
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
		return
	}

	// Calculate withdraw date
	withdrawDate := deposit.CreatedAt.Add(time.Duration(waitingDays) * 24 * time.Hour)
	withdrawDateFormatted := withdrawDate.Format("January 02, 2006") // e.g. April 30, 2025

	// Return the withdraw date
	c.JSON(http.StatusOK, gin.H{
		"withdraw_date": withdrawDateFormatted,
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
	fmt.Println("Returning package:", user.Package)

	var latestDeposit models.Deposit
	if err := initializers.DB.
		Where("email = ? AND status = ?", email, "confirmed").
		Order("created_at desc").
		First(&latestDeposit).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"balance":      user.Balance,
			"packages":     nil, // or "No Package"
			"referralCode": user.ReferID,
		})
		return
	}

	//if latestDeposit.Status == "confirmed" {
	c.JSON(http.StatusOK, gin.H{
		"balance":      user.Balance,
		"packages":     latestDeposit.PackageType,
		"referralCode": user.ReferID,
		//"withdrawDate": user.
	})
	//}

}
