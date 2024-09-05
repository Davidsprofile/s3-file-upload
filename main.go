package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucketName = "file-upload-project-dt"
	region     = "eu-central-1"
)

var s3Client *s3.Client

func main() {
	// Initialize the S3 client
	s3Client = initS3Client()

	// Route to handle file uploads
	http.HandleFunc("/upload", uploadFileHandler)

	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Initialize S3 client using the AWS SDK
func initS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return s3.NewFromConfig(cfg)
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a maximum upload size of 10MB
	r.ParseMultipartForm(10 << 20) // 10MB

	// Retrieve the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Generate a unique filename with a timestamp
	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(handler.Filename))

	// Upload the file to S3
	err = uploadToS3(filename, file)
	if err != nil {
		http.Error(w, "Error uploading the file to S3", http.StatusInternalServerError)
		return
	}

	// Respond with the S3 file URL
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, filename)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", fileURL)
}

func uploadToS3(filename string, file multipart.File) error {
	// Create an uploader with the S3 client
	uploader := manager.NewUploader(s3Client)

	// Upload the file to S3 without the ACL option
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		return err
	}

	log.Printf("File uploaded successfully to S3: %s", filename)
	return nil
}
