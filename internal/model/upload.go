package model

type UploadPresignRequest struct {
	Files []FileRequest `json:"files" validate:"required,dive"`
}

type FileRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type UploadPresignResponse struct {
	Files []FileResponse `json:"files"`
}

type FileResponse struct {
	ObjectKey string `json:"object_key"`
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}
