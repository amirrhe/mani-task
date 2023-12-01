// Inside your fileHandler.go

package handlers

import (
	"retreival/models"
	"retreival/services"
	"retreival/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type FileHandler struct {
	fileService *services.FileService
	log         *zap.Logger
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	log := utils.GetLogger()
	return &FileHandler{fileService, log}
}

func (fh *FileHandler) UploadFile(c *fiber.Ctx) error {
	fileData, err := fh.fileService.ExtractFileDataAndMetadata(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file data or metadata"})
	}

	err = fh.fileService.ProcessFileUpload(fileData)
	if err != nil {
		if err == utils.ErrFileSizeExceedsLimit {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "size limit excced"})
		}
		fh.log.Error("Failed to process file upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process file upload"})
	}

	return c.JSON(fiber.Map{"message": "Files uploaded successfully"})
}

func (fh *FileHandler) GetFile(c *fiber.Ctx) error {
	var request models.FileRequest

	name := c.Query("name")
	tags := c.Query("tags")

	if name == "" && len(tags) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Either 'name' or 'tags' is required"})
	}

	request.Name = name
	request.Tags = strings.Split(tags, ",")

	err := fh.fileService.PublishFileRequest(&request, "file-request-queue")
	if err != nil {
		fh.log.Error("Failed to send file request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send file request"})
	}

	return c.JSON(fiber.Map{"message": "File request sent successfully"})
}

func (fh *FileHandler) RetrieveFileNames(c *fiber.Ctx) error {
	fileNames, err := fh.fileService.RetrieveFileNamesFromQueue("file-names-responses")
	if err != nil {
		fh.log.Error("Failed to retrieve file names", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve file names"})
	}

	return c.JSON(fiber.Map{"fileNames": fileNames})
}
