package aws

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Client es una estructura que contiene una sesión de AWS y el nombre de un bucket de S3
type S3Client struct {
	Session *session.Session
	Bucket  string
}

// NewS3Client crea un nuevo cliente S3 con una sesión y un bucket
func NewS3Client(session *session.Session, bucket string) *S3Client {
	return &S3Client{Session: session, Bucket: bucket}
}

// Upload sube un archivo a S3 y devuelve la URL del objeto
func (c *S3Client) Upload(file io.Reader, keyName string) (string, error) {
	log.Println("Subiendo archivo a S3...")

	// Crea un nuevo uploader con la sesión y la configuración por defecto
	uploader := s3manager.NewUploader(c.Session)

	// Sube el archivo a S3
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(keyName),
		Body:   file,
		ACL:    aws.String("public-read"), // Hace que el objeto sea público
	})
	log.Println("Archivo subido a", c.Bucket, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	// Genera una URL pre-firmada para el objeto
	svc := s3.New(c.Session)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(keyName),
	})

	urlStr, err := req.Presign(15 * time.Hour) // URL válida por 15 horas
	if err != nil {
		return "", fmt.Errorf("failed to sign request, %v", err)
	}

	return urlStr, nil
}

// Delete elimina un archivo de S3
func (c *S3Client) Delete(keyName string) error {
	log.Println("Eliminando archivo de S3...")

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

	log.Println("Archivo eliminado de S3:", keyName)
	return nil
}

// DeleteFolder elimina una carpeta de S3
func (c *S3Client) DeleteFolder(folderName string) error {
	log.Println("Eliminando carpeta de S3...")

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

		log.Println("Objeto eliminado de S3:", *item.Key)
	}

	log.Println("Carpeta eliminada de S3:", folderName)
	return nil
}
