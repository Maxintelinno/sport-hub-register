package service

import (
	"context"
	"fmt"
	"log"
	"sport-hub-register/internal/model"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	Endpoint      = "https://t3.storageapi.dev"
	Region        = "auto"
	AccessKey     = "tid_bEaHLmficNIRTCLJlflthhYLVOFh_sSDasmmWufKrsfWeQSFBZ"
	SecretKey     = "tsec_CwFWlzMyG+9FbHNE1YR64XjwOumjDF0V0dxQad64XyW_Le4XytCnFaFFkbEfD-qZC1ulxa"
	BucketName    = "stocked-pocket-jm-kiclnxm"
	BasePublicURL = "https://t3.storageapi.dev/stocked-pocket-jm-kiclnxm/"
)

type StorageService struct {
	BucketName    string
	BasePublicURL string
	Presigner     *s3.PresignClient
}

func NewStorageService() *StorageService {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               Endpoint,
			SigningRegion:     Region,
			HostnameImmutable: true,
		}, nil
	})

	cfg := aws.Config{
		Region: Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			AccessKey,
			SecretKey,
			"",
		),
		EndpointResolverWithOptions: resolver,
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &StorageService{
		BucketName:    BucketName,
		BasePublicURL: BasePublicURL,
		Presigner:     s3.NewPresignClient(client),
	}
}

func (s *StorageService) GeneratePresignedURLs(req *model.UploadPresignRequest) (*model.UploadPresignResponse, error) {
	ctx := context.Background()

	resp := &model.UploadPresignResponse{
		Files: make([]model.FileResponse, 0, len(req.Files)),
	}

	for _, f := range req.Files {
		objectKey := fmt.Sprintf("fields/%s", f.FileName)

		presignedReq, err := s.Presigner.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.BucketName),
			Key:         aws.String(objectKey),
			ContentType: aws.String(f.ContentType),
		}, s3.WithPresignExpires(10*time.Minute))
		if err != nil {
			return nil, err
		}

		publicURL := fmt.Sprintf("%s%s", s.BasePublicURL, objectKey)

		resp.Files = append(resp.Files, model.FileResponse{
			ObjectKey: objectKey,
			UploadURL: presignedReq.URL,
			PublicURL: publicURL,
		})
	}

	log.Printf("[StorageService] generated %d presigned urls", len(resp.Files))

	return resp, nil
}
