package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func loadCredentials(file string) (string, string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	r := csv.NewReader(f)
	_, err = r.Read() // Ignora la línea de encabezado
	if err != nil {
		return "", "", err
	}

	record, err := r.Read() // Lee la segunda línea
	if err != nil {
		return "", "", err
	}

	return record[0], record[1], nil
}

func upload(sess *session.Session, file io.Reader, bucketName string, keyName string) (string, error) {
	// Crea un nuevo uploader con la sesión y la configuración por defecto
	uploader := s3manager.NewUploader(sess)

	// Sube el archivo a S3
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
		Body:   file,
		ACL:    aws.String("public-read"), // Hace que el objeto sea público
	})
	log.Println("File uploaded to", bucketName, keyName)
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}

	// Genera una URL pre-firmada para el objeto
	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(15 * time.Minute) // URL válida por 15 minutos
	if err != nil {
		return "", fmt.Errorf("failed to sign request, %v", err)
	}

	return urlStr, nil
}

func uploadHandler(sess *session.Session, bucketName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parsea el formulario multiparte
		err := r.ParseMultipartForm(10 << 20) // Máximo de 10 MB
		if err != nil {
			http.Error(w, "Error al parsear el formulario", http.StatusBadRequest)
			return
		}

		// Obtiene el archivo del formulario
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error al obtener el archivo", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Obtiene la ID del usuario y el nombre de la carpeta del formulario
		userID := r.FormValue("userID")
		folderName := r.FormValue("folderName")

		// Construye la clave con la ID del usuario y el nombre de la carpeta
		key := fmt.Sprintf("%s/%s/%s", userID, folderName, header.Filename)

		// Sube el archivo a S3 con la clave
		location, err := upload(sess, file, bucketName, key)
		if err != nil {
			http.Error(w, "Error al subir el archivo", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Archivo subido con éxito. URL: %s", location)
	})
}

func main() {
	accessKey, secretKey, err := loadCredentials("rootkey.csv") // Reemplaza con la ruta de tu archivo
	if err != nil {
		fmt.Println("Error al cargar las credenciales:", err)
		return
	}

	// Crea una nueva sesión con las credenciales de AWS
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // Cambiado a tu región
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})

	if err != nil {
		fmt.Println("Error al crear la sesión:", err)
		return
	}

	http.HandleFunc("/upload", uploadHandler(sess, "sisdis"))
	http.HandleFunc("/delete", deleteHandler(sess, "sisdis"))
	http.HandleFunc("/deleteFolder", deleteFolderHandler(sess, "sisdis"))

	http.ListenAndServe(":8080", nil)
}

func delete(sess *session.Session, bucketName string, keyName string) error {
	// Crea un nuevo servicio S3
	svc := s3.New(sess)

	// Elimina el archivo de S3
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file, %v", err)
	}

	return nil
}

func deleteHandler(sess *session.Session, bucketName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtiene la clave del archivo del formulario
		key := r.FormValue("key")

		// Elimina el archivo de S3
		err := delete(sess, bucketName, key)
		if err != nil {
			http.Error(w, "Error al eliminar el archivo", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Archivo eliminado con éxito")
	})
}

func deleteFolder(sess *session.Session, bucketName string, folderName string) error {
	// Crea un nuevo servicio S3
	svc := s3.New(sess)

	// Lista todos los objetos en la carpeta
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(folderName),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects in folder, %v", err)
	}

	// Elimina cada objeto en la carpeta
	for _, item := range resp.Contents {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    item.Key,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object %s, %v", *item.Key, err)
		}
	}

	return nil
}

func deleteFolderHandler(sess *session.Session, bucketName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtiene el nombre de la carpeta del formulario
		folderName := r.FormValue("folderName")

		// Elimina la carpeta en S3
		err := deleteFolder(sess, bucketName, folderName)
		if err != nil {
			http.Error(w, "Error al eliminar la carpeta", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Carpeta eliminada con éxito")
	})
}
