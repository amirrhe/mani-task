package main

import (
	"log"
	"os"
	"retreival/handlers"
	"retreival/middleware"
	"retreival/models"
	"retreival/repositories"
	"retreival/services"
	"retreival/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/streadway/amqp"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port        string
	DatabaseURL string
	SecretKey   string
	FileLimit   string
	RabbitmqUrl string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		Port:        os.Getenv("APP_PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SecretKey:   os.Getenv("SECRET_KEY"),
		FileLimit:   os.Getenv("FILE_LIMIT"),
		RabbitmqUrl: os.Getenv("RABBITMQ_URL"),
	}
}

func main() {
	config := LoadConfig()

	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	logger := utils.GetLogger()

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userRepo := repositories.NewUserRepository(db)
	jwt := services.NewJWTService(config.SecretKey)
	userService := services.NewUserService(*userRepo, logger, *jwt)
	handler := handlers.NewUserHandler(userService)
	conn, err := amqp.Dial(config.RabbitmqUrl)
	if err != nil {
		log.Fatal("Failed to connect to rabbitmq err is ", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel")
	}
	rabbitService := services.NewRabbitMQService(conn, ch)
	defer conn.Close()
	defer ch.Close()
	fileLimitInt, _ := strconv.Atoi(config.FileLimit)
	fileService := services.NewFileService(*rabbitService, fileLimitInt)
	fileHandler := handlers.NewFileHandler(fileService)
	app := fiber.New()

	v1 := app.Group("/api/v1")
	v1.Post("/user/register", handler.RegisterUser)
	v1.Post("/user/login", handler.Login)
	v1.Post("/file", middleware.JWTAuthMiddleware(jwt, logger), fileHandler.UploadFile)
	v1.Get("/file", middleware.JWTAuthMiddleware(jwt, logger), fileHandler.GetFile)
	v1.Get("/file/names", middleware.JWTAuthMiddleware(jwt, logger), fileHandler.RetrieveFileNames)

	log.Fatal(app.Listen(":" + config.Port))
}
