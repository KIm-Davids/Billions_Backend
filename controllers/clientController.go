package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

		// Increment the referrer's referrals count
		referrer.ReferralsCount += 1

		// Only update the referrer's record (no need to recreate the record)
		if err := initializers.DB.Save(&referrer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update referrer record"})
			return
		}
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
	if err := initializers.DB.Where("LOWER(email) = ?", input.Email).First(&user).Error; err != nil {

	}

	// Check if it's the user's first deposit
	initializers.DB.Model(&models.Deposit{}).Where("LOWER(email) = ?", input.Email).Count(&depositCount)

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
		if err := initializers.DB.Where("LOWER(email) = ?", input.Email).First(&user).Error; err != nil {
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

		// Check if the user has a referrer
		//handle referrals
		if user.ReferredBy != "" {
			// Calculate the referral bonus (5% of the deposit)
			bonus := input.Amount * 0.05

			// Fetch the referrer
			var referrer models.User
			if err := initializers.DB.Where("refer_id = ?", user.ReferredBy).First(&referrer).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Referrer not found"})
				return
			}

			bonus += bonus

			// Log the referral bonus (optional)
			referralBonus := models.ReferralBonus{
				ReferrerID:           referrer.ReferredBy,
				ReferredID:           user.ReferID,
				Amount:               bonus,
				CreatedAt:            time.Now(),
				TransactionProcessed: "true",
				Balance:              bonus,
			}
			if err := initializers.DB.Create(&referralBonus).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log referral bonus"})
				return
			}

			// Add the bonus to the referrer's balance (or profit)

			//if err := initializers.DB.Save(&referrer).Error; err != nil {
			//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit referral bonus"})
			//	return
			//}

		}

	}
	// Successfully logged the transaction and updated user balance
	c.JSON(http.StatusOK, gin.H{"message": "Transaction logged", "transaction": tx})
}

func WithdrawFromProfits(c *gin.Context) {
	var input models.Withdraw

	// Bind input JSON to struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate input
	if input.Email == "" || input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, valid amount, and withdraw_id are required"})
		return
	} else {
		// Set default fields if needed
		input.Status = "pending"
		input.CreatedAt = time.Now()

		// Save to database
		if err := initializers.DB.Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Transaction logged successfully",
			"withdraw": input,
		})
		return
	}
}

