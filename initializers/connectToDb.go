package initializers

import (
	"JWTProject/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	//dsn := "root:password@tcp(localhost:3306)/billions_database?parseTime=true"

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	//database.AutoMigrate(&models.User{})
	database.AutoMigrate(&models.User{}, &models.Admin{})
	database.AutoMigrate(&models.User{}, &models.User{})
	database.AutoMigrate(&models.Deposit{})
	database.AutoMigrate(&models.Withdraw{})
	database.AutoMigrate(&models.Profit{})
	database.AutoMigrate(&models.ReferralBonus{})

	//database.AutoMigrate(&models.Admin{})
	//database.AutoMigrate(&models.Client{})

	// ✅ Add the unique index for daily profit only once source = 'new daily profit'
	database.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_daily_profit 
		ON profits (email, profit_date) 
		WHERE source = 'new daily profit';
	`)

	DB = database

}
