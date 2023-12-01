package services

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type VolumeLimitService struct {
	logger *zap.Logger
}

func NewVolumeLimitService(logger *zap.Logger) *VolumeLimitService {
	return &VolumeLimitService{logger: logger}
}

func (vs *VolumeLimitService) IsWithinLimit(folderPath string, fileSize int64, limit int) (bool, error) {
	fullFolderPath := filepath.Join(".", folderPath)

	if _, err := os.Stat(fullFolderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullFolderPath, 0o755); err != nil {
			vs.logger.Error("Failed to create folder", zap.Error(err))
			return false, err
		}
	}

	currentSize, err := vs.getFolderSize(fullFolderPath)
	if err != nil {
		vs.logger.Error("Failed to get folder size", zap.Error(err))
		return false, err
	}
	int64limit := int64(limit)

	totalSize := currentSize + fileSize

	isWithinLimit := totalSize <= int64limit

	vs.logger.Info("Folder Size",
		zap.String("folderPath", fullFolderPath),
		zap.Int64("currentSize", currentSize),
		zap.Int64("newFileSize", fileSize),
		zap.Int64("limit", int64limit),
		zap.Bool("isWithinLimit", isWithinLimit),
	)

	return isWithinLimit, nil
}

func (vs *VolumeLimitService) getFolderSize(folderPath string) (int64, error) {
	var size int64

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		vs.logger.Error("Failed to calculate folder size", zap.Error(err))
		return 0, err
	}

	return size, nil
}
