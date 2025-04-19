package main

import (
	"JWTProject/controllers"
	"JWTProject/initializers"
	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

//// Transaction struct matches the table schema
//type Transaction struct {
//	UserID        uint      `json:"user_id"`
//	SenderName    string    `json:"senderName"`
//	SenderAddress string    `json:"senderAddress"`
//	Type          string    `json:"transactionType"`
//	Status        string    `json:"status"`
//	PackageType   string    `json:"packageType"`
//	Amount        float64   `json:"amount"`
//	Description   string    `json:"description"`
//	CreatedAt     time.Time `json:"created_at"`
//}
//
//func createTableIfNotExists(db *sql.DB) {
//	// SQL query to create the table based on the Transaction struct
//	createTableQuery := `
//	CREATE TABLE IF NOT EXISTS transactions (
//		id INT AUTO_INCREMENT PRIMARY KEY,
//		user_id INT NOT NULL,
//		sender_name VARCHAR(255) NOT NULL,
//		sender_address VARCHAR(255) NOT NULL,
//		transaction_type VARCHAR(100) NOT NULL,
//		status VARCHAR(50) NOT NULL,
//		package_type VARCHAR(100) NOT NULL,
//		amount DECIMAL(10,2) NOT NULL,
//		description TEXT,
//		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//	);`
//
//	// Execute the query
//	_, err := db.Exec(createTableQuery)
//	if err != nil {
//		log.Fatalf("Error creating table: %v", err)
//	} else {
//		fmt.Println("Table checked/created successfully")
//	}
//}

func main() {

	// Online MySQL connection string (replace with your actual credentials)
	dsn := "root:lAqNzNxCmLbIHKWPfpyeUMbsprDYmMlq@tcp(yamabiko.proxy.rlwy.net:11897)/railway?charset=utf8mb4&parseTime=True&loc=Local"

	// Connect to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	// Check if the table exists and create it if necessary
	//createTableIfNotExists(db)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://www.billionsforextrade.vip", "https://www.billionsforextrade.vip/"},
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//adminRoutes := router.Group("/admin")
	//adminRoutes.Use(middleware.RequireAuth, middleware.IsAdminMiddleware())

	router.POST("/register/admin", controllers.RegisterAdmin)
	router.POST("/register/client", controllers.CreateClient)
	router.POST("/login", controllers.Login)
	router.POST("/deposit", controllers.Deposit)
	//router.POST("/withdraw", controllers.Withdraw)
	//router.POST("/balance", controllers.GetBalance)
	router.GET("/validate", controllers.Validate)
	//router.GET("/get/transaction", middleware.RequireAuth, controllers.GetTransactions)
	//router.GET("/get/users", controllers.GetUsers)
	//router.POST("/referralCode", controllers.GetReferralCode)
	router.POST("/withdrawDate", controllers.GetWithdrawDate)
	router.PATCH("/admin/update/usersBalance", controllers.UpdateUserBalance)
	router.POST("/getUserInfo", controllers.GetUserInfo)
	router.POST("/getAllUser", controllers.GetAllUsers)
	router.POST("/confirmDeposits", controllers.ConfirmDeposit)
	router.POST("/rejectDeposits", controllers.RejectDeposit)
	router.GET("/fetchDeposits", controllers.GetAllDeposits)
	router.GET("/fetchWithdrawals", controllers.GetAllWithdrawals)
	router.POST("/getDailyProfit", controllers.GenerateDailyProfits)
	router.POST("/withdrawBalance", controllers.WithdrawFromBalance)
	router.POST("/withdrawProfit", controllers.WithdrawFromProfits)
	router.GET("/getAllWithdrawProfit", controllers.GetUserWithdrawals)
	router.POST("/confirmDailyProfit", controllers.ConfirmWithdrawProfit)
	router.POST("/rejectWithdraw", controllers.RejectWithdraw)
	router.POST("/checkReferralBonus", controllers.RewardReferrer)
	router.POST("/getReferCount", controllers.CountUserReferrals)
	router.POST("/getReferrerCode", controllers.GetReferralCode)
	//router.POST("/getNetProfit", controllers.GetProfitBalance)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
	//c := cron.New()

	// Run daily at 7 AM
	//location, _ := time.LoadLocation("Africa/Lagos") // Or your preferred zone
	//c = cron.New(cron.WithLocation(location))
	//c.AddFunc("0 7 * * *", controllers.GenerateDailyProfits)
	//
	//currentTime := time.Now().In(location)
	//
	//// Add 1 minute to the current time to run the job 1 minute ahead
	//nextRunTime := currentTime.Add(1 * time.Minute)
	//
	//// Calculate the cron expression based on nextRunTime
	//// The cron format is: minute hour day month weekday
	//cronExpression := fmt.Sprintf("%d %d * * *", nextRunTime.Minute(), nextRunTime.Hour())
	//
	//// Add the cron job with the dynamic cron expression
	//c.AddFunc(cronExpression, controllers.GenerateDailyProfits)
	//
	//c.Start()
	//select {} // keep the program running

}

//func main() {
//	router := gin.Default()
//	router.POST("/signup", controllers.SignUp)
//
//	router.Use(cors.New(cors.Config{
//		AllowOrigins:     []string{"http://localhost:3000"},
//		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
//		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
//		ExposeHeaders:    []string{"Content-Length"},
//		AllowCredentials: true,
//		MaxAge:           12 * time.Hour,
//	}))
//
//	adminRoutes := router.Group("/admin")
//	adminRoutes.Use(middleware.RequireAuth, middleware.IsAdminMiddleware())
//
//	router.POST("/register/admin", controllers.RegisterAdmin)
//	router.POST("/register/client", controllers.CreateClient)
//	router.POST("/login", controllers.Login)
//	router.POST("/deposit", middleware.RequireAuth, controllers.CreateTransaction)
//	router.POST("/withdraw", middleware.RequireAuth, controllers.CreateTransaction)
//	router.GET("/validate", middleware.RequireAuth, controllers.Validate)
//	router.GET("/get/transaction", middleware.RequireAuth, controllers.GetTransactions)
//
//	router.Run()
//	router.Run(":8080")
//}

//database url
