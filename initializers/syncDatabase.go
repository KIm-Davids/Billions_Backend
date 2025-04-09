package initializers

import "JWTProject/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})

}
