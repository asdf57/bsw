package main

import (
	"log"

	"github.com/asdf57/bsw/controller"
	database "github.com/asdf57/bsw/db"
	"github.com/asdf57/bsw/docs"
	_ "github.com/asdf57/bsw/docs"
	"github.com/asdf57/bsw/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @basePath  /

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	db := database.CreateDB()

	// Create the tables
	log.Println("Creating the payment table")
	if err := db.AutoMigrate(&models.PaymentDBEntry{}); err != nil {
		log.Fatal("I have failed to create the Payment table")
	}

	log.Println("Creating the user table")
	if err := db.AutoMigrate(&models.UserDBEntry{}); err != nil {
		log.Fatal("I have failed to create the User table")
	}

	log.Println("Creating the balance table")
	if err := db.AutoMigrate(&models.BalanceDBEntry{}); err != nil {
		log.Fatal("I have failed to create the Balance table")
	}

	docs.SwaggerInfo.Title = "BSW"
	docs.SwaggerInfo.Description = "Free is good!"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	router := gin.Default()
	router.Use(cors.Default())

	ctrl := controller.NewController(db)

	v1 := router.Group("/api/v1")
	{
		payments := v1.Group("/payment")
		{
			payments.GET("/:id", ctrl.GetPayment)
			payments.GET("/all", ctrl.GetPayments)
			payments.POST("", ctrl.PostPayment)
			payments.DELETE("/:id", ctrl.DeletePayment)
		}

		user := v1.Group("/user")
		{
			user.POST("", ctrl.PostUser)
			user.GET("/:name", ctrl.GetUser)
			user.GET("/all", ctrl.GetUsers)
			user.DELETE("/:name", ctrl.DeleteUser)
		}

		balance := v1.Group("/balance")
		{
			balance.GET("/all", ctrl.GetBalances)
			balance.POST("/all", ctrl.UpdateBalances)
		}

		health := v1.Group("/health")
		{
			health.GET("", ctrl.GetDbHealth)
		}

		admin := v1.Group("/admin")
		{
			admin.POST("/backup", ctrl.BackupDB)
			admin.GET("/backup/:filename", ctrl.GetBackup)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":8080")
}
