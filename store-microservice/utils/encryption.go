package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io/ioutil"
	"path/filepath"
)

func EncryptFileBytes(fileBytes []byte, key []byte, filePath string) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	fullFolderPath := filepath.Join(".", filePath)

	ciphertext := make([]byte, aes.BlockSize+len(fileBytes))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], fileBytes)

	encryptedFilePath := fullFolderPath + ".encrypted"

	if err := ioutil.WriteFile(encryptedFilePath, ciphertext, 0o644); err != nil {
		return err
	}

	return nil
}

func DecryptFile(filePath string, key []byte) ([]byte, error) {
	ciphertext, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("invalid ciphertext")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return ciphertext, nil
}
