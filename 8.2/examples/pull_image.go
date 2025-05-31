package examples

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func PullImage() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()) //собираем клиент с опциями из env переменных
	if err != nil {
		panic(err)
	}

	out, err := cli.ImagePull(ctx, "alpine", types.ImagePullOptions{}) //качаем изображение по названию
	if err != nil {
		panic(err)
	}

	defer out.Close() //ответ API

	io.Copy(os.Stdout, out) //выводим ответ в stdout
}
