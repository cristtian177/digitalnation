// main.go
package main

import (
	"encoding/csv"
	"fmt"
	"importlibss/src/aws"
	"importlibss/src/server"
	"net/http"
	"os"

	sdkaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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
func main() {
	accessKey, secretKey, err := loadCredentials("./rootkey.csv") // Reemplaza con la ruta de tu archivo
	if err != nil {
		fmt.Println("Error al cargar las credenciales:", err)
		return
	}

	// Crea una nueva sesión con las credenciales de AWS
	sess, err := session.NewSession(&sdkaws.Config{
		Region:      sdkaws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})

	if err != nil {
		fmt.Println("Error al crear la sesión:", err)
		return
	}

	s3Client := aws.NewS3Client(sess, "sisdis")
	s := server.NewServer(s3Client)

	http.HandleFunc("/upload", s.UploadHandler())
	http.HandleFunc("/delete", s.DeleteHandler())
	http.HandleFunc("/deleteFolder", s.DeleteFolderHandler())

	http.ListenAndServe(":8080", nil)
}
