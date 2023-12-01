package services

import (
	"encoding/json"
	"store/models"
	"store/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StorageService struct {
	rabbitMQService RabbitMQService
	log             *zap.Logger
	db              *gorm.DB
}

func NewStorageService(rabbitMQService RabbitMQService, db *gorm.DB) *StorageService {
	log := utils.GetLogger()
	return &StorageService{rabbitMQService, log, db}
}

func (ss *StorageService) HandleFileRequests(body []byte) ([]string, error) {
	var request models.FileRequest

	if err := json.Unmarshal(body, &request); err != nil {
		ss.log.Error("Failed to unmarshal file request from RabbitMQ", zap.Error(err))
		return nil, err
	}

	fileNames, err := ss.FindFileNames(request)
	if err != nil {
		return nil, err
	}

	return fileNames, nil
}

func (ss *StorageService) FindFileNames(request models.FileRequest) ([]string, error) {
	var files []*models.File

	query := ss.BuildFileQuery(&request)

	if err := query.Find(&files).Error; err != nil {
		ss.log.Error("Failed to find files based on request", zap.Error(err))
		return nil, err
	}

	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = file.FileName
	}

	return fileNames, nil
}

func (ss *StorageService) BuildFileQuery(request *models.FileRequest) *gorm.DB {
	query := ss.db.Model(&models.File{})

	if request.Name != "" {
		query = query.Where("file_name LIKE ?", "%"+request.Name+"%")
	}

	if len(request.Tags) > 0 {
		query = query.Joins("JOIN file_file_tag ON files.id = file_file_tag.file_id").
			Joins("JOIN file_tags ON file_file_tag.file_tag_id = file_tags.id").
			Where("file_tags.name IN (?)", request.Tags).
			Group("files.id").
			Having("COUNT(DISTINCT file_tags.id) = ?", len(request.Tags))
	}

	return query
}
