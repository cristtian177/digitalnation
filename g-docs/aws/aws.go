package aws

import (
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client struct {
	Session *session.Session
	Bucket  string
}

func NewS3Client(session *session.Session, bucket string) *S3Client {
	return &S3Client{Session: session, Bucket: bucket}
}

func (c *S3Client) Upload(file io.Reader, keyName string) (string, error) {
	// Crea un nuevo uploader con la sesión y la configuración por defecto
	uploader := s3manager.NewUploader(c.Session)

	// Sube el archivo a S3
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(keyName),
		Body:   file,
		ACL:    aws.String("public-read"), // Hace que el objeto sea público
	})
	log.Println("File uploaded to", c.Bucket, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	/*

		// Genera una URL pre-firmada para el objeto
		svc := s3.New(c.Session)
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(keyName),
		})

		urlStr, err := req.Presign(15 * time.Minute) // URL válida por 15 minutos
		if err != nil {
			return "", fmt.Errorf("failed to sign request, %v", err)
		}
	*/

	urlStr := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", c.Bucket, keyName)

	return urlStr, nil
}

func (c *S3Client) Delete(keyName string) error {
	// Crea un nuevo servicio S3
	svc := s3.New(c.Session)

	// Elimina el archivo de S3
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(keyName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file, %v", err)
	}

	return nil
}

func (c *S3Client) DeleteFolder(folderName string) error {
	// Crea un nuevo servicio S3
	svc := s3.New(c.Session)

	// Lista todos los objetos en la carpeta
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(c.Bucket),
		Prefix: aws.String(folderName),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects in folder, %v", err)
	}

	// Elimina cada objeto en la carpeta
	for _, item := range resp.Contents {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    item.Key,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object %s, %v", *item.Key, err)
		}
	}

	return nil
}
