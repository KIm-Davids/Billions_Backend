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

	}
	// Successfully logged the transaction and updated user balance
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

func WithdrawFromProfits(c *gin.Context) {
	var input models.Withdraw

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Bind error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.Email == "" || input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and valid amount are required"})
		return
	}

	// Log the withdrawal regardless of confirmation status
	tx := models.Withdraw{
		WithdrawAddress: input.WithdrawAddress,
		Email:           input.Email,
		WalletType:      input.WalletType,
		Status:          input.Status,
		Amount:          input.Amount,
		Description:     input.Description,
		CreatedAt:       time.Now(),
	}

	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log withdrawal"})
		return
	}

	// For non-confirmed withdrawals, just log it and return success
	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal logged and pending admin confirmation", "withdrawal": tx})

	// Only process balance deduction if status is confirmed
	if input.Status == "confirmed" {
		var deposit models.Deposit
		if err := initializers.DB.Where("email = ?", input.Email).First(&deposit).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User deposit not found"})
			return
		}

		minProfitWithdrawal := map[string]float64{
			"test package":    10.0,
			"pro package":     50.0,
			"premium package": 100.0,
		}

		packageKey := strings.ToLower(deposit.PackageType)
		minAmount, exists := minProfitWithdrawal[packageKey]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
			return
		}

		if input.Amount < minAmount {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Minimum profit withdrawal for %s package is $%.2f", packageKey, minAmount),
			})
			return
		}

		var totalProfit float64
		if err := initializers.DB.Model(&models.Profit{}).
			Where("email = ? AND source = ?", input.Email, "daily profit").
			Select("SUM(amount)").Scan(&totalProfit).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate profit balance"})
			return
		}

		if totalProfit < input.Amount {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient profit balance"})
			return
		}

		var user models.User
		if err := initializers.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if input.Status == "confirmed" {
			user.Profit -= input.Amount
			user.Package = deposit.PackageType
		}

		if err := initializers.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profit"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profit withdrawal confirmed and processed", "withdrawal": tx})
		return
	}

}

func WithdrawFromBalance(c *gin.Context) {
	var input models.Withdraw

	// Bind incoming JSON request to the Withdraw model
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Ensure that email and amount are provided
	if input.Email == "" || input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and valid amount are required"})
		return
	}

	// Find the user by email
	var user models.User
	if err := initializers.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get the latest deposit made by the user (used to calculate waiting period)
	var latestDeposit models.Deposit
	if err := initializers.DB.Where("email = ?", input.Email).Order("created_at desc").First(&latestDeposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No deposit found for user"})
		return
	}

	// Define waiting period for each package
	var waitingDays int
	switch strings.ToLower(user.Package) {
	case "test":
		waitingDays = 15
	case "pro":
		waitingDays = 30
	case "premium":
		waitingDays = 40
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown package type"})
		return
	}

	// Calculate when the user is allowed to withdraw from their main balance
	earliestWithdrawDate := latestDeposit.CreatedAt.Add(time.Hour * 24 * time.Duration(waitingDays))

	// If the current date is before the allowed date, block the withdrawal
	if time.Now().Before(earliestWithdrawDate) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":       "Main balance withdrawal is not allowed yet",
			"unlock_date": earliestWithdrawDate.Format("2006-01-02"),
		})
		return
	}

	// Check if the user has enough balance
	if user.Balance < input.Amount {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient main balance"})
		return
	}

	// Deduct the amount from the user's main balance
	user.Balance -= input.Amount
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
		return
	}

	// Log the withdrawal transaction in the database
	tx := models.Withdraw{
		Email:       input.Email,
		WalletType:  input.WalletType,
		Status:      input.Status,
		Amount:      input.Amount,
		Description: input.Description,
		//Source:      "main", // Indicate it's from the main balance
		CreatedAt: time.Now(),
	}

	// Save the withdrawal record
	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log withdrawal"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Main balance withdrawal logged", "withdrawal": tx})
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

var profitRates = map[string]float64{
	"test":    0.008,
	"pro":     0.01,
	"premium": 0.02,
}

func GenerateDailyProfits(c *gin.Context) {
	var deposits []models.Deposit
	userProfits := make(map[string]float64)

	location, _ := time.LoadLocation("Africa/Lagos")
	currentTime := time.Now().In(location)
	todayDate := currentTime.Format("2006-01-02")

	// ✅ Check if profits were already generated today
	var existingProfits []models.Profit
	if err := initializers.DB.Where("DATE(created_at) = ?", todayDate).Find(&existingProfits).Error; err == nil && len(existingProfits) > 0 {
		// ✅ Return the already generated profits
		for _, p := range existingProfits {
			userProfits[p.Email] += p.Amount
		}

		var totalProfits []map[string]interface{}
		for email, profit := range userProfits {
			totalProfits = append(totalProfits, map[string]interface{}{
				"email":  email,
				"profit": profit,
			})
		}

		c.JSON(http.StatusOK, gin.H{"profits": totalProfits, "message": "Profits already generated for today"})
		return
	}

	// 🕖 Only allow profit generation after 7AM
	if currentTime.Hour() < 7 {
		c.JSON(http.StatusOK, gin.H{"profits": []map[string]interface{}{}, "message": "Profits not yet generated. Check after 7AM"})
		return
	}

	// 🚀 Generate profits if not already done
	initializers.DB.Find(&deposits)
	for _, d := range deposits {
		var rate float64
		switch strings.ToLower(d.PackageType) {
		case "test package":
			rate = 0.008
		case "pro package":
			rate = 0.01
		case "premium package":
			rate = 0.012
		default:
			continue
		}

		profitAmount := d.Amount * rate
		userProfits[d.Email] += profitAmount

		// Add a profit record
		newProfit := models.Profit{
			Email:     d.Email,
			Amount:    profitAmount,
			Source:    "daily profit",
			CreatedAt: currentTime,
			Date:      currentTime,
		}
		if err := initializers.DB.Create(&newProfit).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store profit"})
			return
		}

		// Increment user's total profit
		var user models.User
		if err := initializers.DB.Where("email = ?", d.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if err := initializers.DB.Model(&user).
			Update("profit", gorm.Expr("profit + ?", profitAmount)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profit"})
			return
		}
	}

	var totalProfits []map[string]interface{}
	for email, profit := range userProfits {
		totalProfits = append(totalProfits, map[string]interface{}{
			"email":  email,
			"profit": profit,
		})
	}

	c.JSON(http.StatusOK, gin.H{"profits": totalProfits, "message": "Profits successfully generated"})
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
