package main

import (
	"JWTProject/controllers"
	"JWTProject/initializers"
	"JWTProject/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}
func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://new-billions-forex-project.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	adminRoutes := router.Group("/admin")
	adminRoutes.Use(middleware.RequireAuth, middleware.IsAdminMiddleware())

	router.POST("/register/admin", controllers.RegisterAdmin)
	router.POST("/register/client", controllers.CreateClient)
	router.POST("/login", controllers.Login)
	router.POST("/deposit", middleware.RequireAuth, controllers.Deposit)
	router.POST("/withdraw", middleware.RequireAuth, controllers.Withdraw)
	router.POST("/balance", middleware.RequireAuth, controllers.GetBalance)
	router.GET("/validate", middleware.RequireAuth, controllers.Validate)
	router.GET("/get/transaction", middleware.RequireAuth, controllers.GetTransactions)
	router.GET("/get/users", middleware.RequireAuth, controllers.GetUsers)

	router.Run(":8080")
}
