package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math"
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

	// Validate referral code if provided and get the referrer's user ID
	var referrerID string
	if req.ReferredBy != "" {
		var referrer models.User
		if err := initializers.DB.Where("refer_id = ?", req.ReferredBy).First(&referrer).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid referral code"})
			return
		}
		referrerID = referrer.ReferID
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
		ReferredBy: referrerID,
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

	}

	// Check if it's the user's first deposit
	initializers.DB.Model(&models.Deposit{}).Where("email = ?", input.Email).Count(&depositCount)

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

	bonusAmount := input.Amount * 0.05

	if err := initializers.DB.
		Model(&models.Deposit{}).
		Where("email = ?", user.Email).
		Count(&depositCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count deposits"})
		return
	}

	// If this is the user's first-ever deposit (any status)
	//if depositCount == 1 && user.ReferredBy != "" {
	// Log the referral bonus for processing later
	referralBonus := models.ReferralBonus{
		ReferrerID: user.ReferredBy, // should be the referrer’s email or refer_id
		ReferredID: user.ReferID,    // the new user's ID or refer_id
		Amount:     bonusAmount,     // maybe based on % of deposit
		Processed:  "false",
		RewardedAt: time.Now(),
	}

	if err := initializers.DB.Create(&referralBonus).Error; err != nil {
		log.Println("Failed to log referral bonus:", err)
	}
	//}

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
		ProfitType:      "daily_profit",
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
	switch strings.ToLower(latestDeposit.PackageType) {
	case "test package":
		waitingDays = 15
	case "pro package":
		waitingDays = 30
	case "premium package":
		waitingDays = 40
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown package type", "Package type": latestDeposit})
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
		Source:      "main", // Indicate it's from the main balance
		CreatedAt:   time.Now(),
	}

	// Save the withdrawal record
	if err := initializers.DB.Create(&tx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log withdrawal"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Main balance withdrawal logged", "withdrawal": tx})
}

func RewardReferrer(c *gin.Context) {
	// Step 1: Get email from request body
	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing email"})
		return
	}

	// Step 1: Get user by email
	var user models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Step 2: Fetch unprocessed referral bonuses for the user
	var referralBonuses []models.ReferralBonus
	if err := initializers.DB.
		Where("referred_id = ? AND processed = ?", user.ReferID, "false").
		Find(&referralBonuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch referral bonuses"})
		return
	}

	// ✅ Step 2.5: Check if it's 6 PM
	currentTime := time.Now()
	if currentTime.Hour() != 18 {
		c.JSON(http.StatusOK, gin.H{"message": "Referral bonuses are only processed at 6 PM"})
		return
	}

	// Step 3: Loop through each bonus and reward referrer if deposit confirmed
	for _, bonus := range referralBonuses {
		var deposit models.Deposit
		if err := initializers.DB.
			Where("email = ? AND status = ?", user.Email, "confirmed").
			Order("created_at asc").
			First(&deposit).Error; err != nil {
			continue // skip if no confirmed deposit yet
		}

		var referrer models.User
		if err := initializers.DB.
			Where("refer_id = ?", user.ReferredBy).
			First(&referrer).Error; err != nil {
			continue // skip if referrer not found
		}

		// Credit the bonus to the referrer’s balance
		referrer.Profit += bonus.Amount
		if err := initializers.DB.Save(&referrer).Error; err != nil {
			continue // skip if save failed
		}

		// Mark bonus as processed
		bonus.Processed = "true"
		initializers.DB.Save(&bonus)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Referral bonuses processed successfully",
		"referral_code": user.ReferredBy, // or user.ReferredBy depending on what you want to expose
	})
}

var profitRates = map[string]float64{
	"test":    0.008,
	"pro":     0.01,
	"premium": 0.02,
}

