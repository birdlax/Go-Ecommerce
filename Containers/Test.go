// main.go (Final, Definitive, Corrected Version)
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob" // [ทางแก้ที่ถูกต้องที่สุด] import ทั้ง 2 package
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	containerName = "uploads"
)

func getAccountName(connStr string) (string, error) {
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "AccountName=") {
			return strings.TrimPrefix(part, "AccountName="), nil
		}
	}
	return "", errors.New("AccountName not found in connection string")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
	app := fiber.New()
	app.Post("/upload", uploadHandler)
	log.Println("Server started on :8080 with Fiber (Definitive Corrected Version)")
	log.Fatal(app.Listen(":8080"))
}

func uploadHandler(c *fiber.Ctx) error {
	// ส่วนนี้ถูกต้องทั้งหมด
	productId := c.FormValue("productId")
	if productId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Error retrieving the file"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot open uploaded file"})
	}
	defer file.Close()
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	blobPath := fmt.Sprintf("products/%s/%s", productId, newFileName)
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connectionString == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server configuration error: AZURE_STORAGE_CONNECTION_STRING is not set"})
	}
	accountName, err := getAccountName(connectionString)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server configuration error: " + err.Error()})
	}

	// สร้าง Service Client หลัก (ถูกต้อง)
	serviceClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect to storage"})
	}

	contentType := fileHeader.Header.Get("Content-Type")

	// [ทางแก้ที่ถูกต้องที่สุด] เราจะเรียกใช้ UploadStream จาก Service Client โดยตรง
	// และประกอบ Options ให้ถูกต้องตามนี้
	_, err = serviceClient.UploadStream(
		context.Background(),
		containerName,
		blobPath,
		file,
		// [ทางแก้ที่ถูกต้องที่สุด] Options คือ *azblob.UploadStreamOptions
		&azblob.UploadStreamOptions{
			// [ทางแก้ที่ถูกต้องที่สุด] แต่ field ข้างใน HTTPHeaders คือ *blob.HTTPHeaders
			HTTPHeaders: &blob.HTTPHeaders{
				BlobContentType: &contentType,
			},
		},
	)

	if err != nil {
		log.Printf("Failed to upload blob: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload file to Azure"})
	}

	fileURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		accountName,
		containerName,
		blobPath,
	)
	log.Printf("File uploaded successfully. URL: %s", fileURL)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "File uploaded successfully!",
		"url":     fileURL,
	})
}
