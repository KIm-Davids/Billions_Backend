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
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	//database.AutoMigrate(&models.User{})
	database.AutoMigrate(&models.User{}, &models.Admin{})
	database.AutoMigrate(&models.User{}, &models.User{})
	database.AutoMigrate(&models.Deposit{})
	database.AutoMigrate(&models.Withdraw{})
	database.AutoMigrate(&models.ReferralBonus{})

	//database.AutoMigrate(&models.Admin{})
	//database.AutoMigrate(&models.Client{})
	DB = database

}
