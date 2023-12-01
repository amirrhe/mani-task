package services

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"retreival/models"
	"retreival/utils"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type FileService struct {
	rabbitMQService RabbitMQService
	fileLimit       int
	log             *zap.Logger
}

func NewFileService(rabbitMQService RabbitMQService, fileLimit int) *FileService {
	log := utils.GetLogger()
	return &FileService{rabbitMQService, fileLimit, log}
}

func (fs *FileService) ExtractFileDataAndMetadata(c *fiber.Ctx) (*models.FileData, error) {
	fileType := c.FormValue("type")
	fileTags := strings.Split(c.FormValue("tags"), ",")

	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println("here1")
		fs.log.Error("Failed to retrieve file", zap.Error(err))
		return nil, utils.ErrNoFileUploaded
	}

	src, err := file.Open()
	if err != nil {
		fmt.Println("here2")
		fs.log.Error("Failed to open uploaded file", zap.Error(err))
		return nil, fmt.Errorf("failed to open uploaded file: %s", err.Error())
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		fmt.Println("here3")
		return nil, fmt.Errorf("failed to read file content: %s", err.Error())
	}

	if file.Size > int64(fs.fileLimit) {
		return nil, utils.ErrFileSizeExceedsLimit
	}

	fileData := &models.FileData{
		FileName:  file.Filename,
		FileType:  fileType,
		FileSize:  file.Size,
		FileTags:  fileTags,
		FileBytes: fileBytes,
		TagName:   fileTags,
		Type:      fileType,
	}

	return fileData, nil
}

func (fs *FileService) ProcessFileUpload(fileData *models.FileData) error {
	if fileData.FileSize > int64(fs.fileLimit) {
		return utils.ErrFileSizeExceedsLimit
	}

	err := fs.rabbitMQService.PublishFileData(fileData, "file-data-queue")
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileService) PublishFileRequest(request *models.FileRequest, queueName string) error {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fs.log.Error("Failed to marshal file request to JSON", zap.Error(err))
		return err
	}
	err = fs.rabbitMQService.publishToQueue(requestJSON, queueName)
	if err != nil {
		fs.log.Error("Failed to publish file request", zap.Error(err))
		return err
	}
	fs.log.Info("file request published successfully", zap.String("QueueName", queueName))
	return nil
}

func (fs *FileService) RetrieveFileNamesFromQueue(queueName string) ([]string, error) {
	msgs, err := fs.rabbitMQService.ConsumeQueue(queueName)
	if err != nil {
		fs.log.Error("Failed to consume from queue", zap.Error(err))
		return nil, err
	}

	for msg := range msgs {
		fileNameBytes := msg.Body

		var fileNames []string
		if err := json.Unmarshal(fileNameBytes, &fileNames); err != nil {
			fs.log.Error("Failed to unmarshal file names", zap.Error(err))
			continue
		}

		if len(fileNames) > 0 {
			return fileNames, nil
		}
	}

	return nil, nil
}
