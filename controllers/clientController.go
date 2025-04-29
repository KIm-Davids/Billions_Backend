package controllers

import (
	"JWTProject/initializers"
	"JWTProject/models"
	"JWTProject/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

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

	//Generate unique referral ID
	var referID string
	for {
		temp, _ := utils.GenerateUniqueReferralID(6)
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
		user.Package = "Pro package" // Set the user's package to the package type of this deposit

		if err := initializers.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
			return
		}

		// Check if the user has a referrer
		//handle referrals

		// Add the bonus to the referrer's balance (or profit)

		//if err := initializers.DB.Save(&referrer).Error; err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit referral bonus"})
		//	return
		//}

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
			"withdraw": input.Amount,
		})
		return
	}
}

func WithdrawProfitsCtx(c *gin.Context) {

	var input struct {
		Email  string  `json:"email"`
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check for existing withdrawals (both confirmed and pending status)
	var existingWithdrawal models.Withdraw
	err := initializers.DB.
		Where("email = ? ", input.Email).
		Order("created_at DESC").
		First(&existingWithdrawal).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch latest withdrawal record"})
		return
	}

	//fetch the latest withdrawal request that is pending then use that to fetch the amount and then fetch the sum of the profits and then minus the two together 9'9

	// If a withdrawal exists with the same withdraw_id and it's already confirmed or pending, reject the request
	if err == nil && (existingWithdrawal.Status == "withdrawn") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This withdrawal request has been withdrawn"})
		return
	}

	if existingWithdrawal.Status == "completed" {
		// Handle the logic to process the withdrawal
		var deposit models.Deposit
		// Find the latest confirmed deposit
		if err := initializers.DB.Where("email = ? AND status = ?", existingWithdrawal.Email, "confirmed").First(&deposit).Error; err != nil {
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
		if existingWithdrawal.Amount < minAmount {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Minimum profit withdrawal for %s package is $%.2f", packageKey, minAmount),
			})
			return
		}

		// Fetch total profit available for withdrawal
		var totalProfit float64
		if err := initializers.DB.Model(&models.Profit{}).
			Where("email = ? AND source = ?", existingWithdrawal.Email, "new daily profit").
			Select("SUM(new_profit)").Scan(&totalProfit).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate profit balance"})
			return
		}

		// Check if the user has enough profit to withdraw
		if totalProfit < existingWithdrawal.Amount {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient profit balance"})
			return
		}

		// Deduct the requested withdrawal amount from the profit record
		//netProfit := totalProfit - input.Amount

		//if netProfit < 0 {
		//	netProfit = 0
		//}

		newProfitRecord := models.Profit{
			Email:           existingWithdrawal.Email,
			Source:          "daily profit",
			NetProfitStatus: "deducted", // or whatever status makes sense in your context
			CreatedAt:       time.Now(),
			Date:            time.Now(), // or the original profit date if applicable
			NewProfit:       -existingWithdrawal.Amount,
		}
		if err := initializers.DB.Create(&newProfitRecord).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updated net profit"})
			return
		}

		// Update withdrawal status to confirmed after processing
		existingWithdrawal.Status = "withdrawn"
		if err := initializers.DB.Save(&existingWithdrawal).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdrawal status"})
			return
		}

		//var profit models.Profit
		//err := initializers.DB.Where("email = ? AND source = ?", input.Email, "daily profit").
		//	Order("created_at DESC").
		//	First(&profit).Error
		//
		//if err != nil {
		//	if err == gorm.ErrRecordNotFound {
		//		c.JSON(http.StatusOK, gin.H{"new_profit": 0}) // No profit found
		//		return
		//	}
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		//	return
		//}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"message": "Profit withdrawal confirmed and processed",
			//"withdrawal": netProfit,
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

