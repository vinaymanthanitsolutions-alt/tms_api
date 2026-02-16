package main

import (
	"backend/internal/config"
	"backend/web/routes"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("APP_PORT")
	ginMode := os.Getenv("GIN_MODE")

	if port == "" {
		port = "8080"
	}

	if err := config.InitDB(); err != nil {
		log.Fatal("MySQL Connection Error:", err)
	}
	defer config.CloseDB()

	gin.SetMode(ginMode)

	r := gin.Default()

	r.Use(cors.New(config.CorsConfig))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Backend running with Gin",
		})
	})
	routes.Routes(r)
	routes.Router(r)
	routes.TeamRouter(r)
	routes.ProjectRoutes(r)
	routes.TaskRoutes(r)
	routes.QueryRoutes(r)

	log.Println("Server running on port:", port)
	log.Println("Gin Mode:", ginMode)

	r.Run(":" + port)
}
