# Здесь первый контейнер с библиотеками голанг будет только строить наш сервер, имя у него будет build
FROM golang:1.16-alpine AS build

# определяем рабочую директорию, теперь все команды будут запускаться отсюда
WORKDIR /app

# копируем файлы зависимостей
COPY go.mod ./
COPY go.sum ./
# качаем зависимости
RUN go mod download

# копируем исходники
COPY *.go ./

# собираем сервер
RUN go build -o /server

# а здесь мы только копируем уже собранный сервер и запускаем его
FROM gcr.io/distroless/base-debian10

WORKDIR /

# команда from определит имя контейнера откуда мы копируем собранный сервер, если бы мы не обьявили AS build, то было бы --from=0
COPY --from=build /server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]