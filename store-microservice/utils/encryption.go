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

// DecryptFile decrypts the encrypted file at the specified path using the provided key and
// returns the decrypted file bytes.
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

// package utils

// import (
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"crypto/rand"
// 	"encoding/hex"
// 	"errors"
// 	"io"
// 	"os"
// 	"path/filepath"
// )

// const (
// 	AES256KeySize = 32
// )

// func EncryptAndSaveTemp(fileBytes []byte, encryptedFilePath string, keyHex string) error {
// 	key, err := hex.DecodeString(keyHex)
// 	if err != nil {
// 		return err
// 	}

// 	if len(key) != AES256KeySize {
// 		return errors.New("invalid key size")
// 	}

// 	tempFile, err := os.CreateTemp("", "tempfile_")
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		tempFile.Close()
// 		os.Remove(tempFile.Name())
// 	}()

// 	if _, err := tempFile.Write(fileBytes); err != nil {
// 		return err
// 	}

// 	if _, err := tempFile.Seek(0, 0); err != nil {
// 		return err
// 	}
// 	fullFolderPath := filepath.Join(".", encryptedFilePath)
// 	encryptedFile, err := os.Create(fullFolderPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer encryptedFile.Close()

// 	iv := make([]byte, aes.BlockSize)
// 	if _, err := rand.Read(iv); err != nil {
// 		return err
// 	}
// 	if _, err := encryptedFile.Write(iv); err != nil {
// 		return err
// 	}

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return err
// 	}

// 	stream := cipher.NewCFBEncrypter(block, iv)
// 	writer := &cipher.StreamWriter{S: stream, W: encryptedFile}

// 	if _, err := io.Copy(writer, tempFile); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func Decrypt(filePath string, keyHex string) ([]byte, error) {
// 	key, err := hex.DecodeString(keyHex)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(key) != AES256KeySize {
// 		return nil, errors.New("invalid key size")
// 	}

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	iv := make([]byte, aes.BlockSize)
// 	if _, err := file.Read(iv); err != nil {
// 		return nil, err
// 	}

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	stream := cipher.NewCFBDecrypter(block, iv)

// 	ciphertext, err := io.ReadAll(file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	plaintext := make([]byte, len(ciphertext))
// 	stream.XORKeyStream(plaintext, ciphertext)

// 	return plaintext, nil
// }
