package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"store/models"
	"store/services"
	"store/utils"
	"strconv"

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

func createQueueIfNotExist(queueName string, conn *amqp.Connection) error {
	ch, _ := conn.Channel()
	ch.QueueDeclare(queueName, true, false, false, false, nil)
	_, err := ch.QueueInspect(queueName)
	if err != nil {
		_, err := ch.QueueDeclare(
			queueName, // Queue name
			true,      // Durable
			false,     // Delete when unused
			false,     // Exclusive
			false,     // No-wait
			nil,       // Arguments
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	config := LoadConfig()
	logger := utils.GetLogger()

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
		logger.Error("Failed to connect to rabbitmq err is ", zap.Error(err))
	}
	err = createQueueIfNotExist("file-request-queue", conn)
	if err != nil {
		logger.Fatal("Failed to create or check file-request-queue ", zap.Error(err))
	}
	err = createQueueIfNotExist("file-data-queue", conn)
	if err != nil {
		logger.Fatal("Failed to create or check file-data-queue", zap.Error(err))
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Warn("Failed to open a channel")
	}
	rabbitService := services.NewRabbitMQService(conn, ch)
	// defer conn.Close()
	// defer ch.Close()

	fileService := services.NewFileSystemService(logger)
	metaDataService := services.NewMetadataService(db)
	volumeLimitService := services.NewVolumeLimitService(logger)

	storageService := services.NewStorageService(*rabbitService, db)

	fileRequestMsgs, err := rabbitService.ConsumeQueue("file-request-queue")
	if err != nil {
		logger.Warn("Failed to consume from file-request-queue", zap.Error(err))
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
		logger.Warn("Failed to consume from queue", zap.Error(err))
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
