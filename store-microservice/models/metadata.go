package models

import (
	"time"

	"gorm.io/gorm"
)

type FileTag struct {
	gorm.Model
	Name string
}

type File struct {
	gorm.Model
	FileName  string
	FileType  string
	FileSize  int64
	FileTags  []FileTag `gorm:"many2many:file_file_tag;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileData struct {
	FileName  string   `json:"file_name"`
	FileType  string   `json:"file_type"`
	FileSize  int64    `json:"file_size"`
	FileTags  []string `json:"file_tags"`
	FileBytes []byte   `json:"-"`
	TagName   []string `json:"tag_name"`
	Type      string   `json:"type"`
}

type FileRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