func ProcessReferralBonus(c *gin.Context) {
	var req struct {
		Email      string `json:"email"` // This should be the referred user's email
		ReferrerId string `json:"referrerId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Step 1: Find the referred user
	var referredUser models.User
	if err := initializers.DB.Where("email = ?", req.Email).First(&referredUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found here"})
		return
	}

	// Step 3: Check if referral bonus was already given and processed
	var existingBonus models.ReferralBonus
	var referBonus float64
	if err := initializers.DB.
		Where("referrer_id = ? AND referred_id = ? AND transaction_processed = ? ", referredUser.ReferredBy, referredUser.ReferID, "true").
		First(&existingBonus).Error; err == nil {
		return
	}

	if existingBonus.TransactionProcessed == "true" {
		referBonus += existingBonus.Balance
		c.JSON(http.StatusOK, gin.H{"error": "Referral bonus already processed", "Referrer_Bonus": referBonus})
		return
	}

	// Check if a bonus ALREADY exists for this referrer + referred user
	var count int64
	initializers.DB.Model(&models.ReferralBonus{}).
		Where("referrer_id = ? AND referred_id = ?", referredUser.ReferredBy, referredUser.ReferID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Referral bonus already processed for this user"})
		return
	}

	// Step 4: Get the latest confirmed deposit
	var deposit models.Deposit
	if err := initializers.DB.
		Where("email = ? AND status = ?", req.Email, "confirmed").
		Order("created_at desc").
		First(&deposit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No confirmed deposits found"})
		return
	}

	// Step 5: Find the referrer
	var referrer models.User
	if err := initializers.DB.Where("refer_id = ?", referredUser.ReferredBy).First(&referrer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referrer not found"})
		return
	}

	err := initializers.DB.Where("refer_id = ?", referredUser.ReferredBy).First(&referrer).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Only return if it's a real database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Referrer not found"})
		return
	}

	// Step 6: Calculate and log bonus
	bonus := deposit.Amount * 0.05

	referralBonus := models.ReferralBonus{
		Email:                referrer.Email, // Now correctly logged under the referrer's email
		ReferrerID:           referrer.ReferID,
		ReferredID:           referredUser.ReferID,
		Amount:               bonus,
		RewardedAt:           time.Now(),
		Processed:            "true",
		TransactionProcessed: "true",
		CreatedAt:            time.Now(),
		Balance:              bonus,
	}

	if err := initializers.DB.Create(&referralBonus).Error; err != nil {
		log.Println("Failed to create referral bonus:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log referral bonus"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged referral bonus successfully",
	})
}

//func FundReferrerBonus(c *gin.Context) {
//	var req struct {
//		ReferralId string `json:"referralId"`
//	}
//
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
//		return
//	}
//
//	// Step 1: Find the user making the request
//	var user models.User
//	if err := initializers.DB.Where("refer_id = ?", req.ReferralId).First(&user).Error; err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//		return
//	}
//
//	//var referrer models.User
//
//	// Step 2: Find the referrer based on ReferredBy (this is usually the refer_id)
//	//err := initializers.DB.Where("referrer_id = ?", user.ReferID).First(&referrer).Error
//	//if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
//	//	// Only return if it's a real database error
//	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking referrer"})
//	//	return
//	//}
//
//	// Optional: Check if there was no actual referrer
//	//if referrer.Email == "" {
//	//	c.JSON(http.StatusNotFound, gin.H{"error": "You have no referrer"})
//	//	return
//	//}
//
//	// Step 3: Fetch the referral bonus from the ReferralBonus table
//	var totalBonus float64
//	if err := initializers.DB.
//		Model(&models.ReferralBonus{}).
//		Where("referrer_id = ? AND transaction_processed = ?", req.ReferralId, "true").
//		Select("SUM(balance)").
//		Scan(&totalBonus).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bonus data"})
//		return
//	}
//
//	// Step 4: Ensure only the actual referrer can view their bonus
//	//if req.Email != referrer.Email {
//	//	c.JSON(http.StatusOK, gin.H{
//	//		"referrer_email": referrer.Email,
//	//		"total_bonus":    0, // Not allowed to see this bonus
//	//	})
//	//	return
//	//}
//
//	// Step 5: Return referrer's bonus
//	c.JSON(http.StatusOK, gin.H{
//		//"referrer_email": referrer.Email,
//		"total_bonus": totalBonus,
//	})
//}

func ReferBonus(c *gin.Context) {
	type ReferIDRequest struct {
		Email   string `json:"email"`
		ReferID string `json:"referrerId" binding:"required"`
	}

	var request ReferIDRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refer ID is required"})
		return
	}

	var user models.User
	if err := initializers.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.ReferredBy == "" {
		// Check if this user is a referrer (meaning others used their ReferID)
		var count int64
		if err := initializers.DB.Model(&models.User{}).
			Where("referred_by = ?", user.ReferID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if count == 0 {
			// Nobody referred by this user -> truly no referrer activity
			c.JSON(http.StatusNotFound, gin.H{"message": "User has no referrer or referrals"})
			return
		} else {

			var totalBonus float64
			if err := initializers.DB.Model(&models.ReferralBonus{}).
				Where("referrer_id = ? AND transaction_processed = ?", user.ReferID, "true").
				Select("COALESCE(SUM(balance), 0)").Scan(&totalBonus).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total referral bonus"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"referrer_id": user.ReferID,
				"total_bonus": totalBonus,
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Referral Bonus is processing"})

}

func GenerateDailyProfits(c *gin.Context) {

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
	} else {
		// Check user's balance instead of deposit
		var user models.User
		err := initializers.DB.Where("email = ?", req.Email).First(&user).Error

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Check if balance is valid
		if user.Balance <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User has no balance for profit generation"})
			return
		}

		// ✅ Determine rate based on package
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

		// ✅ Calculate profit
		profitAmount := user.Balance * 0.01
		if profitAmount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Profit amount must be greater than zero"})
			return
		}

		// ✅ Save the profit record
		newProfit := models.Profit{
			Email: user.Email,
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

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Daily profit processing..."})

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
			Email:       withdrawal.Email,
			Amount:      withdrawal.Amount,
			Status:      withdrawal.Status,
			CreatedAt:   withdrawal.CreatedAt.Format("2006-01-02 15:04:05"),
			Description: withdrawal.Description,
			//ProfitType:      withdrawal.ProfitType,
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
		Email  string  `json:"email"`
		Amount float64 `json:"withdrawAmount"` // Email of the user whose daily profit is to be fetched
	}

	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Fetch the user's daily profit from the profit table based on their email and source = "new daily profit"
	var totalDailyProfit float64
	if err := initializers.DB.Model(&models.Profit{}).
		Where("email = ?", req.Email).
		Select("COALESCE(SUM(new_profit), 0)"). // 👈 here: use COALESCE
		Scan(&totalDailyProfit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate daily profit"})
		return
	}

	//var netProfit float64
	//netProfit -= totalDailyProfit - req.Amount

	// Return the total daily profit of the user
	c.JSON(http.StatusOK, gin.H{
		"email":        req.Email,
		"daily_profit": totalDailyProfit,
	})
	return
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

func SaveWithdrawAmount(c *gin.Context) {
	var req struct {
		Email          string  `json:"email"`
		DeductedProfit float64 `json:"deductedProfit"` // 👈 profit after withdrawal
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update the user's record
	if err := initializers.DB.Model(&models.Profit{}).
		Where("email = ?", req.Email).
		Update("new_profit", req.DeductedProfit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profit updated successfully!"})
}

func ProcessReferralWithdrawal(c *gin.Context) {
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
			"withdraw": input.Amount,
		})
		return
	}
}

func WithdrawReferralBonus(c *gin.Context) {
	var req struct {
		Email         string  `json:"email"`
		DeductedBonus float64 `json:"deductedBonus"` // 👈 bonus after withdrawal
	}

	var existingReferralBonus models.ReferralBonus

	// Check if any record has "transaction_processed" = "withdrawn" for this email
	err := initializers.DB.Where("email = ? AND transaction_processed = ?", req.Email, "withdrawn").First(&existingReferralBonus).Error
	if err == nil {
		// Found a record => block the operation
		c.JSON(http.StatusBadRequest, gin.H{"error": "Referral bonus already withdrawn"})
		return
	} else if err != gorm.ErrRecordNotFound {
		// Some other DB error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Step 1: Fetch the total balance for the user
	var totalBalance float64
	if err := initializers.DB.Model(&models.ReferralBonus{}).
		Where("email = ?", req.Email).
		Select("COALESCE(SUM(balance), 0)"). // 🔥 important to prevent NULL
		Scan(&totalBalance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total referral balance"})
		return
	}

	// Step 2: Calculate the new total
	newTotal := totalBalance - req.DeductedBonus
	if newTotal < 0 {
		newTotal = 0 // prevent negative totals
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update the user's referral bonus
	// Step 1: Find the first matching ReferralBonus where transaction_processed = "true"
	var bonus models.ReferralBonus
	if err := initializers.DB.Where("email = ? AND transaction_processed = ?", req.Email, "true").
		First(&bonus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No pending referral bonus found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while fetching referral bonus"})
		return
	}

	// Step 2: Update only that one record
	bonus.Total = newTotal
	bonus.TransactionProcessed = "withdrawn"

	if err := initializers.DB.Save(&bonus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update referral bonus"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Referral bonus updated successfully!"})
}
