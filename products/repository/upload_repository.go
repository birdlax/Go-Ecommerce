package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

// UploadRepository คือ Interface (Port) ที่ Service จะเรียกใช้
// โดยไม่จำเป็นต้องรู้ว่าข้างหลังทำงานกับ Azure, S3, หรือ Google Cloud Storage
type UploadRepository interface {
	UploadFile(ctx context.Context, path string, file io.Reader, contentType string) (string, error)
}

// azureUploadRepository คือ struct ที่ทำงานกับ Azure จริงๆ (Adapter)
// มันจะ implement UploadRepository interface
type azureUploadRepository struct {
	client        *azblob.Client
	accountName   string
	containerName string
}

// NewAzureUploadRepository คือ Constructor ที่รับค่าตั้งต้นและสร้าง instance ของ repository
// สังเกตว่าฟังก์ชันนี้จะ return เป็น Interface Type เพื่อบังคับใช้หลัก Dependency Inversion
func NewAzureUploadRepository(connectionString, containerName string) (UploadRepository, error) {
	// 1. ดึง Account Name จาก Connection String เพื่อใช้สร้าง URL
	accountName, err := getAccountName(connectionString)
	if err != nil {
		return nil, fmt.Errorf("cannot parse account name: %w", err)
	}

	// 2. สร้าง Service Client หลักสำหรับเชื่อมต่อ Azure
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create azure client: %w", err)
	}

	// 3. คืนค่า struct ที่มี client และข้อมูลที่จำเป็นพร้อมใช้งาน
	return &azureUploadRepository{
		client:        client,
		accountName:   accountName,
		containerName: containerName,
	}, nil
}

// UploadFile คือ Method ที่ implement การอัปโหลดไฟล์จริงๆ
func (r *azureUploadRepository) UploadFile(ctx context.Context, path string, file io.Reader, contentType string) (string, error) {
	// เรียกใช้ SDK ของ Azure เพื่อทำการอัปโหลด
	// โดยใช้ค่าต่างๆ ที่ถูกเก็บไว้ใน struct (client, containerName)
	_, err := r.client.UploadStream(
		ctx,
		r.containerName,
		path, // path ที่ Service สร้างและส่งมาให้
		file,
		&azblob.UploadStreamOptions{
			HTTPHeaders: &blob.HTTPHeaders{
				BlobContentType: &contentType,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob to azure: %w", err)
	}

	// ประกอบร่าง Full URL เพื่อส่งคืนให้ Service
	fileURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		r.accountName,
		r.containerName,
		path,
	)

	return fileURL, nil
}

// getAccountName เป็น helper function ที่ใช้ภายใน package นี้เท่านั้น
func getAccountName(connStr string) (string, error) {
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName="), nil
		}
	}
	return "", errors.New("AccountName not found in connection string")
}
