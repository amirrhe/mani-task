package models

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