//
//func GenerateDailyProfits(c *gin.Context) {
//	type ProfitRequest struct {
//		Email string `json:"email"`
//	}
//
//	type ProfitResponse struct {
//		Email  string  `json:"email"`
//		Profit float64 `json:"profit"`
//	}
//
//	var requestBody ProfitRequest
//	if err := c.ShouldBindJSON(&requestBody); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
//		return
//	}
//
//	email := requestBody.Email
//	userProfits := make(map[string]float64)
//
//	location, _ := time.LoadLocation("Africa/Lagos")
//	currentTime := time.Now().In(location)
//
//	var deposit models.Deposit
//	if err := initializers.DB.Where("email = ?", email).Order("created_at DESC").First(&deposit).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposit"})
//		return
//	}
//
//	//✅ Prevent duplicate profit for today
//	var existingProfit models.Profit
//	if err := initializers.DB.
//		Where("email = ? AND DATE(date) = ?", email, currentTime.Format("2006-01-02")).
//		First(&existingProfit).Error; err == nil {
//		c.JSON(http.StatusConflict, gin.H{"message": "Profit already generated for today"})
//		return
//	}
//
//	daysSinceDeposit := math.Floor(currentTime.Sub(deposit.CreatedAt).Hours() / 24)
//
//	var rate float64
//	switch strings.ToLower(deposit.PackageType) {
//	case "test package":
//		rate = 0.008
//	case "pro package":
//		rate = 0.01
//	case "premium package":
//		rate = 0.012
//	default:
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
//		return
//	}
//
//	profitAmount := deposit.Amount * rate * daysSinceDeposit
//	if profitAmount <= 0 {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Profit amount must be greater than zero"})
//		return
//	}
//
//	userProfits[deposit.Email] += profitAmount
//
//	// ✅ Save the profit record (but don’t add to balance yet)
//	newProfit := models.Profit{
//		Email:     deposit.Email,
//		Amount:    profitAmount,
//		Source:    "daily profit",
//		CreatedAt: currentTime,
//		Date:      currentTime,
//	}
//	if err := initializers.DB.Create(&newProfit).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store profit"})
//		return
//	}
//
//	msg := "Profit calculated and returned. Will be added to balance at 6PM."
//
//	var totalProfits []ProfitResponse
//	if profit, exists := userProfits[email]; exists {
//		totalProfits = append(totalProfits, ProfitResponse{
//			Email:  email,
//			Profit: profit,
//		})
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"profits": totalProfits,
//		"message": msg,
//	})
//
//	// ✅ Check if it's after 6PM in Africa/Lagos
//	sixPM := time.Date(
//		currentTime.Year(), currentTime.Month(), currentTime.Day(),
//		18, 0, 0, 0, location,
//	)
//
//	if currentTime.After(sixPM) {
//		var user models.User
//		if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
//			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//			return
//		}
//
//		user.Balance += profitAmount
//		user.Profit += profitAmount
//
//		if err := initializers.DB.Save(&user).Error; err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
//			return
//		}
//
//		msg = "Profit calculated and added to balance (after 6PM)."
//	}
//
//	//var totalProfits []ProfitResponse
//	if profit, exists := userProfits[email]; exists {
//		totalProfits = append(totalProfits, ProfitResponse{
//			Email:  email,
//			Profit: profit,
//		})
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"profits": totalProfits,
//		"message": msg,
//	})
//}

