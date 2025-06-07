# Контейнер для билда и запуска голанг
FROM golang:1.16-alpine

WORKDIR /app

# Качаем модули
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Копируем файлы
COPY *.go ./

# Собираем образ
RUN go build -o /server

# Открывавем порт для запросов
EXPOSE 8080

# Запускаем наш сервер
ENTRYPOINT ["/server"]
