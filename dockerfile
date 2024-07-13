# Этап сборки
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы и собираем приложение
COPY . ./
RUN go build -o app ./cmd/main.go

# Этап выполнения
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app .

EXPOSE 8080

# Команда для запуска приложения
CMD ["./app"]
