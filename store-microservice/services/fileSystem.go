package services

import (
	"encoding/hex"

	"store/utils"

	"go.uber.org/zap"
)

type FileSystemService struct {
	log *zap.Logger
}

func NewFileSystemService(log *zap.Logger) *FileSystemService {
	return &FileSystemService{log}
}

func (fs *FileSystemService) EncryptAndSaveFile(fileBytes []byte, filePath string, key []byte) error {
	key, _ = hex.DecodeString(string(key))
	err := utils.EncryptFileBytes(fileBytes, key, filePath)
	if err != nil {
		fs.log.Error("Failed to encrypt and save file", zap.String("filePath", filePath), zap.Error(err))
		return err
	}

	fs.log.Info("File encrypted and saved successfully", zap.String("filePath", filePath))
	return nil
}

func (fs *FileSystemService) DecryptFile(filePath string, key []byte) ([]byte, error) {
	plaintext, err := utils.DecryptFile(filePath, key)
	if err != nil {
		fs.log.Error("Failed to decrypt file", zap.String("filePath", filePath), zap.Error(err))
		return nil, err
	}

	fs.log.Info("File decrypted successfully", zap.String("filePath", filePath))
	return plaintext, nil
}
