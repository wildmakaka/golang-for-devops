package examples

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io"
	"os"
)

//сохраним изображение на диск для передачи в репозиторий или на другую машину напрямую
func SaveImage(imageid string, filepath string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	images := make([]string, 1) //сохранение образов работает со слайсом изображений, поэтому создадим его
	images[0] = imageid
	reader, err := cli.ImageSave(ctx, images) //сохраняем изображения, в данном случае поток идет в reader
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	file, err := os.Create(filepath) //создаем файл для записи
	defer file.Close()
	writtenBytes, err := io.Copy(file, reader) //копируем все содержимое reader в файл
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Bytes written %d\n", writtenBytes)
}