func GenerateDailyProfits(c *gin.Context) {
	type ProfitRequest struct {
		Email string `json:"email"`
	}

	type ProfitResponse struct {
		Email  string  `json:"email"`
		Profit float64 `json:"profit"`
	}

	var requestBody ProfitRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := requestBody.Email
	location, _ := time.LoadLocation("Africa/Lagos")
	currentTime := time.Now().In(location)

	// ✅ Check for existing profit for today
	var existingProfit models.Profit
	if err := initializers.DB.
		Where("email = ? AND DATE(date) = ?", email, currentTime.Format("2006-01-02")).
		First(&existingProfit).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"profits": []ProfitResponse{
				{
					Email:  existingProfit.Email,
					Profit: existingProfit.Amount,
				},
			},
			"message": "Profit already generated for today",
		})
		return
	}

	// ✅ Get user's latest deposit
	var deposit models.Deposit
	if err := initializers.DB.Where("email = ?", email).Order("created_at DESC").First(&deposit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deposit"})
		return
	}

	// ✅ Calculate days since deposit
	daysSinceDeposit := math.Floor(currentTime.Sub(deposit.CreatedAt).Hours() / 24)

	// ✅ Determine rate based on package
	var rate float64
	switch strings.ToLower(deposit.PackageType) {
	case "test package":
		rate = 0.008
	case "pro package":
		rate = 0.01
	case "premium package":
		rate = 0.012
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
		return
	}

	// ✅ Calculate profit
	profitAmount := deposit.Amount * rate * daysSinceDeposit
	if profitAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profit amount must be greater than zero"})
		return
	}

	// ✅ Save the profit record
	newProfit := models.Profit{
		Email:     deposit.Email,
		Amount:    profitAmount,
		Source:    "daily profit",
		CreatedAt: currentTime,
		Date:      currentTime,
	}
	if err := initializers.DB.Create(&newProfit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store profit"})
		return
	}

	// ✅ Check if it's after 6PM in Africa/Lagos
	sixPM := time.Date(
		currentTime.Year(), currentTime.Month(), currentTime.Day(),
		18, 0, 0, 0, location,
	)

	message := "Profit calculated and returned. Will be added to balance at 6PM."
	if currentTime.After(sixPM) {
		var user models.User
		if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.Balance += profitAmount
		user.Profit += profitAmount

		if err := initializers.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
			return
		}

		message = "Profit calculated and added to balance (after 6PM)."
	}

	// ✅ Return profit response
	c.JSON(http.StatusOK, gin.H{
		"profits": []ProfitResponse{
			{
				Email:  email,
				Profit: profitAmount,
			},
		},
		"message": message,
	})
}

func GetReferralCode(c *gin.Context) {
	type RequestBody struct {
		Email string `json:"email"`
	}

	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	var user models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"referral_code": user.ReferID})
}

func GetUserWithdrawals(c *gin.Context) {
	// Get the email from query parameters
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
		return
	}

	// Define the structure of the response
	type WithdrawResponse struct {
		Email           string  `json:"email"`
		Amount          float64 `json:"amount"`
		Status          string  `json:"status"`
		CreatedAt       string  `json:"created_at"`
		Description     string  `json:"description"`
		ProfitType      string  `json:"profit_type"`
		WalletType      string  `json:"wallet_type"`
		WithdrawAddress string  `json:"withdrawAddress"`
		WithdrawId      uint    `json:"withdrawId"`
		Source          string  `json:"source"` // ← Added `json` tag here
	}

	// Prepare the user's withdrawal info
	var withdrawals []models.Withdraw

	// Find all withdrawal records (no filtering by user)
	if err := initializers.DB.Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve withdrawals"})
		return
	}

	// Check if there are no withdrawals found for the user
	if len(withdrawals) == 0 {
		c.JSON(http.StatusOK, gin.H{"withdrawals": []WithdrawResponse{}})
		return
	}

	// Prepare the response
	var withdrawResponse []WithdrawResponse
	for _, withdrawal := range withdrawals {
		withdrawResponse = append(withdrawResponse, WithdrawResponse{
			Email:           withdrawal.Email,
			Amount:          withdrawal.Amount,
			Status:          withdrawal.Status,
			CreatedAt:       withdrawal.CreatedAt.Format("2006-01-02 15:04:05"),
			Description:     withdrawal.Description,
			ProfitType:      withdrawal.ProfitType,
			WalletType:      withdrawal.WalletType,
			WithdrawAddress: withdrawal.WithdrawAddress,
			WithdrawId:      withdrawal.WithdrawID,
			Source:          "main", // ← You can change this to anything meaningful

		})
	}

	// Return the withdrawal data
	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawResponse})
}

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

func CountUserReferrals(c *gin.Context) {
	type ReferralRequest struct {
		Email string `json:"email"`
	}

	var request ReferralRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get the user by email to retrieve their ReferralID
	var user models.User
	if err := initializers.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Count how many users have this user's referral ID as their referrer
	var count int64
	if err := initializers.DB.Model(&models.User{}).
		Where("refer_id = ?", user.ReferID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count referrals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":          user.Email,
		"referral_id":    user.ReferID,
		"referral_count": count,
	})
}