func WithdrawProfitsCtx(c *gin.Context) {

	var input struct {
		Email  string  `json:"email"`
		Amount float64 `json:"amount"`
	}

	// Check for existing withdrawals (both confirmed and pending status)
	var existingWithdrawal models.Withdraw
	err := initializers.DB.Where("LOWER(email) = ? ", input.Email).First(&existingWithdrawal).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawal record"})
		return
	}

	//fetch the latest withdrawal request that is pending then use that to fetch the amount and then fetch the sum of the profits and then minus the two together 9'9

	// If a withdrawal exists with the same withdraw_id and it's already confirmed or pending, reject the request
	if err == nil && (existingWithdrawal.Status == "withdrawn") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This withdrawal request has been withdrawn"})
		return
	}

	if existingWithdrawal.Status == "confirmed" {
		// Handle the logic to process the withdrawal
		var deposit models.Deposit
		// Find the latest confirmed deposit
		if err := initializers.DB.Where("LOWER(email) = ? AND status = ?", input.Email, "confirmed").First(&deposit).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User deposit not found"})
			return
		}

		// Define minimum withdrawal limits per package
		minProfitWithdrawal := map[string]float64{
			"test package":    10.0,
			"pro package":     50.0,
			"premium package": 100.0,
		}

		// Check if the package type is valid and retrieve the minimum amount
		packageKey := strings.ToLower(deposit.PackageType)
		minAmount, exists := minProfitWithdrawal[packageKey]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
			return
		}

		// Check if the requested amount is greater than the minimum required
		if input.Amount < minAmount {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Minimum profit withdrawal for %s package is $%.2f", packageKey, minAmount),
			})
			return
		}

		// Fetch total profit available for withdrawal
		var totalProfit float64
		if err := initializers.DB.Model(&models.Profit{}).
			Where("LOWER(email) = ? AND source = ?", input.Email, "new daily profit").
			Select("SUM(new_profit)").Scan(&totalProfit).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate profit balance"})
			return
		}

		// Check if the user has enough profit to withdraw
		if totalProfit < input.Amount {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient profit balance"})
			return
		}

		// Deduct the requested withdrawal amount from the profit record
		totalProfit -= input.Amount

		// Update withdrawal status to confirmed after processing
		existingWithdrawal.Status = "withdrawn"
		if err := initializers.DB.Save(&existingWithdrawal).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdrawal status"})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"message":    "Profit withdrawal confirmed and processed",
			"withdrawal": existingWithdrawal,
		})
		return

	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Profit withdrawal pending please wait"})
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
	if err := initializers.DB.Where("LOWER(email) = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get the latest deposit made by the user (used to calculate waiting period)
	var latestDeposit models.Deposit
	if err := initializers.DB.Where("LOWER(email) = ?", input.Email).Order("created_at desc").First(&latestDeposit).Error; err != nil {
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

func GetReferrerBonusDetails(c *gin.Context) {
	var req struct {
		ReferID string `json:"refer_id"` // This should be the unique referrer ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Fetch the referral bonus using the refer ID
	var referrerBonus models.ReferralBonus
	if err := initializers.DB.Where("referrer_id = ?", req.ReferID).First(&referrerBonus).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referral bonus record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"referrer_id":      req.ReferID,
		"referrer_balance": referrerBonus.Balance,
	})
}

//
//func RewardReferrer(c *gin.Context) {
//	var req struct {
//		Email    string `json:"email"`
//		Referrer string `json:"referrer"`
//	}
//
//	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing email"})
//		return
//	}
//
//	var user models.User
//	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	var referralBonuses []models.ReferralBonus
//	if err := initializers.DB.
//		Where("referred_id = ? AND processed = ?", user.ReferID, "false").
//		Find(&referralBonuses).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch referral bonuses"})
//		return
//	}
//
//	currentTime := time.Now()
//	if currentTime.Hour() != 18 {
//		c.JSON(http.StatusOK, gin.H{"message": "Referral bonuses are only displayed at 6 PM"})
//		return
//	}
//
//	var displayBonuses []models.ReferralBonus
//
//	for _, bonus := range referralBonuses {
//		var deposit models.Deposit
//		if err := initializers.DB.
//			Where("email = ? AND status = ?", user.Email, "confirmed").
//			Order("created_at asc").
//			First(&deposit).Error; err != nil {
//			continue
//		}
//
//		var referrer models.User
//		if err := initializers.DB.
//			Where("refer_id = ?", user.ReferredBy).
//			First(&referrer).Error; err != nil {
//			continue
//		}
//
//		// Do NOT update referrer.Profit — just collect data to return
//		displayBonuses = append(displayBonuses, bonus)
//
//		// Optional: mark as processed to avoid re-showing
//		bonus.Processed = "true"
//		initializers.DB.Save(&bonus)
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message":          "Referral bonuses fetched for display",
//		"referral_bonuses": displayBonuses,
//		"referrer":         user.ReferredBy,
//	})
//}

//func RewardReferrer(c *gin.Context) {
//	var req struct {
//		Email    string `json:"email"`
//		Referrer string `json:"referrerId"`
//	}
//
//	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing email"})
//		return
//	}
//
//	var user models.User
//	if err := initializers.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	//currentTime := time.Now()
//	//if currentTime.Hour() != 18 {
//	//	c.JSON(http.StatusOK, gin.H{"message": "Referral bonuses are only processed at 6 PM"})
//	//	return
//	//}
//
//	// Check if bonus already processed
//	var existingBonus models.ReferralBonus
//	if err := initializers.DB.
//		Where("referred_id = ? AND referrer_id = ? AND processed = ?", user.ReferID, req.Referrer, "true").
//		First(&existingBonus).Error; err == nil {
//		c.JSON(http.StatusOK, gin.H{"message": "Referral bonus already processed"})
//	}
//
//	// Get the latest confirmed deposit by the referred user
//	var deposit models.Deposit
//	if err := initializers.DB.
//		Where("email = ? AND status = ?", user.Email, "confirmed").
//		Order("created_at DESC").
//		First(&deposit).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "No confirmed deposit found"})
//	}
//
//	// Find the referrer
//	var referrer models.User
//	if err := initializers.DB.
//		Where("referred_by = ?", req.Referrer).
//		First(&referrer).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Referrer not found"})
//	}
//
//	// Calculate 5% referral bonus
//	bonusAmount := 0.05 * deposit.Amount
//
//	// Save to referral bonus table
//	newBonus := models.ReferralBonus{
//		Email:      user.Email,
//		ReferrerID: req.Referrer,
//		ReferredID: user.ReferID,
//		Amount:     bonusAmount,
//		RewardedAt: time.Now(),
//		Processed:  "true",
//	}
//
//	if err := initializers.DB.Create(&newBonus).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save referral bonus"})
//		return
//	}
//
//	//// ✅ Optionally: add to referrer’s profit table with tag
//	//newProfit := models.Profit{
//	//	Email:           referrer.Email,
//	//	Amount:          bonusAmount,
//	//	Source:          "referrer bonus",
//	//	Date:            time.Now(),
//	//	CreatedAt:       time.Now(),
//	//	NetProfitStatus: "processed",
//	//}
//
//	if err := initializers.DB.Create(&newProfit).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profit record"})
//		return
//	}
//
//	if newBonus.ReferrerID == req.Referrer {
//		c.JSON(http.StatusOK, gin.H{
//			"bonus_amount": bonusAmount,
//		})
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message":     "Referral bonus processed successfully",
//		"referrer_id": req.Referrer,
//		//"bonus_amount": bonusAmount,
//	})
//}

//func GetReferrerBonusTotal(c *gin.Context) {
//	var req struct {
//		ReferrerId string `json:"referrerId"`
//	}
//
//	if err := c.ShouldBindJSON(&req); err != nil || req.ReferrerId == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Referrer ID is required"})
//		return
//	}
//
//	var totalBonus float64
//
//	// Sum only processed bonuses for the referrer
//	err := initializers.DB.
//		Model(&models.ReferralBonus{}).
//		Where("referrer_id = ? AND processed = ?", req.ReferrerId, "true").
//		Select("COALESCE(SUM(amount), 0)").
//		Scan(&totalBonus).Error
//
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate referral bonus"})
//		return
//	}
//
//	if req.ReferrerId != tota
//
//	c.JSON(http.StatusOK, gin.H{
//		"referrer_id": req.ReferrerId,
//		"total_bonus": totalBonus,
//	})
//}

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
	//type ProfitRequest struct {
	//	Email string `json:"email"`
	//}
	//
	//type ProfitResponse struct {
	//	Email     string  `json:"email"`
	//	Profit    float64 `json:"profit"`
	//	NetProfit interface{}
	//}
	//
	//var requestBody ProfitRequest
	//if err := c.ShouldBindJSON(&requestBody); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	//	return
	//}
	//
	//email := requestBody.Email
	//location, _ := time.LoadLocation("Africa/Lagos")
	//currentTime := time.Now().In(location)
	//
	//// ✅ Check for existing profit for today
	//var profitGeneratedToday bool
	//var existingProfit models.Profit
	//err := initializers.DB.
	//	Where("email = ? AND DATE(date) = ?", email, currentTime.Format("2006-01-02")).
	//	First(&existingProfit).Error
	//
	//if err == nil {
	//	profitGeneratedToday = true // Mark that profit already exists for today
	//}
	//
	//// fetch user deposit
	//var deposit models.Deposit
	//if err := initializers.DB.
	//	Where("email = ? AND status = ?", email, "confirmed"). // Assuming `confirmed` field is a boolean
	//	Order("created_at DESC").
	//	First(&deposit).Error; err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch confirmed deposit"})
	//	return
	//}
	//
	//// ✅ Calculate days since deposit
	//daysSinceDeposit := math.Max(1, math.Floor(currentTime.Sub(deposit.CreatedAt).Hours()/24))
	//
	//// ✅ Determine rate based on package
	//var rate float64
	//switch strings.ToLower(deposit.PackageType) {
	//case "test package":
	//	rate = 0.008
	//case "pro package":
	//	rate = 0.01
	//case "premium package":
	//	rate = 0.012
	//default:
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
	//	return
	//}
	//
	//// ✅ Calculate profit
	//profitAmount := deposit.Amount * rate * daysSinceDeposit
	//if profitAmount <= 0 {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Profit amount must be greater than zero"})
	//	return
	//}
	//
	//// ✅ Save the profit record
	//newProfit := models.Profit{
	//	Email:           deposit.Email,
	//	Amount:          profitAmount,
	//	Source:          "daily profit",
	//	CreatedAt:       currentTime,
	//	Date:            currentTime,
	//	NetProfitStatus: "updatedProfit",
	//}
	//if err := initializers.DB.Create(&newProfit).Error; err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store profit"})
	//	return
	//}
	//
	//// ✅ Check if it's after 6PM in Africa/Lagos
	//sixPM := time.Date(
	//	currentTime.Year(), currentTime.Month(), currentTime.Day(),
	//	18, 0, 0, 0, location,
	//)
	//
	////message = "Profit calculated and returned. Will be added to balance at 6PM."
	//if currentTime.After(sixPM) {
	//	var user models.User
	//	if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
	//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//		return
	//	}
	//
	//	//user.Balance += profitAmount
	//	user.Profit += profitAmount
	//
	//	if err := initializers.DB.Save(&user).Error; err != nil {
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
	//		return
	//	}
	//
	//	//message = "Profit calculated and added to balance (after 6PM)."
	//}
	//
	//var latestUpdatedProfit models.Profit
	//
	//if profitGeneratedToday {
	//
	//	if err := initializers.DB.
	//		Where("email = ? ", email).
	//		Order("created_at DESC").
	//		First(&latestUpdatedProfit).Error; err != nil {
	//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//		return
	//	}
	//
	//	if latestUpdatedProfit.Source == "net profit calculation" {
	//		if err := initializers.DB.
	//			Where("email = ? AND source = ?", email, "net profit calculation").
	//			Order("created_at DESC").
	//			First(&latestUpdatedProfit).Error; err != nil {
	//			c.JSON(http.StatusNotFound, gin.H{"error": "No net profit entry with updatedProfit status found"})
	//			return
	//		}
	//
	//		var user models.User
	//		if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
	//			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//			return
	//		}
	//
	//		if user.ProfitAddedStatus == "true" {
	//			user.Balance += latestUpdatedProfit.Amount
	//			user.ProfitAddedStatus = "true"
	//			c.JSON(http.StatusConflict, gin.H{"error": "Profit already added to balance"})
	//			return
	//		}
	//
	//		if err := initializers.DB.Save(&user).Error; err != nil {
	//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
	//			return
	//		}
	//
	//	}
	//
	//	if latestUpdatedProfit.Source == "daily profit" {
	//		if err := initializers.DB.
	//			Where("email = ? AND source = ?", email, "daily profit").
	//			Order("created_at DESC").
	//			First(&latestUpdatedProfit).Error; err != nil {
	//			c.JSON(http.StatusNotFound, gin.H{"error": "No net profit entry with updatedProfit status found"})
	//			return
	//		}
	//
	//		var user models.User
	//		if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
	//			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//			return
	//		}
	//
	//		if user.ProfitAddedStatus == "true" {
	//			user.Balance += latestUpdatedProfit.Amount
	//			user.ProfitAddedStatus = "true"
	//			c.JSON(http.StatusConflict, gin.H{"error": "Profit already added to balance"})
	//			return
	//		}
	//
	//		if err := initializers.DB.Save(&user).Error; err != nil {
	//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
	//			return
	//		}
	//	}
	//}
	//
	//c.JSON(http.StatusOK, gin.H{
	//	"message":    "Latest updated profit found",
	//	"net_profit": latestUpdatedProfit.Amount,
	//	"entry":      latestUpdatedProfit,
	//})

	//New Method

	//check to see if the user has not been given profit for the day

	// Request body struct
	type DepositRequest struct {
		Email string `json:"email"`
	}

	var req DepositRequest

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	//check if already credited for daily bonus
	location, _ := time.LoadLocation("Africa/Lagos") // or your server's TZ
	//today := time.Now().In(location).Truncate(24 * time.Hour)
	currentTime := time.Now().In(location)
	today := currentTime.Format("2006-01-02") // use string format for date comparison

	//check if it already exist
	var existing models.Profit
	err := initializers.DB.Where("email = ? AND profit_date = ? AND source = ?", req.Email, today, "new daily profit").First(&existing).Error

	if err == nil { // Profit already exists for today
		c.JSON(http.StatusOK, gin.H{"received_today": true})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	//check for confirmed deposit and the time it occurred
	var deposit models.Deposit
	err = initializers.DB.Where("email = ? AND status = ?", req.Email, "confirmed").
		Order("created_at DESC").
		First(&deposit).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
		return
	}

	//location, _ = time.LoadLocation("Africa/Lagos")
	//currentTime = time.Now().In(location)

	// ✅ Calculate days since deposit
	daysSinceDeposit := math.Max(1, math.Floor(currentTime.Sub(deposit.CreatedAt).Hours()/24))

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
		Email: deposit.Email,
		//Amount:          profitAmount,
		Source:          "new daily profit",
		CreatedAt:       currentTime,
		Date:            currentTime,
		ProfitDate:      today, // this is key!
		NetProfitStatus: "profit updated",
		NewProfit:       profitAmount,
	}
	if err := initializers.DB.Create(&newProfit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store profit"})
		return
	}
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

