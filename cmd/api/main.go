package main

import (
	"log"

	"github.com/asdf57/bsw/docs"
	"github.com/asdf57/bsw/internal/db"
	"github.com/asdf57/bsw/internal/http/handlers"
	httprouter "github.com/asdf57/bsw/internal/http/router"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	bswDB := db.NewBswDB()

	docs.SwaggerInfo.Title = "BSW"
	docs.SwaggerInfo.Description = "Free is good!"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	h := &handlers.Handlers{Db: bswDB}

	engine, err := httprouter.NewRouter(h)
	if err != nil {
		log.Fatalf("router init failed: %v", err)
	}

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
