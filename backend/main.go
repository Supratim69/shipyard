package main

import (
	"log"

	"oneclickdevenv/backend/db"
	"oneclickdevenv/backend/middleware"
	"oneclickdevenv/backend/routes"
	"oneclickdevenv/backend/services"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system env")
	}

	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	if err := db.Migrate(); err != nil {
		log.Fatal("DB migration failed:", err)
	}

	services.StartTTLReaper()

	r := gin.Default()

	auth := r.Group("/api")
	auth.Use(middleware.JWTAuthMiddleware())
	auth.POST("/provision", routes.ProvisionVM)
	auth.POST("/destroy", routes.DestroyVM)

	r.GET("/api/auth/github/login", routes.GitHubLogin)
	r.GET("/api/auth/github/callback", routes.GitHubCallback)

	log.Println("Server running on :8080")
	r.Run(":8080")
}
