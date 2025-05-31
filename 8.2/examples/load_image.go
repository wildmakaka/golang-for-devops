package examples

import (
	"bufio"
	"context"
	"github.com/docker/docker/client"
	"io"
	"os"
)

//загрузка имейджа из файла
func LoadImage(filepath string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	f, err := os.Open(filepath) //открываем файл откуда будем читать наш(и) имейдж(и)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)                   //создаем ридер
	load, err := cli.ImageLoad(ctx, reader, false) //этого вполне достаточно для sdk
	if err != nil {
		return
	}

	defer load.Body.Close() //не забываем закрыть ответ от sdk

	io.Copy(os.Stdout, load.Body) //показываем в stdout
}
