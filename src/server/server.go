package server

import (
	"fmt"
	"importlibss/src/aws"
	"net/http"
)

type Server struct {
	S3Client *aws.S3Client
}

func NewServer(s3Client *aws.S3Client) *Server {
	return &Server{S3Client: s3Client}
}

func (s *Server) UploadHandler() http.HandlerFunc {
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
		location, err := s.S3Client.Upload(file, key)
		if err != nil {
			http.Error(w, "Error al subir el archivo", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Archivo subido con éxito. URL: %s", location)
	})
}

func (s *Server) DeleteHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtiene la clave del archivo del formulario
		key := r.FormValue("key")

		// Elimina el archivo de S3
		err := s.S3Client.Delete(key)
		if err != nil {
			http.Error(w, "Error al eliminar el archivo", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Archivo eliminado con éxito")
	})
}

func (s *Server) DeleteFolderHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtiene el nombre de la carpeta del formulario
		folderName := r.FormValue("folderName")

		// Elimina la carpeta en S3
		err := s.S3Client.DeleteFolder(folderName)
		if err != nil {
			http.Error(w, "Error al eliminar la carpeta", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Carpeta eliminada con éxito")
	})
}
