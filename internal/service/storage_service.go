package service

import (
	"fmt"
	"sport-hub-register/internal/model"

	"github.com/google/uuid"
)

type StorageService struct {
	BaseUploadURL string
	BasePublicURL string
}

func NewStorageService() *StorageService {
	// In a real scenario, these would come from environment variables
	return &StorageService{
		BaseUploadURL: "https://t3.storageapi.dev/",
		BasePublicURL: "https://t3.storageapi.dev/stocked-pocket-jm-kiclnxm/",
	}
}

func (s *StorageService) GeneratePresignedURLs(req *model.UploadPresignRequest) (*model.UploadPresignResponse, error) {
	resp := &model.UploadPresignResponse{
		Files: make([]model.FileResponse, 0, len(req.Files)),
	}

	for _, f := range req.Files {
		// Generate a unique object key
		objectKey := fmt.Sprintf("fields/%s-%s", uuid.New().String(), f.FileName)

		resp.Files = append(resp.Files, model.FileResponse{
			ObjectKey: objectKey,
			UploadURL: s.BaseUploadURL, // Currently a dummy URL as per user example
			PublicURL: fmt.Sprintf("%s%s", s.BasePublicURL, objectKey),
		})
	}

	return resp, nil
}
