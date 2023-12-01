package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"store/models"
	"store/services"
	"store/utils"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port        string
	DatabaseURL string
	SecretKey   string
	FileLimit   string
	RabbitmqUrl string
	FilePath    string
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
		FilePath:    os.Getenv("FILE_PATH"),
	}
}

func main() {
	config := LoadConfig()

	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	err = db.AutoMigrate(&models.File{}, &models.FileTag{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	conn, err := amqp.Dial(config.RabbitmqUrl)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ err is ", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel")
	}
	defer conn.Close()

	logger := utils.GetLogger()
	rabbitService := services.NewRabbitMQService(conn, ch)
	fileService := services.NewFileSystemService(logger)
	metaDataService := services.NewMetadataService(db)
	volumeLimitService := services.NewVolumeLimitService(logger)

	storageService := services.NewStorageService(*rabbitService, db)

	fileRequestMsgs, err := rabbitService.ConsumeQueue("file-request-queue")
	if err != nil {
		logger.Error("Failed to consume from file-request-queue", zap.Error(err))
	}

	logger.Info("Listening to 'file-request-queue'...")

	go func() {
		for {
			msg, ok := <-fileRequestMsgs
			if !ok {
				logger.Info("Channel closed, exiting")
				break
			}

			go func(msg amqp.Delivery) {
				fileNames, err := storageService.HandleFileRequests(msg.Body)
				if err != nil {
					logger.Error("Failed to handle file requests", zap.Error(err))
					return
				}
				fileNamesJSON, err := json.Marshal(fileNames)
				if err != nil {
					logger.Error("Failed to marshal file names to JSON", zap.Error(err))
					return
				}
				err = rabbitService.PublishToQueue(fileNamesJSON, "file-names-responses")
				if err != nil {
					logger.Error("Failed to publish file names to queue", zap.Error(err))
					return
				}
			}(msg)
		}
	}()

	msgs, err := rabbitService.ConsumeQueue("file-data-queue")
	if err != nil {
		logger.Error("Failed to consume from queue", zap.Error(err))

		// _, err := rabbitService.ch.QueueDeclare(
		// 	"file-data-queue", // Name of the queue
		// 	false,             // Durable
		// 	false,             // Delete when unused
		// 	false,             // Exclusive
		// 	false,             // No-wait
		// 	nil,               // Arguments
		// )
		// if err != nil {
		// 	logger.Error("Failed to create consume queue", zap.Error(err))
		// }
	}

	logger.Info("Listening to 'file-data-queue'...")

	for {
		msg, ok := <-msgs
		if !ok {
			logger.Info("Channel closed, exiting")
			break
		}

		var fileData models.FileData
		err := json.Unmarshal(msg.Body, &fileData)
		if err != nil {
			logger.Error("Failed to unmarshal file data from message", zap.Error(err))
			continue
		}

		logger.Info("Received file data", zap.String("fileName", fileData.FileName))

		err = metaDataService.SaveFileData(&fileData)
		if err != nil {
			logger.Error("Failed to save metadata in the database", zap.Error(err))
			continue
		}
		logger.Info("Metadata saved successfully", zap.String("fileName", fileData.FileName))

		intFileLimit, _ := strconv.Atoi(config.FileLimit)
		isWithinLimit, err := volumeLimitService.IsWithinLimit(config.FilePath, fileData.FileSize, intFileLimit)
		if err != nil {
			logger.Error("Error checking volume limit", zap.Error(err))
			continue
		}

		if isWithinLimit {
			filePath := filepath.Join(config.FilePath, fileData.FileName)
			err = fileService.EncryptAndSaveFile(fileData.FileBytes, filePath, []byte(config.SecretKey))
			if err != nil {
				logger.Error("Failed to save encrypted file", zap.Error(err))
				continue
			}
			logger.Info("File saved successfully", zap.String("filePath", filePath))
		} else {
			logger.Warn("Volume limit exceeded, file not saved", zap.String("fileName", fileData.FileName))
		}
	}
}
