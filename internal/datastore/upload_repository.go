package datastore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	// [แก้ไข] ลบ import ที่ไม่ได้ใช้ออก และเพิ่มอันที่ถูกต้อง
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

type UploadRepository interface {
	UploadFile(ctx context.Context, path string, file io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, path string) error
	MoveFile(ctx context.Context, sourcePath string, destinationPath string) error
}

type azureUploadRepository struct {
	serviceClient *service.Client
	accountName   string
	containerName string
}

func NewAzureUploadRepository(connectionString, containerName string) (UploadRepository, error) {
	accountName, err := getAccountName(connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot parse account name: %w", err)
	}
	client, err := service.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create azure service client: %w", err)
	}
	return &azureUploadRepository{
		serviceClient: client,
		accountName:   accountName,
		containerName: containerName,
	}, nil
}

func (r *azureUploadRepository) UploadFile(ctx context.Context, path string, file io.Reader, contentType string) (string, error) {
	containerClient := r.serviceClient.NewContainerClient(r.containerName)
	blobClient := containerClient.NewBlockBlobClient(path)

	_, err := blobClient.UploadStream(ctx, file,
		// [แก้ไข] ใช้ Type ที่ถูกต้องจาก package 'blockblob'
		&blockblob.UploadStreamOptions{
			HTTPHeaders: &blob.HTTPHeaders{ // HTTPHeaders ยังคงมาจาก package blob
				BlobContentType: &contentType,
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to upload blob to azure: %w", err)
	}
	return path, nil
}

func (r *azureUploadRepository) DeleteFile(ctx context.Context, path string) error {
	containerClient := r.serviceClient.NewContainerClient(r.containerName)
	blobClient := containerClient.NewBlobClient(path)
	_, err := blobClient.Delete(ctx, nil)
	return err
}

func (r *azureUploadRepository) MoveFile(ctx context.Context, sourcePath string, destinationPath string) error {
	containerClient := r.serviceClient.NewContainerClient(r.containerName)
	sourceBlobClient := containerClient.NewBlobClient(sourcePath)
	destBlobClient := containerClient.NewBlobClient(destinationPath)

	sourceURL := sourceBlobClient.URL()

	_, err := destBlobClient.StartCopyFromURL(ctx, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("failed to start copy blob from %s: %w", sourceURL, err)
	}

	_, err = sourceBlobClient.Delete(ctx, nil)
	if err != nil {
		log.Printf("WARNING: failed to delete source blob after copy: %s, error: %v", sourcePath, err)
	}

	return nil
}

func getAccountName(connStr string) (string, error) {
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName="), nil
		}
	}
	return "", errors.New("AccountName not found in connection string")
}
