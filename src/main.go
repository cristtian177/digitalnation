package main

import (
	"encoding/csv"
	"importlibss/src/aws"
	"importlibss/src/server"
	"log"
	"net/http"
	"os"

	sdkaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// loadCredentials carga las credenciales de AWS desde un archivo CSV
func loadCredentials(file string) (string, string, error) {
	f, err := os.Open(file) // Abre el archivo
	if err != nil {
		return "", "", err
	}
	defer f.Close() // Asegura que el archivo se cierre al final

	r := csv.NewReader(f) // Crea un nuevo lector CSV
	_, err = r.Read()     // Ignora la línea de encabezado
	if err != nil {
		return "", "", err
	}

	record, err := r.Read() // Lee la segunda línea
	if err != nil {
		return "", "", err
	}

	return record[0], record[1], nil // Devuelve las credenciales
}

func main() {
	log.Println("Iniciando la aplicación...")

	accessKey, secretKey, err := loadCredentials("./rootkey.csv") // Carga las credenciales de AWS
	if err != nil {
		log.Println("Error al cargar las credenciales:", err)
		return
	}

	// Crea una nueva sesión con las credenciales de AWS
	sess, err := session.NewSession(&sdkaws.Config{
		Region:      sdkaws.String("us-east-1"),                                 // Configura la región
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""), // Configura las credenciales
	})

	if err != nil {
		log.Println("Error al crear la sesión:", err)
		return
	}

	s3Client := aws.NewS3Client(sess, "sisdis") // Crea un nuevo cliente S3
	s := server.NewServer(s3Client)             // Crea un nuevo servidor

	// Configura los manejadores de rutas
	http.HandleFunc("/upload", s.UploadHandler())
	http.HandleFunc("/delete", s.DeleteHandler())
	http.HandleFunc("/deleteFolder", s.DeleteFolderHandler())

	log.Println("Servidor iniciado en el puerto 8080")
	http.ListenAndServe(":8080", nil) // Inicia el servidor en el puerto 8080
}
