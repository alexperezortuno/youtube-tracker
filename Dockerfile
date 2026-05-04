# BUILD STAGE
FROM golang:1.26-alpine AS builder

WORKDIR /app

# dependencias básicas
RUN apk add --no-cache git

# copiar mod primero (cache eficiente)
COPY go.mod go.sum ./
RUN go mod download

# copiar código
COPY . .

# build binario estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o app ./cmd/main.go


# RUNTIME STAGE
FROM alpine:3.20

WORKDIR /app

# certificados para HTTPS
RUN apk add --no-cache ca-certificates

# copiar binario
COPY --from=builder /app/app .

# opcional: copiar archivo de canales
COPY channels.txt ./channels.txt

# puerto opcional (si agregas API luego)
EXPOSE 8080

# variables por defecto
ENV GOMAXPROCS=2

# comando
ENTRYPOINT ["./app"]
