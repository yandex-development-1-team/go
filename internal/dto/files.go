package dto

type FileUploadResponse struct {
	UUID     string `json:"uuid"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}