func GetDailyProfit(c *gin.Context) {
	var req struct {
		Email string `json:"email"` // Email of the user whose daily profit is to be fetched
	}

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Fetch the user's daily profit from the profit table based on their email and source = "new daily profit"
	var totalDailyProfit float64
	if err := initializers.DB.Model(&models.Profit{}).
		Where("email = ? AND source = ?", req.Email, "new daily profit").
		Select("SUM(new_profit)").Scan(&totalDailyProfit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate daily profit"})
		return
	}

	// Return the total daily profit of the user
	c.JSON(http.StatusOK, gin.H{
		"email":        req.Email,
		"daily_profit": totalDailyProfit,
	})
	return
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

//func GetWithdrawDate(c *gin.Context) {
//	// You can receive email or user ID from query params
//	// Define a struct for the expected JSON body
//	var request struct {
//		Email string `json:"email" binding:"required"`
//	}
//
//	// Parse JSON body
//	if err := c.ShouldBindJSON(&request); err != nil || request.Email == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
//		return
//	}
//
//	email := request.Email
//
//	if email == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
//		return
//	}
//
//	// Fetch the user by email
//	var user models.User
//	if err := initializers.DB.Where("email = ?", email).First(&user).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	// Get the most recent confirmed deposit
//	var deposit models.Deposit
//	if err := initializers.DB.
//		Where("email = ? AND status = ?", email, "confirmed").
//		Order("created_at desc").
//		First(&deposit).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Deposit not found"})
//		return
//	}
//
//	// Determine the number of days based on the package
//	var waitingDays int
//	switch strings.ToLower(deposit.PackageType) {
//	case "test package":
//		waitingDays = 15
//	case "pro package":
//		waitingDays = 30
//	case "premium package":
//		waitingDays = 40
//	default:
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package type"})
//		return
//	}
//
//	// Calculate withdraw date
//	withdrawDate := deposit.CreatedAt.Add(time.Duration(waitingDays) * 24 * time.Hour)
//	withdrawDateFormatted := withdrawDate.Format("January 02, 2006") // e.g. April 30, 2025
//
//	// Return the withdraw date
//	c.JSON(http.StatusOK, gin.H{
//		"withdraw_date": withdrawDateFormatted,
//		"package":       user.Package,
//		"days_waiting":  waitingDays,
//	})
//}

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

func CountReferrals(c *gin.Context) {
	type ReferralRequest struct {
		Email string `json:"email"`
	}
	var request ReferralRequest

	// Bind the incoming JSON request body to the ReferralRequest struct
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Retrieve the user by their email to get the referral count
	var user models.User
	err := initializers.DB.Where("email = ?", request.Email).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return the referral count as a JSON response
	c.JSON(http.StatusOK, gin.H{"referral_count": user.ReferralsCount})
}
