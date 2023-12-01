package services

import (
	"time"

	"store/models"

	"gorm.io/gorm"
)

type MetadataService struct {
	db *gorm.DB
}

func NewMetadataService(db *gorm.DB) *MetadataService {
	return &MetadataService{db}
}

func (ms *MetadataService) SaveFileData(fileData *models.FileData) error {
	file := models.File{
		FileName:  fileData.FileName,
		FileType:  fileData.FileType,
		FileSize:  fileData.FileSize,
		FileTags:  []models.FileTag{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, tagName := range fileData.FileTags {
		tag := models.FileTag{Name: tagName}
		file.FileTags = append(file.FileTags, tag)
	}

	if err := ms.db.Create(&file).Error; err != nil {
		return err
	}

	return nil
}
