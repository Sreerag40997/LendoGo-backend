package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// UploadFileToS3 securely sends a file from Fiber to your AWS bucket using AWS SDK V2
func UploadFileToS3(fileHeader *multipart.FileHeader) (string, error) {
	// 1. Open the file sent from React
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 2. Generate a secure, unique filename (e.g., "kyc_docs/550e8400..._selfie.jpg")
	extension := filepath.Ext(fileHeader.Filename)
	uniqueFileName := fmt.Sprintf("kyc_docs/%s%s", uuid.New().String(), extension)

	// 3. Get your AWS credentials from the .env file
	bucket := os.Getenv("AWS_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")

	// 4. Load the AWS V2 Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %v", err)
	}

	// 5. Create the S3 client and Uploader
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	// 6. Perform the upload
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uniqueFileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// 7. Return the exact URL where the file is now saved!
	return result.Location, nil
}