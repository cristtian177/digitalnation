# Usar la imagen oficial de Go versión 1.18 como base
FROM golang:1.18

# Establecer /app como el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar go.mod y go.sum al directorio de trabajo
COPY go.mod go.sum ./

# Descargar todas las dependencias especificadas en go.mod y go.sum
RUN go mod download

# Copiar el resto de los archivos al directorio de trabajo
COPY . .

# Compilar la aplicación Go para producir un ejecutable
RUN go build -o main src/main.go

# Ejecutar el ejecutable main por defecto cuando se inicie el contenedor
CMD ["./main"]